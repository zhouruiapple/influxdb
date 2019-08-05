package kv_test

import (
	"context"
	"testing"

	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/kv"
	influxdbtesting "github.com/influxdata/influxdb/testing"
)

func TestBoltNotificationEndpointStore(t *testing.T) {
	influxdbtesting.NotificationEndpointStore(initBoltNotificationEndpointStore, t)
}

func TestNotificationEndpointStore(t *testing.T) {
	influxdbtesting.NotificationEndpointStore(initInmemNotificationEndpointStore, t)
}

func initBoltNotificationEndpointStore(f influxdbtesting.NotificationEndpointFields, t *testing.T) (influxdb.NotificationEndpointStore, func()) {
	s, closeBolt, err := NewTestBoltStore()
	if err != nil {
		t.Fatalf("failed to create new kv store: %v", err)
	}

	svc, closeSvc := initNotificationEndpointStore(s, f, t)
	return svc, func() {
		closeSvc()
		closeBolt()
	}
}

func initInmemNotificationEndpointStore(f influxdbtesting.NotificationEndpointFields, t *testing.T) (influxdb.NotificationEndpointStore, func()) {
	s, closeBolt, err := NewTestInmemStore()
	if err != nil {
		t.Fatalf("failed to create new kv store: %v", err)
	}

	svc, closeSvc := initNotificationEndpointStore(s, f, t)
	return svc, func() {
		closeSvc()
		closeBolt()
	}
}

func initNotificationEndpointStore(s kv.Store, f influxdbtesting.NotificationEndpointFields, t *testing.T) (influxdb.NotificationEndpointStore, func()) {
	svc := kv.NewService(s)
	svc.IDGenerator = f.IDGenerator
	svc.TimeGenerator = f.TimeGenerator
	if f.TimeGenerator == nil {
		svc.TimeGenerator = influxdb.RealTimeGenerator{}
	}

	ctx := context.Background()
	if err := svc.Initialize(ctx); err != nil {
		t.Fatalf("error initializing user service: %v", err)
	}

	for _, nr := range f.NotificationEndpoints {
		if err := svc.PutNotificationEndpoint(ctx, nr); err != nil {
			t.Fatalf("failed to populate notification endpoint: %v", err)
		}
	}

	for _, m := range f.UserResourceMappings {
		if err := svc.CreateUserResourceMapping(ctx, m); err != nil {
			t.Fatalf("failed to populate user resource mapping: %v", err)
		}
	}

	for _, o := range f.Orgs {
		if err := svc.PutOrganization(ctx, o); err != nil {
			t.Fatalf("failed to populate org: %v", err)
		}
	}

	return svc, func() {
		for _, nr := range f.NotificationEndpoints {
			if err := svc.DeleteNotificationEndpoint(ctx, nr.GetID()); err != nil {
				t.Logf("failed to remove notification endpoint: %v", err)
			}
		}
		for _, urm := range f.UserResourceMappings {
			if err := svc.DeleteUserResourceMapping(ctx, urm.ResourceID, urm.UserID); err != nil && influxdb.ErrorCode(err) != influxdb.ENotFound {
				t.Logf("failed to remove urm endpoint: %v", err)
			}
		}
		for _, o := range f.Orgs {
			if err := svc.DeleteOrganization(ctx, o.ID); err != nil {
				t.Fatalf("failed to remove org: %v", err)
			}
		}
	}
}
