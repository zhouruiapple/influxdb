package testing

import (
	"bytes"
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/mock"
	"github.com/influxdata/influxdb/notification/endpoint"
)

const (
	notificationEndpointOneID   = "020f755c3c082000"
	notificationEndpointTwoID   = "020f755c3c082001"
	notificationEndpointThreeID = "020f755c3c082002"
)

var notificationEndpointCmpOptions = cmp.Options{
	cmp.Comparer(func(x, y []byte) bool {
		return bytes.Equal(x, y)
	}),
	cmp.Transformer("Sort", func(in []influxdb.NotificationEndpoint) []influxdb.NotificationEndpoint {
		out := append([]influxdb.NotificationEndpoint(nil), in...) // Copy input to avoid mutating it
		sort.Slice(out, func(i, j int) bool {
			return out[i].GetID() > out[j].GetID()
		})
		return out
	}),
}

// NotificationEndpointFields will include the IDGenerator, and notificationEndpoints
type NotificationEndpointFields struct {
	IDGenerator           influxdb.IDGenerator
	TimeGenerator         influxdb.TimeGenerator
	NotificationEndpoints []influxdb.NotificationEndpoint
	Organizations         []*influxdb.Organization
}

type notificationEndpointServiceF func(
	init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointService, string, func()),
	t *testing.T,
)

// NotificationEndpointService tests all the service functions.
func NotificationEndpointService(
	init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointService, string, func()),
	t *testing.T,
) {
	tests := []struct {
		name string
		fn   notificationEndpointServiceF
	}{
		{
			name: "CreateNotificationEndpoint",
			fn:   CreateNotificationEndpoint,
		},
		{
			name: "FindNotificationEndpointByID",
			fn:   FindNotificationEndpointByID,
		},
		{
			name: "FindNotificationEndpoints",
			fn:   FindNotificationEndpoints,
		},
		// {
		// 	name: "FindNotificationEndpoint",
		// 	fn:   FindNotificationEndpoint,
		// },
		// {
		// 	name: "UpdateNotificationEndpoint",
		// 	fn:   UpdateNotificationEndpoint,
		// },
		// {
		// 	name: "DeleteNotificationEndpoint",
		// 	fn:   DeleteNotificationEndpoint,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(init, t)
		})
	}
}

// CreateNotificationEndpoint testing
func CreateNotificationEndpoint(
	init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointService, string, func()),
	t *testing.T,
) {
	type args struct {
		notificationEndpoint influxdb.NotificationEndpoint
		userID               influxdb.ID
	}
	type wants struct {
		err                   error
		notificationEndpoints []influxdb.NotificationEndpoint
	}

	tests := []struct {
		name   string
		fields NotificationEndpointFields
		args   args
		wants  wants
	}{
		{
			name: "basic create notificationEndpoint",
			fields: NotificationEndpointFields{
				IDGenerator: &mock.IDGenerator{
					IDFn: func() influxdb.ID {
						return MustIDBase16(notificationEndpointThreeID)
					},
				},
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:          MustIDBase16(notificationEndpointOneID),
							Name:        "name1",
							Description: "description1",
							OrgID:       MustIDBase16(orgOneID),
							Status:      influxdb.Active,
						},
						URL:   "slackurl",
						Token: influxdb.SecretField("token"),
					},
					&endpoint.SMTP{
						Base: endpoint.Base{
							ID:          MustIDBase16(notificationEndpointTwoID),
							Name:        "name2",
							Description: "description2",
							OrgID:       MustIDBase16(orgTwoID),
							Status:      influxdb.Active,
						},
						Host: "http://www.host.com",
						Port: 8080,
					},
				},
				Organizations: []*influxdb.Organization{
					{
						Name: "theorg",
						ID:   MustIDBase16(orgOneID),
					},
					{
						Name: "otherorg",
						ID:   MustIDBase16(orgTwoID),
					},
				},
			},
			args: args{
				userID: MustIDBase16(userOneID),
				notificationEndpoint: &endpoint.PagerDuty{
					Base: endpoint.Base{
						Name:        "name3",
						Description: "description3",
						OrgID:       MustIDBase16(orgTwoID),
						Status:      influxdb.Active,
					},
					URL:        "pgurl",
					RoutingKey: influxdb.SecretField("key"),
				},
			},
			wants: wants{
				notificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:          MustIDBase16(notificationEndpointOneID),
							Name:        "name1",
							Description: "description1",
							OrgID:       MustIDBase16(orgOneID),
							Status:      influxdb.Active,
						},
						URL:   "slackurl",
						Token: influxdb.SecretField("token"),
					},
					&endpoint.SMTP{
						Base: endpoint.Base{
							ID:          MustIDBase16(notificationEndpointTwoID),
							Name:        "name2",
							Description: "description2",
							OrgID:       MustIDBase16(orgTwoID),
							Status:      influxdb.Active,
						},
						Host: "http://www.host.com",
						Port: 8080,
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:          MustIDBase16(notificationEndpointThreeID),
							Name:        "name3",
							Description: "description3",
							OrgID:       MustIDBase16(orgTwoID),
							Status:      influxdb.Active,
							CRUDLog: influxdb.CRUDLog{
								UpdatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
								CreatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
							},
						},
						URL:        "pgurl",
						RoutingKey: influxdb.SecretField("key"),
					},
				},
			},
		},
		{
			name: "org does not exist",
			fields: NotificationEndpointFields{
				IDGenerator:   mock.NewIDGenerator(notificationEndpointOneID, t),
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:          MustIDBase16(notificationEndpointOneID),
							Name:        "name1",
							Description: "description1",
							OrgID:       MustIDBase16(orgOneID),
							Status:      influxdb.Active,
						},
						URL:   "slackurl",
						Token: influxdb.SecretField("token"),
					},
				},
				Organizations: []*influxdb.Organization{
					{
						Name: "theorg",
						ID:   MustIDBase16(orgOneID),
					},
				},
			},
			args: args{
				userID: MustIDBase16(userOneID),
				notificationEndpoint: &endpoint.PagerDuty{
					Base: endpoint.Base{
						Name:        "name3",
						Description: "description3",
						OrgID:       MustIDBase16(orgTwoID),
						Status:      influxdb.Active,
					},
					URL:        "pgurl",
					RoutingKey: influxdb.SecretField("key"),
				},
			},
			wants: wants{
				notificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:          MustIDBase16(notificationEndpointOneID),
							Name:        "name1",
							Description: "description1",
							OrgID:       MustIDBase16(orgOneID),
							Status:      influxdb.Active,
						},
						URL:   "slackurl",
						Token: influxdb.SecretField("token"),
					},
				},
				err: &influxdb.Error{
					Code: influxdb.ENotFound,
					Msg:  "organization not found",
					Op:   influxdb.OpCreateNotificationEndpoint,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, opPrefix, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()
			err := s.CreateNotificationEndpoint(ctx, tt.args.notificationEndpoint, tt.args.userID)
			diffPlatformErrors(tt.name, err, tt.wants.err, opPrefix, t)

			defer s.DeleteNotificationEndpoint(ctx, tt.args.notificationEndpoint.GetID())

			notificationEndpoints, _, err := s.FindNotificationEndpoints(ctx, influxdb.NotificationEndpointFilter{})
			if err != nil {
				t.Fatalf("failed to retrieve notificationEndpoints: %v", err)
			}
			if diff := cmp.Diff(notificationEndpoints, tt.wants.notificationEndpoints, notificationEndpointCmpOptions...); diff != "" {
				t.Errorf("notificationEndpoints are different -got/+want\ndiff %s", diff)
			}
		})
	}
}

// FindNotificationEndpointByID testing
func FindNotificationEndpointByID(
	init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointService, string, func()),
	t *testing.T,
) {
	type args struct {
		id influxdb.ID
	}
	type wants struct {
		err                  error
		notificationEndpoint influxdb.NotificationEndpoint
	}

	tests := []struct {
		name   string
		fields NotificationEndpointFields
		args   args
		wants  wants
	}{
		{
			name: "basic find notificationEndpoint by id",
			fields: NotificationEndpointFields{
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:          MustIDBase16(notificationEndpointOneID),
							OrgID:       MustIDBase16(orgOneID),
							Name:        "notificationEndpoint1",
							Description: "description1",
							Status:      influxdb.Active,
						},
						URL:   "slackurl",
						Token: influxdb.SecretField("token"),
					},
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:          MustIDBase16(notificationEndpointTwoID),
							OrgID:       MustIDBase16(orgOneID),
							Name:        "notificationEndpoint2",
							Description: "description2",
							Status:      influxdb.Active,
						},
						URL:   "slackurl",
						Token: influxdb.SecretField("token"),
					},
				},
				Organizations: []*influxdb.Organization{
					{
						Name: "theorg",
						ID:   MustIDBase16(orgOneID),
					},
				},
			},
			args: args{
				id: MustIDBase16(notificationEndpointTwoID),
			},
			wants: wants{
				notificationEndpoint: &endpoint.Slack{
					Base: endpoint.Base{
						ID:          MustIDBase16(notificationEndpointTwoID),
						OrgID:       MustIDBase16(orgOneID),
						Name:        "notificationEndpoint2",
						Description: "description2",
						Status:      influxdb.Active,
					},
					URL:   "slackurl",
					Token: influxdb.SecretField("token"),
				},
			},
		},
		{
			name: "find notificationEndpoint by id not exist",
			fields: NotificationEndpointFields{
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:          MustIDBase16(notificationEndpointOneID),
							OrgID:       MustIDBase16(orgOneID),
							Name:        "notificationEndpoint1",
							Description: "description1",
							Status:      influxdb.Active,
						},
						URL:   "slackurl",
						Token: influxdb.SecretField("token"),
					},
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:          MustIDBase16(notificationEndpointTwoID),
							OrgID:       MustIDBase16(orgOneID),
							Name:        "notificationEndpoint2",
							Description: "description2",
							Status:      influxdb.Active,
						},
						URL:   "slackurl",
						Token: influxdb.SecretField("token"),
					},
				},
				Organizations: []*influxdb.Organization{
					{
						Name: "theorg",
						ID:   MustIDBase16(orgOneID),
					},
				},
			},
			args: args{
				id: MustIDBase16(notificationEndpointThreeID),
			},
			wants: wants{
				err: &influxdb.Error{
					Code: influxdb.ENotFound,
					Op:   influxdb.OpFindNotificationEndpointByID,
					Msg:  "notification endpoint not found",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, opPrefix, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()

			notificationEndpoint, err := s.FindNotificationEndpointByID(ctx, tt.args.id)
			diffPlatformErrors(tt.name, err, tt.wants.err, opPrefix, t)

			if diff := cmp.Diff(notificationEndpoint, tt.wants.notificationEndpoint, notificationEndpointCmpOptions...); diff != "" {
				t.Errorf("notificationEndpoint is different -got/+want\ndiff %s", diff)
			}
		})
	}
}

// FindNotificationEndpoints testing
func FindNotificationEndpoints(
	init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointService, string, func()),
	t *testing.T,
) {
	type args struct {
		ID             influxdb.ID
		name           string
		organization   string
		organizationID influxdb.ID
		findOptions    influxdb.FindOptions
	}

	type wants struct {
		notificationEndpoints []influxdb.NotificationEndpoint
		err                   error
	}
	tests := []struct {
		name   string
		fields NotificationEndpointFields
		args   args
		wants  wants
	}{
		{
			name: "find all notificationEndpoints",
			fields: NotificationEndpointFields{
				Organizations: []*influxdb.Organization{
					{
						Name: "theorg",
						ID:   MustIDBase16(orgOneID),
					},
					{
						Name: "otherorg",
						ID:   MustIDBase16(orgTwoID),
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:          MustIDBase16(notificationEndpointOneID),
							OrgID:       MustIDBase16(orgOneID),
							Name:        "notificationEndpoint1",
							Description: "description1",
							Status:      influxdb.Active,
						},
						URL:   "slackurl",
						Token: influxdb.SecretField("token"),
					},
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:          MustIDBase16(notificationEndpointTwoID),
							OrgID:       MustIDBase16(orgTwoID),
							Name:        "notificationEndpoint2",
							Description: "description2",
							Status:      influxdb.Active,
						},
						URL:   "slackurl",
						Token: influxdb.SecretField("token"),
					},
				},
			},
			args: args{},
			wants: wants{
				notificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:          MustIDBase16(notificationEndpointOneID),
							OrgID:       MustIDBase16(orgOneID),
							Name:        "notificationEndpoint1",
							Description: "description1",
							Status:      influxdb.Active,
						},
						URL:   "slackurl",
						Token: influxdb.SecretField("token"),
					},
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:          MustIDBase16(notificationEndpointTwoID),
							OrgID:       MustIDBase16(orgTwoID),
							Name:        "notificationEndpoint2",
							Description: "description2",
							Status:      influxdb.Active,
						},
						URL:   "slackurl",
						Token: influxdb.SecretField("token"),
					},
				},
			},
		},
		// {
		// 	name: "find all notificationEndpoints by offset and limit",
		// 	fields: NotificationEndpointFields{
		// 		Organizations: []*influxdb.Organization{
		// 			{
		// 				Name: "theorg",
		// 				ID:   MustIDBase16(orgOneID),
		// 			},
		// 		},
		// 		NotificationEndpoints: []*influxdb.NotificationEndpoint{
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointOneID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "abc",
		// 			},
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointTwoID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "def",
		// 			},
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointThreeID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "xyz",
		// 			},
		// 		},
		// 	},
		// 	args: args{
		// 		findOptions: influxdb.FindOptions{
		// 			Offset: 1,
		// 			Limit:  1,
		// 		},
		// 	},
		// 	wants: wants{
		// 		notificationEndpoints: []*influxdb.NotificationEndpoint{
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointTwoID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "def",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	name: "find all notificationEndpoints by descending",
		// 	fields: NotificationEndpointFields{
		// 		Organizations: []*influxdb.Organization{
		// 			{
		// 				Name: "theorg",
		// 				ID:   MustIDBase16(orgOneID),
		// 			},
		// 		},
		// 		NotificationEndpoints: []*influxdb.NotificationEndpoint{
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointOneID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "abc",
		// 			},
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointTwoID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "def",
		// 			},
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointThreeID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "xyz",
		// 			},
		// 		},
		// 	},
		// 	args: args{
		// 		findOptions: influxdb.FindOptions{
		// 			Offset:     1,
		// 			Descending: true,
		// 		},
		// 	},
		// 	wants: wants{
		// 		notificationEndpoints: []*influxdb.NotificationEndpoint{
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointTwoID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "def",
		// 			},
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointOneID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "abc",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	name: "find notificationEndpoints by organization name",
		// 	fields: NotificationEndpointFields{
		// 		Organizations: []*influxdb.Organization{
		// 			{
		// 				Name: "theorg",
		// 				ID:   MustIDBase16(orgOneID),
		// 			},
		// 			{
		// 				Name: "otherorg",
		// 				ID:   MustIDBase16(orgTwoID),
		// 			},
		// 		},
		// 		NotificationEndpoints: []*influxdb.NotificationEndpoint{
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointOneID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "abc",
		// 			},
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointTwoID),
		// 				OrgID: MustIDBase16(orgTwoID),
		// 				Name:  "xyz",
		// 			},
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointThreeID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "123",
		// 			},
		// 		},
		// 	},
		// 	args: args{
		// 		organization: "theorg",
		// 	},
		// 	wants: wants{
		// 		notificationEndpoints: []*influxdb.NotificationEndpoint{
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointOneID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "abc",
		// 			},
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointThreeID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "123",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	name: "find notificationEndpoints by organization id",
		// 	fields: NotificationEndpointFields{
		// 		Organizations: []*influxdb.Organization{
		// 			{
		// 				Name: "theorg",
		// 				ID:   MustIDBase16(orgOneID),
		// 			},
		// 			{
		// 				Name: "otherorg",
		// 				ID:   MustIDBase16(orgTwoID),
		// 			},
		// 		},
		// 		NotificationEndpoints: []*influxdb.NotificationEndpoint{
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointOneID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "abc",
		// 			},
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointTwoID),
		// 				OrgID: MustIDBase16(orgTwoID),
		// 				Name:  "xyz",
		// 			},
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointThreeID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "123",
		// 			},
		// 		},
		// 	},
		// 	args: args{
		// 		organizationID: MustIDBase16(orgOneID),
		// 	},
		// 	wants: wants{
		// 		notificationEndpoints: []*influxdb.NotificationEndpoint{
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointOneID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "abc",
		// 			},
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointThreeID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "123",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	name: "find notificationEndpoint by name",
		// 	fields: NotificationEndpointFields{
		// 		Organizations: []*influxdb.Organization{
		// 			{
		// 				Name: "theorg",
		// 				ID:   MustIDBase16(orgOneID),
		// 			},
		// 		},
		// 		NotificationEndpoints: []*influxdb.NotificationEndpoint{
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointOneID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "abc",
		// 			},
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointTwoID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "xyz",
		// 			},
		// 		},
		// 	},
		// 	args: args{
		// 		name: "xyz",
		// 	},
		// 	wants: wants{
		// 		notificationEndpoints: []*influxdb.NotificationEndpoint{
		// 			{
		// 				ID:    MustIDBase16(notificationEndpointTwoID),
		// 				OrgID: MustIDBase16(orgOneID),
		// 				Name:  "xyz",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	name: "missing notificationEndpoint returns no notificationEndpoints",
		// 	fields: NotificationEndpointFields{
		// 		Organizations: []*influxdb.Organization{
		// 			{
		// 				Name: "theorg",
		// 				ID:   MustIDBase16(orgOneID),
		// 			},
		// 		},
		// 		NotificationEndpoints: []*influxdb.NotificationEndpoint{},
		// 	},
		// 	args: args{
		// 		name: "xyz",
		// 	},
		// 	wants: wants{},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, opPrefix, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()

			filter := influxdb.NotificationEndpointFilter{}
			if tt.args.ID.Valid() {
				filter.ID = &tt.args.ID
			}
			if tt.args.organizationID.Valid() {
				filter.OrgID = &tt.args.organizationID
			}
			if tt.args.organization != "" {
				filter.Organization = &tt.args.organization
			}

			notificationEndpoints, _, err := s.FindNotificationEndpoints(ctx, filter, tt.args.findOptions)
			diffPlatformErrors(tt.name, err, tt.wants.err, opPrefix, t)

			if diff := cmp.Diff(notificationEndpoints, tt.wants.notificationEndpoints, notificationEndpointCmpOptions...); diff != "" {
				t.Errorf("notificationEndpoints are different -got/+want\ndiff %s", diff)
			}
		})
	}
}

//
// // DeleteNotificationEndpoint testing
// func DeleteNotificationEndpoint(
// 	init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointService, string, func()),
// 	t *testing.T,
// ) {
// 	type args struct {
// 		ID string
// 	}
// 	type wants struct {
// 		err                   error
// 		notificationEndpoints []*influxdb.NotificationEndpoint
// 	}
//
// 	tests := []struct {
// 		name   string
// 		fields NotificationEndpointFields
// 		args   args
// 		wants  wants
// 	}{
// 		{
// 			name: "delete notificationEndpoints using exist id",
// 			fields: NotificationEndpointFields{
// 				Organizations: []*influxdb.Organization{
// 					{
// 						Name: "theorg",
// 						ID:   MustIDBase16(orgOneID),
// 					},
// 				},
// 				NotificationEndpoints: []*influxdb.NotificationEndpoint{
// 					{
// 						Name:  "A",
// 						ID:    MustIDBase16(notificationEndpointOneID),
// 						OrgID: MustIDBase16(orgOneID),
// 					},
// 					{
// 						Name:  "B",
// 						ID:    MustIDBase16(notificationEndpointThreeID),
// 						OrgID: MustIDBase16(orgOneID),
// 					},
// 				},
// 			},
// 			args: args{
// 				ID: notificationEndpointOneID,
// 			},
// 			wants: wants{
// 				notificationEndpoints: []*influxdb.NotificationEndpoint{
// 					{
// 						Name:  "B",
// 						ID:    MustIDBase16(notificationEndpointThreeID),
// 						OrgID: MustIDBase16(orgOneID),
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "delete notificationEndpoints using id that does not exist",
// 			fields: NotificationEndpointFields{
// 				Organizations: []*influxdb.Organization{
// 					{
// 						Name: "theorg",
// 						ID:   MustIDBase16(orgOneID),
// 					},
// 				},
// 				NotificationEndpoints: []*influxdb.NotificationEndpoint{
// 					{
// 						Name:  "A",
// 						ID:    MustIDBase16(notificationEndpointOneID),
// 						OrgID: MustIDBase16(orgOneID),
// 					},
// 					{
// 						Name:  "B",
// 						ID:    MustIDBase16(notificationEndpointThreeID),
// 						OrgID: MustIDBase16(orgOneID),
// 					},
// 				},
// 			},
// 			args: args{
// 				ID: "1234567890654321",
// 			},
// 			wants: wants{
// 				err: &influxdb.Error{
// 					Op:   influxdb.OpDeleteNotificationEndpoint,
// 					Msg:  "notificationEndpoint not found",
// 					Code: influxdb.ENotFound,
// 				},
// 				notificationEndpoints: []*influxdb.NotificationEndpoint{
// 					{
// 						Name:  "A",
// 						ID:    MustIDBase16(notificationEndpointOneID),
// 						OrgID: MustIDBase16(orgOneID),
// 					},
// 					{
// 						Name:  "B",
// 						ID:    MustIDBase16(notificationEndpointThreeID),
// 						OrgID: MustIDBase16(orgOneID),
// 					},
// 				},
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s, opPrefix, done := init(tt.fields, t)
// 			defer done()
// 			ctx := context.Background()
// 			err := s.DeleteNotificationEndpoint(ctx, MustIDBase16(tt.args.ID))
// 			diffPlatformErrors(tt.name, err, tt.wants.err, opPrefix, t)
//
// 			filter := influxdb.NotificationEndpointFilter{}
// 			notificationEndpoints, _, err := s.FindNotificationEndpoints(ctx, filter)
// 			if err != nil {
// 				t.Fatalf("failed to retrieve notificationEndpoints: %v", err)
// 			}
// 			if diff := cmp.Diff(notificationEndpoints, tt.wants.notificationEndpoints, notificationEndpointCmpOptions...); diff != "" {
// 				t.Errorf("notificationEndpoints are different -got/+want\ndiff %s", diff)
// 			}
// 		})
// 	}
// }
//
// // FindNotificationEndpoint testing
// func FindNotificationEndpoint(
// 	init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointService, string, func()),
// 	t *testing.T,
// ) {
// 	type args struct {
// 		name           string
// 		organizationID influxdb.ID
// 	}
//
// 	type wants struct {
// 		notificationEndpoint *influxdb.NotificationEndpoint
// 		err                  error
// 	}
//
// 	tests := []struct {
// 		name   string
// 		fields NotificationEndpointFields
// 		args   args
// 		wants  wants
// 	}{
// 		{
// 			name: "find notificationEndpoint by name",
// 			fields: NotificationEndpointFields{
// 				Organizations: []*influxdb.Organization{
// 					{
// 						Name: "theorg",
// 						ID:   MustIDBase16(orgOneID),
// 					},
// 				},
// 				NotificationEndpoints: []*influxdb.NotificationEndpoint{
// 					{
// 						ID:    MustIDBase16(notificationEndpointOneID),
// 						OrgID: MustIDBase16(orgOneID),
// 						Name:  "abc",
// 					},
// 					{
// 						ID:    MustIDBase16(notificationEndpointTwoID),
// 						OrgID: MustIDBase16(orgOneID),
// 						Name:  "xyz",
// 					},
// 				},
// 			},
// 			args: args{
// 				name:           "abc",
// 				organizationID: MustIDBase16(orgOneID),
// 			},
// 			wants: wants{
// 				notificationEndpoint: &influxdb.NotificationEndpoint{
// 					ID:    MustIDBase16(notificationEndpointOneID),
// 					OrgID: MustIDBase16(orgOneID),
// 					Name:  "abc",
// 				},
// 			},
// 		},
// 		{
// 			name: "missing notificationEndpoint returns error",
// 			fields: NotificationEndpointFields{
// 				Organizations: []*influxdb.Organization{
// 					{
// 						Name: "theorg",
// 						ID:   MustIDBase16(orgOneID),
// 					},
// 				},
// 				NotificationEndpoints: []*influxdb.NotificationEndpoint{},
// 			},
// 			args: args{
// 				name:           "xyz",
// 				organizationID: MustIDBase16(orgOneID),
// 			},
// 			wants: wants{
// 				err: &influxdb.Error{
// 					Code: influxdb.ENotFound,
// 					Op:   influxdb.OpFindNotificationEndpoint,
// 					Msg:  "notificationEndpoint \"xyz\" not found",
// 				},
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s, opPrefix, done := init(tt.fields, t)
// 			defer done()
// 			ctx := context.Background()
// 			filter := influxdb.NotificationEndpointFilter{}
// 			if tt.args.name != "" {
// 				filter.Name = &tt.args.name
// 			}
// 			if tt.args.organizationID.Valid() {
// 				filter.OrganizationID = &tt.args.organizationID
// 			}
//
// 			notificationEndpoint, err := s.FindNotificationEndpoint(ctx, filter)
// 			diffPlatformErrors(tt.name, err, tt.wants.err, opPrefix, t)
//
// 			if diff := cmp.Diff(notificationEndpoint, tt.wants.notificationEndpoint, notificationEndpointCmpOptions...); diff != "" {
// 				t.Errorf("notificationEndpoints are different -got/+want\ndiff %s", diff)
// 			}
// 		})
// 	}
// }
//
// // UpdateNotificationEndpoint testing
// func UpdateNotificationEndpoint(
// 	init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointService, string, func()),
// 	t *testing.T,
// ) {
// 	type args struct {
// 		name        string
// 		id          influxdb.ID
// 		retention   int
// 		description *string
// 	}
// 	type wants struct {
// 		err                  error
// 		notificationEndpoint *influxdb.NotificationEndpoint
// 	}
//
// 	tests := []struct {
// 		name   string
// 		fields NotificationEndpointFields
// 		args   args
// 		wants  wants
// 	}{
// 		{
// 			name: "update name",
// 			fields: NotificationEndpointFields{
// 				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
// 				Organizations: []*influxdb.Organization{
// 					{
// 						Name: "theorg",
// 						ID:   MustIDBase16(orgOneID),
// 					},
// 				},
// 				NotificationEndpoints: []*influxdb.NotificationEndpoint{
// 					{
// 						ID:    MustIDBase16(notificationEndpointOneID),
// 						OrgID: MustIDBase16(orgOneID),
// 						Name:  "notificationEndpoint1",
// 					},
// 					{
// 						ID:    MustIDBase16(notificationEndpointTwoID),
// 						OrgID: MustIDBase16(orgOneID),
// 						Name:  "notificationEndpoint2",
// 					},
// 				},
// 			},
// 			args: args{
// 				id:   MustIDBase16(notificationEndpointOneID),
// 				name: "changed",
// 			},
// 			wants: wants{
// 				notificationEndpoint: &influxdb.NotificationEndpoint{
// 					ID:    MustIDBase16(notificationEndpointOneID),
// 					OrgID: MustIDBase16(orgOneID),
// 					Name:  "changed",
// 					CRUDLog: influxdb.CRUDLog{
// 						UpdatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "update name unique",
// 			fields: NotificationEndpointFields{
// 				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
// 				Organizations: []*influxdb.Organization{
// 					{
// 						Name: "theorg",
// 						ID:   MustIDBase16(orgOneID),
// 					},
// 				},
// 				NotificationEndpoints: []*influxdb.NotificationEndpoint{
// 					{
// 						ID:    MustIDBase16(notificationEndpointOneID),
// 						OrgID: MustIDBase16(orgOneID),
// 						Name:  "notificationEndpoint1",
// 					},
// 					{
// 						ID:    MustIDBase16(notificationEndpointTwoID),
// 						OrgID: MustIDBase16(orgOneID),
// 						Name:  "notificationEndpoint2",
// 					},
// 				},
// 			},
// 			args: args{
// 				id:   MustIDBase16(notificationEndpointOneID),
// 				name: "notificationEndpoint2",
// 			},
// 			wants: wants{
// 				err: &influxdb.Error{
// 					Code: influxdb.EConflict,
// 					Msg:  "notificationEndpoint name is not unique",
// 				},
// 			},
// 		},
// 		{
// 			name: "update retention",
// 			fields: NotificationEndpointFields{
// 				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
// 				Organizations: []*influxdb.Organization{
// 					{
// 						Name: "theorg",
// 						ID:   MustIDBase16(orgOneID),
// 					},
// 				},
// 				NotificationEndpoints: []*influxdb.NotificationEndpoint{
// 					{
// 						ID:    MustIDBase16(notificationEndpointOneID),
// 						OrgID: MustIDBase16(orgOneID),
// 						Name:  "notificationEndpoint1",
// 					},
// 					{
// 						ID:    MustIDBase16(notificationEndpointTwoID),
// 						OrgID: MustIDBase16(orgOneID),
// 						Name:  "notificationEndpoint2",
// 					},
// 				},
// 			},
// 			args: args{
// 				id:        MustIDBase16(notificationEndpointOneID),
// 				retention: 100,
// 			},
// 			wants: wants{
// 				notificationEndpoint: &influxdb.NotificationEndpoint{
// 					ID:              MustIDBase16(notificationEndpointOneID),
// 					OrgID:           MustIDBase16(orgOneID),
// 					Name:            "notificationEndpoint1",
// 					RetentionPeriod: 100 * time.Minute,
// 					CRUDLog: influxdb.CRUDLog{
// 						UpdatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "update description",
// 			fields: NotificationEndpointFields{
// 				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
// 				Organizations: []*influxdb.Organization{
// 					{
// 						Name: "theorg",
// 						ID:   MustIDBase16(orgOneID),
// 					},
// 				},
// 				NotificationEndpoints: []*influxdb.NotificationEndpoint{
// 					{
// 						ID:    MustIDBase16(notificationEndpointOneID),
// 						OrgID: MustIDBase16(orgOneID),
// 						Name:  "notificationEndpoint1",
// 					},
// 					{
// 						ID:    MustIDBase16(notificationEndpointTwoID),
// 						OrgID: MustIDBase16(orgOneID),
// 						Name:  "notificationEndpoint2",
// 					},
// 				},
// 			},
// 			args: args{
// 				id:          MustIDBase16(notificationEndpointOneID),
// 				description: stringPtr("desc1"),
// 			},
// 			wants: wants{
// 				notificationEndpoint: &influxdb.NotificationEndpoint{
// 					ID:          MustIDBase16(notificationEndpointOneID),
// 					OrgID:       MustIDBase16(orgOneID),
// 					Name:        "notificationEndpoint1",
// 					Description: "desc1",
// 					CRUDLog: influxdb.CRUDLog{
// 						UpdatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "update retention and name",
// 			fields: NotificationEndpointFields{
// 				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
// 				Organizations: []*influxdb.Organization{
// 					{
// 						Name: "theorg",
// 						ID:   MustIDBase16(orgOneID),
// 					},
// 				},
// 				NotificationEndpoints: []*influxdb.NotificationEndpoint{
// 					{
// 						ID:    MustIDBase16(notificationEndpointOneID),
// 						OrgID: MustIDBase16(orgOneID),
// 						Name:  "notificationEndpoint1",
// 					},
// 					{
// 						ID:    MustIDBase16(notificationEndpointTwoID),
// 						OrgID: MustIDBase16(orgOneID),
// 						Name:  "notificationEndpoint2",
// 					},
// 				},
// 			},
// 			args: args{
// 				id:        MustIDBase16(notificationEndpointTwoID),
// 				retention: 101,
// 				name:      "changed",
// 			},
// 			wants: wants{
// 				notificationEndpoint: &influxdb.NotificationEndpoint{
// 					ID:              MustIDBase16(notificationEndpointTwoID),
// 					OrgID:           MustIDBase16(orgOneID),
// 					Name:            "changed",
// 					RetentionPeriod: 101 * time.Minute,
// 					CRUDLog: influxdb.CRUDLog{
// 						UpdatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "update retention and same name",
// 			fields: NotificationEndpointFields{
// 				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
// 				Organizations: []*influxdb.Organization{
// 					{
// 						Name: "theorg",
// 						ID:   MustIDBase16(orgOneID),
// 					},
// 				},
// 				NotificationEndpoints: []*influxdb.NotificationEndpoint{
// 					{
// 						ID:    MustIDBase16(notificationEndpointOneID),
// 						OrgID: MustIDBase16(orgOneID),
// 						Name:  "notificationEndpoint1",
// 					},
// 					{
// 						ID:    MustIDBase16(notificationEndpointTwoID),
// 						OrgID: MustIDBase16(orgOneID),
// 						Name:  "notificationEndpoint2",
// 					},
// 				},
// 			},
// 			args: args{
// 				id:        MustIDBase16(notificationEndpointTwoID),
// 				retention: 101,
// 				name:      "notificationEndpoint2",
// 			},
// 			wants: wants{
// 				notificationEndpoint: &influxdb.NotificationEndpoint{
// 					ID:              MustIDBase16(notificationEndpointTwoID),
// 					OrgID:           MustIDBase16(orgOneID),
// 					Name:            "notificationEndpoint2",
// 					RetentionPeriod: 101 * time.Minute,
// 					CRUDLog: influxdb.CRUDLog{
// 						UpdatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
// 					},
// 				},
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s, opPrefix, done := init(tt.fields, t)
// 			defer done()
// 			ctx := context.Background()
//
// 			upd := influxdb.NotificationEndpointUpdate{}
// 			if tt.args.name != "" {
// 				upd.Name = &tt.args.name
// 			}
// 			if tt.args.retention != 0 {
// 				d := time.Duration(tt.args.retention) * time.Minute
// 				upd.RetentionPeriod = &d
// 			}
//
// 			upd.Description = tt.args.description
//
// 			notificationEndpoint, err := s.UpdateNotificationEndpoint(ctx, tt.args.id, upd)
// 			diffPlatformErrors(tt.name, err, tt.wants.err, opPrefix, t)
//
// 			if diff := cmp.Diff(notificationEndpoint, tt.wants.notificationEndpoint, notificationEndpointCmpOptions...); diff != "" {
// 				t.Errorf("notificationEndpoint is different -got/+want\ndiff %s", diff)
// 			}
// 		})
// 	}
// }
