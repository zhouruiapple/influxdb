package kv

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/influxdata/influxdb/notification/endpoint"

	"github.com/influxdata/influxdb"
)

var (
	notificationEndpointBucket = []byte("notificationendpointv1")

	// ErrNotificationEndpointNotFound is used when the notification endpoint is not found.
	ErrNotificationEndpointNotFound = &influxdb.Error{
		Msg:  "notification endpoint not found",
		Code: influxdb.ENotFound,
	}

	// ErrInvalidNotificationEndpointID is used when the service was provided
	// an invalid ID format.
	ErrInvalidNotificationEndpointID = &influxdb.Error{
		Code: influxdb.EInvalid,
		Msg:  "provided notification endpoint ID has invalid format",
	}
)

var _ influxdb.NotificationEndpointService = (*Service)(nil)

func (s *Service) initializeNotificationEndpoint(ctx context.Context, tx Tx) error {
	if _, err := s.notificationEndpointBucket(tx); err != nil {
		return err
	}
	return nil
}

// UnavailableNotificationEndpointServiceError is used if we aren't able to interact with the
// store, it means the store is not available at the moment (e.g. network).
func UnavailableNotificationEndpointServiceError(err error) *influxdb.Error {
	return &influxdb.Error{
		Code: influxdb.EInternal,
		Msg:  fmt.Sprintf("Unable to connect to notification endpoint store service. Please try again; Err: %v", err),
		Op:   "kv/notificationEndpoint",
	}
}

// InternalNotificationEndpointServiceError is used when the error comes from an
// internal system.
func InternalNotificationEndpointServiceError(err error) *influxdb.Error {
	return &influxdb.Error{
		Code: influxdb.EInternal,
		Msg:  fmt.Sprintf("Unknown internal notification endpoint data error; Err: %v", err),
		Op:   "kv/notificationEndpoint",
	}
}

func (s *Service) notificationEndpointBucket(tx Tx) (Bucket, error) {
	b, err := tx.Bucket(notificationEndpointBucket)
	if err != nil {
		return nil, UnavailableNotificationEndpointServiceError(err)
	}
	return b, nil
}

// CreateNotificationEndpoint creates a new notification endpoint and sets b.ID with the new identifier.
func (s *Service) CreateNotificationEndpoint(ctx context.Context, nr influxdb.NotificationEndpoint, userID influxdb.ID) error {
	return s.kv.Update(ctx, func(tx Tx) error {
		return s.createNotificationEndpoint(ctx, tx, nr, userID)
	})
}

func (s *Service) createNotificationEndpoint(ctx context.Context, tx Tx, nr influxdb.NotificationEndpoint, userID influxdb.ID) error {
	id := s.IDGenerator.ID()
	nr.SetID(id)
	now := s.TimeGenerator.Now()
	nr.SetCreatedAt(now)
	nr.SetUpdatedAt(now)
	if err := s.putNotificationEndpoint(ctx, tx, nr); err != nil {
		return err
	}

	urm := &influxdb.UserResourceMapping{
		ResourceID:   id,
		UserID:       userID,
		UserType:     influxdb.Owner,
		ResourceType: influxdb.NotificationEndpointResourceType,
	}
	return s.createUserResourceMapping(ctx, tx, urm)
}

// PatchNotificationEndpoint updates a single  notification endpoint with changeset.
// Returns the new notification endpoint state after update.
func (s *Service) PatchNotificationEndpoint(ctx context.Context, id influxdb.ID, upd influxdb.NotificationEndpointUpdate) (influxdb.NotificationEndpoint, error) {
	var nr influxdb.NotificationEndpoint
	if err := s.kv.Update(ctx, func(tx Tx) (err error) {
		nr, err = s.patchNotificationEndpoint(ctx, tx, id, upd)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return nr, nil
}

func (s *Service) patchNotificationEndpoint(ctx context.Context, tx Tx, id influxdb.ID, upd influxdb.NotificationEndpointUpdate) (influxdb.NotificationEndpoint, error) {
	nr, err := s.findNotificationEndpointByID(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if upd.Name != nil {
		nr.SetName(*upd.Name)
	}
	if upd.Description != nil {
		nr.SetDescription(*upd.Description)
	}
	if upd.Status != nil {
		nr.SetStatus(*upd.Status)
	}
	nr.SetUpdatedAt(s.TimeGenerator.Now())
	err = s.putNotificationEndpoint(ctx, tx, nr)
	if err != nil {
		return nil, err
	}

	return nr, nil
}

// PutNotificationEndpoint put a notification endpoint to storage.
func (s *Service) PutNotificationEndpoint(ctx context.Context, nr influxdb.NotificationEndpoint) error {
	return s.kv.Update(ctx, func(tx Tx) (err error) {
		return s.putNotificationEndpoint(ctx, tx, nr)
	})
}

func (s *Service) putNotificationEndpoint(ctx context.Context, tx Tx, nr influxdb.NotificationEndpoint) error {
	if err := nr.Valid(); err != nil {
		return err
	}
	encodedID, _ := nr.GetID().Encode()

	v, err := json.Marshal(nr)
	if err != nil {
		return err
	}

	bucket, err := s.notificationEndpointBucket(tx)
	if err != nil {
		return err
	}

	if err := bucket.Put(encodedID, v); err != nil {
		return UnavailableNotificationEndpointServiceError(err)
	}
	return nil
}

// FindNotificationEndpointByID returns a single notification endpoint by ID.
func (s *Service) FindNotificationEndpointByID(ctx context.Context, id influxdb.ID) (influxdb.NotificationEndpoint, error) {
	var (
		nr  influxdb.NotificationEndpoint
		err error
	)

	err = s.kv.View(ctx, func(tx Tx) error {
		nr, err = s.findNotificationEndpointByID(ctx, tx, id)
		return err
	})

	return nr, err
}

func (s *Service) findNotificationEndpointByID(ctx context.Context, tx Tx, id influxdb.ID) (influxdb.NotificationEndpoint, error) {
	encID, err := id.Encode()
	if err != nil {
		return nil, ErrInvalidNotificationEndpointID
	}

	bucket, err := s.notificationEndpointBucket(tx)
	if err != nil {
		return nil, err
	}

	v, err := bucket.Get(encID)
	if IsNotFound(err) {
		return nil, ErrNotificationEndpointNotFound
	}
	if err != nil {
		return nil, InternalNotificationEndpointServiceError(err)
	}

	return endpoint.UnmarshalJSON(v)
}

// FindNotificationEndpoints returns a list of notification endpoints that match filter and the total count of matching notification endpoints.
// Additional options provide pagination & sorting.
func (s *Service) FindNotificationEndpoints(ctx context.Context, filter influxdb.NotificationEndpointFilter, opt ...influxdb.FindOptions) (nrs []influxdb.NotificationEndpoint, n int, err error) {
	err = s.kv.View(ctx, func(tx Tx) error {
		nrs, n, err = s.findNotificationEndpoints(ctx, tx, filter)
		return err
	})
	return nrs, n, err
}

func (s *Service) findNotificationEndpoints(ctx context.Context, tx Tx, filter influxdb.NotificationEndpointFilter, opt ...influxdb.FindOptions) ([]influxdb.NotificationEndpoint, int, error) {
	nrs := make([]influxdb.NotificationEndpoint, 0)

	// m, err := s.findUserResourceMappings(ctx, tx, filter.UserResourceMappingFilter)
	// if err != nil {
	// 	return nil, 0, err
	// }
	//
	// if len(m) == 0 {
	// 	return nrs, 0, nil
	// }
	//
	// idMap := make(map[influxdb.ID]bool)
	// for _, item := range m {
	// 	idMap[item.ResourceID] = false
	// }

	if filter.OrgID != nil || filter.Organization != nil {
		o, err := s.FindOrganization(ctx, influxdb.OrganizationFilter{
			ID:   filter.OrgID,
			Name: filter.Organization,
		})

		if err != nil {
			return nrs, 0, err
		}
		filter.OrgID = &o.ID
	}

	var offset, limit, count int
	var descending bool
	if len(opt) > 0 {
		offset = opt[0].Offset
		limit = opt[0].Limit
		descending = opt[0].Descending
	}
	filterFn := filterNotificationEndpointsFn(filter)
	err := s.forEachNotificationEndpoint(ctx, tx, descending, func(nr *influxdb.NotificationEndpoint) bool {
		if filterFn(nr) {
			if count >= offset {
				nrs = append(nrs, *nr)
			}
			count++
		}

		if limit > 0 && len(nrs) >= limit {
			return false
		}

		return true
	})

	return nrs, len(nrs), err
}

// forEachNotificationEndpoint will iterate through all notification endpoints while fn returns true.
func (s *Service) forEachNotificationEndpoint(ctx context.Context, tx Tx, descending bool, fn func(*influxdb.NotificationEndpoint) bool) error {

	bkt, err := s.notificationEndpointBucket(tx)
	if err != nil {
		return err
	}

	cur, err := bkt.Cursor()
	if err != nil {
		return err
	}

	var k, v []byte
	if descending {
		k, v = cur.Last()
	} else {
		k, v = cur.First()
	}

	for k != nil {
		nr, err := endpoint.UnmarshalJSON(v)
		if err != nil {
			return err
		}
		if !fn(&nr) {
			break
		}

		if descending {
			k, v = cur.Prev()
		} else {
			k, v = cur.Next()
		}
	}

	return nil
}

func filterNotificationEndpointsFn(filter influxdb.NotificationEndpointFilter) func(nr *influxdb.NotificationEndpoint) bool {
	// if filter.OrgID != nil {
	// 	return func(nr influxdb.NotificationEndpoint) bool {
	// 		return nr.Get
	// 		_, ok := idMap[nr.GetID()]
	// 		return nr.GetOrgID() == *filter.OrgID && ok
	// 	}
	// }

	return func(nr *influxdb.NotificationEndpoint) bool {
		return true
	}
}

// DeleteNotificationEndpoint removes a notification endpoint by ID.
func (s *Service) DeleteNotificationEndpoint(ctx context.Context, id influxdb.ID) error {
	return s.kv.Update(ctx, func(tx Tx) error {
		return s.deleteNotificationEndpoint(ctx, tx, id)
	})
}

func (s *Service) deleteNotificationEndpoint(ctx context.Context, tx Tx, id influxdb.ID) error {
	encodedID, err := id.Encode()
	if err != nil {
		return ErrInvalidNotificationEndpointID
	}

	bucket, err := s.notificationEndpointBucket(tx)
	if err != nil {
		return err
	}

	_, err = bucket.Get(encodedID)
	if IsNotFound(err) {
		return ErrNotificationEndpointNotFound
	}
	if err != nil {
		return InternalNotificationEndpointServiceError(err)
	}

	if err := bucket.Delete(encodedID); err != nil {
		return InternalNotificationEndpointServiceError(err)
	}

	return s.deleteUserResourceMappings(ctx, tx, influxdb.UserResourceMappingFilter{
		ResourceID:   id,
		ResourceType: influxdb.NotificationEndpointResourceType,
	})
}
