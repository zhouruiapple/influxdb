package testing

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/mock"
	"github.com/influxdata/influxdb/notification"
	"github.com/influxdata/influxdb/notification/endpoint"
)

// NotificationEndpointFields includes prepopulated data for mapping tests.
type NotificationEndpointFields struct {
	IDGenerator           influxdb.IDGenerator
	TimeGenerator         influxdb.TimeGenerator
	NotificationEndpoints []influxdb.NotificationEndpoint
	Orgs                  []*influxdb.Organization
}

// var timeGen1 = mock.TimeGenerator{FakeValue: time.Date(2006, time.July, 13, 4, 19, 10, 0, time.UTC)}
// var timeGen2 = mock.TimeGenerator{FakeValue: time.Date(2006, time.July, 14, 5, 23, 53, 10, time.UTC)}
// var time3 = time.Date(2006, time.July, 15, 5, 23, 53, 10, time.UTC)

var notificationEndpointCmpOptions = cmp.Options{
	cmp.Transformer("Sort", func(in []influxdb.NotificationEndpoint) []influxdb.NotificationEndpoint {
		out := append([]influxdb.NotificationEndpoint(nil), in...)
		sort.Slice(out, func(i, j int) bool {
			return out[i].GetID() > out[j].GetID()
		})
		return out
	}),
}

// NotificationEndpointStore tests all the service functions.
func NotificationEndpointStore(
	init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointStore, func()), t *testing.T,
) {
	tests := []struct {
		name string
		fn   func(init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointStore, func()),
			t *testing.T)
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
		{
			name: "UpdateNotificationEndpoint",
			fn:   UpdateNotificationEndpoint,
		},
		{
			name: "PatchNotificationEndpoint",
			fn:   PatchNotificationEndpoint,
		},
		{
			name: "DeleteNotificationEndpoint",
			fn:   DeleteNotificationEndpoint,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(init, t)
		})
	}
}

// CreateNotificationEndpoint testing.
func CreateNotificationEndpoint(
	init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointStore, func()),
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
			name: "basic create notification endpoint",
			fields: NotificationEndpointFields{
				IDGenerator:   mock.NewIDGenerator(twoID, t),
				TimeGenerator: fakeGenerator,
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							OrgID:           MustIDBase16(fourID),
							UserID:          MustIDBase16(twoID),
							Description:     "description1",
							AuthorizationID: MustIDBase16(threeID),
							Name:            "name1",
							Status:          influxdb.Active,
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						URL:   "url1",
						Token: influxdb.SecretField("secret1"),
					},
				},
			},
			args: args{
				userID: MustIDBase16(sixID),
				notificationEndpoint: &endpoint.SMTP{
					Base: endpoint.Base{
						ID:              MustIDBase16(threeID),
						OrgID:           MustIDBase16(fourID),
						UserID:          MustIDBase16(fiveID),
						Description:     "description2",
						AuthorizationID: MustIDBase16(oneID),
						Name:            "name2",
						Status:          influxdb.Active,
					},
					URL:   "url2",
					Token: influxdb.SecretField("secret2"),
				},
			},
			wants: wants{
				notificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							TagEndpoints: []notification.TagEndpoint{
								{
									Tag: notification.Tag{
										Key:   "k1",
										Value: "v1",
									},
									Operator: notification.NotEqual,
								},
								{
									Tag: notification.Tag{
										Key:   "k2",
										Value: "v2",
									},
									Operator: notification.RegexEqual,
								},
							},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.SMTP{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							EndpointID:      IDPtr(MustIDBase16(fiveID)),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							TagEndpoints: []notification.TagEndpoint{
								{
									Tag: notification.Tag{
										Key:   "k1",
										Value: "v1",
									},
									Operator: notification.NotEqual,
								},
								{
									Tag: notification.Tag{
										Key:   "k2",
										Value: "v2",
									},
									Operator: notification.RegexEqual,
								},
							},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: fakeDate,
								UpdatedAt: fakeDate,
							},
						},
						SubjectTemp: "subject1",
						To:          "example@host.com",
						BodyTemp:    "msg1",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()
			err := s.CreateNotificationEndpoint(ctx, tt.args.notificationEndpoint, tt.args.userID)
			if (err != nil) != (tt.wants.err != nil) {
				t.Fatalf("expected error '%v' got '%v'", tt.wants.err, err)
			}
			if tt.wants.err == nil && !tt.args.notificationEndpoint.GetID().Valid() {
				t.Fatalf("notification endpoint ID not set from CreateNotificationEndpoint")
			}

			if err != nil && tt.wants.err != nil {
				if influxdb.ErrorCode(err) != influxdb.ErrorCode(tt.wants.err) {
					t.Fatalf("expected error messages to match '%v' got '%v'", influxdb.ErrorCode(tt.wants.err), influxdb.ErrorCode(err))
				}
			}

			nrs, _, err := s.FindNotificationEndpoints(ctx, influxdb.NotificationEndpointFilter{})
			if err != nil {
				t.Fatalf("failed to retrieve notification endpoints: %v", err)
			}
			if diff := cmp.Diff(nrs, tt.wants.notificationEndpoints, notificationEndpointCmpOptions...); diff != "" {
				t.Errorf("notificationEndpoints are different -got/+want\ndiff %s", diff)
			}
		})
	}
}

// FindNotificationEndpointByID testing.
func FindNotificationEndpointByID(
	init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointStore, func()),
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
			name: "bad id",
			fields: NotificationEndpointFields{
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink2",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				id: influxdb.ID(0),
			},
			wants: wants{
				err: &influxdb.Error{
					Code: influxdb.EInvalid,
					Msg:  "provided notification endpoint ID has invalid format",
				},
			},
		},
		{
			name: "not found",
			fields: NotificationEndpointFields{
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink2",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				id: MustIDBase16(threeID),
			},
			wants: wants{
				err: &influxdb.Error{
					Code: influxdb.ENotFound,
					Msg:  "notification endpoint not found",
				},
			},
		},
		{
			name: "basic find telegraf config by id",
			fields: NotificationEndpointFields{
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink2",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				id: MustIDBase16(twoID),
			},
			wants: wants{
				notificationEndpoint: &endpoint.PagerDuty{
					Base: endpoint.Base{
						ID:              MustIDBase16(twoID),
						Name:            "name2",
						AuthorizationID: MustIDBase16(threeID),
						OrgID:           MustIDBase16(fourID),
						Status:          influxdb.Active,
						RunbookLink:     "runbooklink2",
						SleepUntil:      &time3,
						Every:           influxdb.Duration{Duration: time.Hour},
						CRUDLog: influxdb.CRUDLog{
							CreatedAt: timeGen1.Now(),
							UpdatedAt: timeGen2.Now(),
						},
					},
					MessageTemp: "msg",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()

			nr, err := s.FindNotificationEndpointByID(ctx, tt.args.id)
			ErrorsEqual(t, err, tt.wants.err)
			if diff := cmp.Diff(nr, tt.wants.notificationEndpoint, notificationEndpointCmpOptions...); diff != "" {
				t.Errorf("notification endpoint is different -got/+want\ndiff %s", diff)
			}
		})
	}
}

// FindNotificationEndpoints testing
func FindNotificationEndpoints(
	init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointStore, func()),
	t *testing.T,
) {
	type args struct {
		filter influxdb.NotificationEndpointFilter
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
			name: "find nothing (empty set)",
			fields: NotificationEndpointFields{
				UserResourceMappings:  []*influxdb.UserResourceMapping{},
				NotificationEndpoints: []influxdb.NotificationEndpoint{},
			},
			args: args{
				filter: influxdb.NotificationEndpointFilter{
					UserResourceMappingFilter: influxdb.UserResourceMappingFilter{
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
			},
			wants: wants{
				notificationEndpoints: []influxdb.NotificationEndpoint{},
			},
		},
		{
			name: "find all notification endpoints",
			fields: NotificationEndpointFields{
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink2",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				filter: influxdb.NotificationEndpointFilter{
					UserResourceMappingFilter: influxdb.UserResourceMappingFilter{
						UserID:       MustIDBase16(sixID),
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
			},
			wants: wants{
				notificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink2",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						MessageTemp: "msg",
					},
				},
			},
		},
		{
			name: "find owners only",
			fields: NotificationEndpointFields{
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink2",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				filter: influxdb.NotificationEndpointFilter{
					UserResourceMappingFilter: influxdb.UserResourceMappingFilter{
						UserID:       MustIDBase16(sixID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserType:     influxdb.Owner,
					},
				},
			},
			wants: wants{
				notificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
				},
			},
		},
		{
			name: "filter by organization id only",
			fields: NotificationEndpointFields{
				Orgs: []*influxdb.Organization{
					{
						ID:   MustIDBase16(oneID),
						Name: "org1",
					},
					{
						ID:   MustIDBase16(fourID),
						Name: "org4",
					},
				},
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
					},
					{
						ResourceID:   MustIDBase16(fourID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							OrgID:           MustIDBase16(fourID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr1",
						},
						Channel:         "ch1",
						MessageTemplate: "msg1",
					},
					&endpoint.SMTP{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							OrgID:           MustIDBase16(fourID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr2",
						},
						SubjectTemp: "subject2",
						To:          "astA@fadac.com",
						BodyTemp:    "body2",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(fourID),
							OrgID:           MustIDBase16(oneID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr3",
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				filter: influxdb.NotificationEndpointFilter{
					OrgID: idPtr(MustIDBase16(oneID)),
				},
			},
			wants: wants{
				notificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(fourID),
							OrgID:           MustIDBase16(oneID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr3",
						},
						MessageTemp: "msg",
					},
				},
			},
		},
		{
			name: "filter by organization name only",
			fields: NotificationEndpointFields{
				Orgs: []*influxdb.Organization{
					{
						ID:   MustIDBase16(oneID),
						Name: "org1",
					},
					{
						ID:   MustIDBase16(fourID),
						Name: "org4",
					},
				},
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
					},
					{
						ResourceID:   MustIDBase16(fourID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							OrgID:           MustIDBase16(fourID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr1",
						},
						Channel:         "ch1",
						MessageTemplate: "msg1",
					},
					&endpoint.SMTP{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							OrgID:           MustIDBase16(fourID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr2",
						},
						SubjectTemp: "subject2",
						To:          "astA@fadac.com",
						BodyTemp:    "body2",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(fourID),
							OrgID:           MustIDBase16(oneID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr3",
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				filter: influxdb.NotificationEndpointFilter{
					Organization: strPtr("org4"),
				},
			},
			wants: wants{
				notificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							OrgID:           MustIDBase16(fourID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr1",
						},
						Channel:         "ch1",
						MessageTemplate: "msg1",
					},
					&endpoint.SMTP{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							OrgID:           MustIDBase16(fourID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr2",
						},
						SubjectTemp: "subject2",
						To:          "astA@fadac.com",
						BodyTemp:    "body2",
					},
				},
			},
		},
		{
			name: "find owners and restrict by organization",
			fields: NotificationEndpointFields{
				Orgs: []*influxdb.Organization{
					{
						ID:   MustIDBase16(oneID),
						Name: "org1",
					},
					{
						ID:   MustIDBase16(fourID),
						Name: "org4",
					},
				},
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
					},
					{
						ResourceID:   MustIDBase16(fourID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							OrgID:           MustIDBase16(fourID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr1",
						},
						Channel:         "ch1",
						MessageTemplate: "msg1",
					},
					&endpoint.SMTP{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							OrgID:           MustIDBase16(fourID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr2",
						},
						SubjectTemp: "subject2",
						To:          "astA@fadac.com",
						BodyTemp:    "body2",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(fourID),
							OrgID:           MustIDBase16(oneID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr3",
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				filter: influxdb.NotificationEndpointFilter{
					OrgID: idPtr(MustIDBase16(oneID)),
					UserResourceMappingFilter: influxdb.UserResourceMappingFilter{
						UserID:       MustIDBase16(sixID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserType:     influxdb.Owner,
					},
				},
			},
			wants: wants{
				notificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(fourID),
							OrgID:           MustIDBase16(oneID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr3",
						},
						MessageTemp: "msg",
					},
				},
			},
		},
		{
			name: "look for organization not bound to any notification endpoint",
			fields: NotificationEndpointFields{
				Orgs: []*influxdb.Organization{
					{
						ID:   MustIDBase16(oneID),
						Name: "org1",
					},
					{
						ID:   MustIDBase16(fourID),
						Name: "org4",
					},
				},
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
					},
					{
						ResourceID:   MustIDBase16(fourID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							OrgID:           MustIDBase16(fourID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr1",
						},
						Channel:         "ch1",
						MessageTemplate: "msg1",
					},
					&endpoint.SMTP{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							OrgID:           MustIDBase16(fourID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr2",
						},
						SubjectTemp: "subject2",
						To:          "astA@fadac.com",
						BodyTemp:    "body2",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(fourID),
							OrgID:           MustIDBase16(fourID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr3",
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				filter: influxdb.NotificationEndpointFilter{
					OrgID: idPtr(MustIDBase16(oneID)),
				},
			},
			wants: wants{
				notificationEndpoints: []influxdb.NotificationEndpoint{},
			},
		},
		{
			name: "find nothing",
			fields: NotificationEndpointFields{
				Orgs: []*influxdb.Organization{
					{
						ID:   MustIDBase16(oneID),
						Name: "org1",
					},
					{
						ID:   MustIDBase16(fourID),
						Name: "org4",
					},
				},
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
					},
					{
						ResourceID:   MustIDBase16(fourID),
						ResourceType: influxdb.NotificationEndpointResourceType,
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							OrgID:           MustIDBase16(fourID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr1",
						},
						Channel:         "ch1",
						MessageTemplate: "msg1",
					},
					&endpoint.SMTP{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							OrgID:           MustIDBase16(fourID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr2",
						},
						SubjectTemp: "subject2",
						To:          "astA@fadac.com",
						BodyTemp:    "body2",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(fourID),
							OrgID:           MustIDBase16(fourID),
							AuthorizationID: MustIDBase16(threeID),
							Status:          influxdb.Active,
							Name:            "nr3",
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				filter: influxdb.NotificationEndpointFilter{
					UserResourceMappingFilter: influxdb.UserResourceMappingFilter{
						UserID:       MustIDBase16(fourID),
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
			},
			wants: wants{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()

			nrs, n, err := s.FindNotificationEndpoints(ctx, tt.args.filter)
			ErrorsEqual(t, err, tt.wants.err)
			if n != len(tt.wants.notificationEndpoints) {
				t.Fatalf("notification endpoints length is different got %d, want %d", n, len(tt.wants.notificationEndpoints))
			}

			if diff := cmp.Diff(nrs, tt.wants.notificationEndpoints, notificationEndpointCmpOptions...); diff != "" {
				t.Errorf("notification endpoints are different -got/+want\ndiff %s", diff)
			}
		})
	}
}

// UpdateNotificationEndpoint testing.
func UpdateNotificationEndpoint(
	init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointStore, func()),
	t *testing.T,
) {
	type args struct {
		userID               influxdb.ID
		id                   influxdb.ID
		notificationEndpoint influxdb.NotificationEndpoint
	}

	type wants struct {
		notificationEndpoint influxdb.NotificationEndpoint
		err                  error
	}
	tests := []struct {
		name   string
		fields NotificationEndpointFields
		args   args
		wants  wants
	}{
		{
			name: "can't find the id",
			fields: NotificationEndpointFields{
				TimeGenerator: fakeGenerator,
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink2",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				userID: MustIDBase16(sixID),
				id:     MustIDBase16(fourID),
				notificationEndpoint: &endpoint.PagerDuty{
					Base: endpoint.Base{
						ID:              MustIDBase16(twoID),
						Name:            "name2",
						AuthorizationID: MustIDBase16(threeID),
						OrgID:           MustIDBase16(fourID),
						Status:          influxdb.Inactive,
						RunbookLink:     "runbooklink3",
						SleepUntil:      &time3,
						Every:           influxdb.Duration{Duration: time.Hour * 2},
					},
					MessageTemp: "msg2",
				},
			},
			wants: wants{
				err: &influxdb.Error{
					Code: influxdb.ENotFound,
					Msg:  "notification endpoint not found",
				},
			},
		},
		{
			name: "regular update",
			fields: NotificationEndpointFields{
				TimeGenerator: fakeGenerator,
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink2",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				userID: MustIDBase16(sixID),
				id:     MustIDBase16(twoID),
				notificationEndpoint: &endpoint.PagerDuty{
					Base: endpoint.Base{
						AuthorizationID: MustIDBase16(threeID),
						Name:            "name3",
						OrgID:           MustIDBase16(fourID),
						Status:          influxdb.Inactive,
						RunbookLink:     "runbooklink3",
						SleepUntil:      &time3,
						Every:           influxdb.Duration{Duration: time.Hour * 2},
					},
					MessageTemp: "msg2",
				},
			},
			wants: wants{
				notificationEndpoint: &endpoint.PagerDuty{
					Base: endpoint.Base{
						ID:              MustIDBase16(twoID),
						Name:            "name3",
						AuthorizationID: MustIDBase16(threeID),
						OrgID:           MustIDBase16(fourID),
						Status:          influxdb.Inactive,
						RunbookLink:     "runbooklink3",
						SleepUntil:      &time3,
						Every:           influxdb.Duration{Duration: time.Hour * 2},
						CRUDLog: influxdb.CRUDLog{
							CreatedAt: timeGen1.Now(),
							UpdatedAt: fakeDate,
						},
					},
					MessageTemp: "msg2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()

			tc, err := s.UpdateNotificationEndpoint(ctx, tt.args.id,
				tt.args.notificationEndpoint, tt.args.userID)
			ErrorsEqual(t, err, tt.wants.err)
			if diff := cmp.Diff(tc, tt.wants.notificationEndpoint, notificationEndpointCmpOptions...); tt.wants.err == nil && diff != "" {
				t.Errorf("notificationEndpoints are different -got/+want\ndiff %s", diff)
			}
		})
	}
}

// PatchNotificationEndpoint testing.
func PatchNotificationEndpoint(
	init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointStore, func()),
	t *testing.T,
) {

	name3 := "name2"
	status3 := influxdb.Inactive

	type args struct {
		//userID           influxdb.ID
		id  influxdb.ID
		upd influxdb.NotificationEndpointUpdate
	}

	type wants struct {
		notificationEndpoint influxdb.NotificationEndpoint
		err                  error
	}
	tests := []struct {
		name   string
		fields NotificationEndpointFields
		args   args
		wants  wants
	}{
		{
			name: "can't find the id",
			fields: NotificationEndpointFields{
				TimeGenerator: fakeGenerator,
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink2",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				id: MustIDBase16(fourID),
				upd: influxdb.NotificationEndpointUpdate{
					Name:   &name3,
					Status: &status3,
				},
			},
			wants: wants{
				err: &influxdb.Error{
					Code: influxdb.ENotFound,
					Msg:  "notification endpoint not found",
				},
			},
		},
		{
			name: "regular update",
			fields: NotificationEndpointFields{
				TimeGenerator: fakeGenerator,
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							Status:          influxdb.Active,
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							Status:          influxdb.Active,
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							RunbookLink:     "runbooklink2",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				id: MustIDBase16(twoID),
				upd: influxdb.NotificationEndpointUpdate{
					Name:   &name3,
					Status: &status3,
				},
			},
			wants: wants{
				notificationEndpoint: &endpoint.PagerDuty{
					Base: endpoint.Base{
						ID:              MustIDBase16(twoID),
						Name:            name3,
						Status:          status3,
						AuthorizationID: MustIDBase16(threeID),
						OrgID:           MustIDBase16(fourID),
						RunbookLink:     "runbooklink2",
						SleepUntil:      &time3,
						Every:           influxdb.Duration{Duration: time.Hour},
						CRUDLog: influxdb.CRUDLog{
							CreatedAt: timeGen1.Now(),
							UpdatedAt: fakeDate,
						},
					},
					MessageTemp: "msg",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()

			tc, err := s.PatchNotificationEndpoint(ctx, tt.args.id, tt.args.upd)
			ErrorsEqual(t, err, tt.wants.err)
			if diff := cmp.Diff(tc, tt.wants.notificationEndpoint, notificationEndpointCmpOptions...); tt.wants.err == nil && diff != "" {
				t.Errorf("notificationEndpoints are different -got/+want\ndiff %s", diff)
			}
		})
	}
}

// DeleteNotificationEndpoint testing.
func DeleteNotificationEndpoint(
	init func(NotificationEndpointFields, *testing.T) (influxdb.NotificationEndpointStore, func()),
	t *testing.T,
) {
	type args struct {
		id     influxdb.ID
		userID influxdb.ID
	}

	type wants struct {
		notificationEndpoints []influxdb.NotificationEndpoint
		userResourceMappings  []*influxdb.UserResourceMapping
		err                   error
	}
	tests := []struct {
		name   string
		fields NotificationEndpointFields
		args   args
		wants  wants
	}{
		{
			name: "bad id",
			fields: NotificationEndpointFields{
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink2",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				id:     influxdb.ID(0),
				userID: MustIDBase16(sixID),
			},
			wants: wants{
				err: &influxdb.Error{
					Code: influxdb.EInvalid,
					Msg:  "provided notification endpoint ID has invalid format",
				},
				userResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
				notificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink2",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						MessageTemp: "msg",
					},
				},
			},
		},
		{
			name: "none existing config",
			fields: NotificationEndpointFields{
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink2",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				id:     MustIDBase16(fourID),
				userID: MustIDBase16(sixID),
			},
			wants: wants{
				err: &influxdb.Error{
					Code: influxdb.ENotFound,
					Msg:  "notification endpoint not found",
				},
				userResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
				notificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink2",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						MessageTemp: "msg",
					},
				},
			},
		},
		{
			name: "regular delete",
			fields: NotificationEndpointFields{
				UserResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
					{
						ResourceID:   MustIDBase16(twoID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Member,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
				NotificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
					&endpoint.PagerDuty{
						Base: endpoint.Base{
							ID:              MustIDBase16(twoID),
							Name:            "name2",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink2",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						MessageTemp: "msg",
					},
				},
			},
			args: args{
				id:     MustIDBase16(twoID),
				userID: MustIDBase16(sixID),
			},
			wants: wants{
				userResourceMappings: []*influxdb.UserResourceMapping{
					{
						ResourceID:   MustIDBase16(oneID),
						UserID:       MustIDBase16(sixID),
						UserType:     influxdb.Owner,
						ResourceType: influxdb.NotificationEndpointResourceType,
					},
				},
				notificationEndpoints: []influxdb.NotificationEndpoint{
					&endpoint.Slack{
						Base: endpoint.Base{
							ID:              MustIDBase16(oneID),
							Name:            "name1",
							AuthorizationID: MustIDBase16(threeID),
							OrgID:           MustIDBase16(fourID),
							Status:          influxdb.Active,
							RunbookLink:     "runbooklink1",
							SleepUntil:      &time3,
							Every:           influxdb.Duration{Duration: time.Hour},
							CRUDLog: influxdb.CRUDLog{
								CreatedAt: timeGen1.Now(),
								UpdatedAt: timeGen2.Now(),
							},
						},
						Channel:         "channel1",
						MessageTemplate: "msg1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()
			err := s.DeleteNotificationEndpoint(ctx, tt.args.id)
			ErrorsEqual(t, err, tt.wants.err)

			filter := influxdb.NotificationEndpointFilter{
				UserResourceMappingFilter: influxdb.UserResourceMappingFilter{
					UserID:       tt.args.userID,
					ResourceType: influxdb.NotificationEndpointResourceType,
				},
			}
			nrs, n, err := s.FindNotificationEndpoints(ctx, filter)
			if err != nil && tt.wants.err == nil {
				t.Fatalf("expected errors to be nil got '%v'", err)
			}

			if err != nil && tt.wants.err != nil {
				if want, got := tt.wants.err.Error(), err.Error(); want != got {
					t.Fatalf("expected error '%v' got '%v'", tt.wants.err, err)
				}
			}

			if n != len(tt.wants.notificationEndpoints) {
				t.Fatalf("notification endpoints length is different got %d, want %d", n, len(tt.wants.notificationEndpoints))
			}
			if diff := cmp.Diff(nrs, tt.wants.notificationEndpoints, notificationEndpointCmpOptions...); diff != "" {
				t.Errorf("notification endpoints are different -got/+want\ndiff %s", diff)
			}

			urms, _, err := s.FindUserResourceMappings(ctx, influxdb.UserResourceMappingFilter{
				UserID:       tt.args.userID,
				ResourceType: influxdb.NotificationEndpointResourceType,
			})
			if err != nil {
				t.Fatalf("failed to retrieve user resource mappings: %v", err)
			}
			if diff := cmp.Diff(urms, tt.wants.userResourceMappings, userResourceMappingCmpOptions...); diff != "" {
				t.Errorf("user resource mappings are different -got/+want\ndiff %s", diff)
			}
		})
	}
}
