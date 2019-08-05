package kv_test

import (
	"context"
	"testing"

	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/kv"
	influxdbtesting "github.com/influxdata/influxdb/testing"
)

func TestNotificationEndpointService(t *testing.T) {
	influxdbtesting.NotificationEndpointService(initInmemNotificationEndpointService, t)
}

func initInmemNotificationEndpointService(f influxdbtesting.NotificationEndpointFields, t *testing.T) (influxdb.NotificationEndpointService, string, func()) {
	s, closeInmem, err := NewTestInmemStore()
	if err != nil {
		t.Fatalf("failed to create new kv store: %v", err)
	}

	svc, op, closeSvc := initNotificationEndpointService(s, f, t)
	return svc, op, func() {
		closeSvc()
		closeInmem()
	}
}

func initNotificationEndpointService(s kv.Store, f influxdbtesting.NotificationEndpointFields, t *testing.T) (influxdb.NotificationEndpointService, string, func()) {
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

	for _, o := range f.Organizations {
		if err := svc.PutOrganization(ctx, o); err != nil {
			t.Fatalf("failed to populate org: %v", err)
		}
	}

	return svc, kv.OpPrefix, func() {
		for _, nr := range f.NotificationEndpoints {
			if err := svc.DeleteNotificationEndpoint(ctx, nr.GetID()); err != nil {
				t.Logf("failed to remove notification endpoint: %v", err)
			}
		}
		for _, o := range f.Organizations {
			if err := svc.DeleteOrganization(ctx, o.ID); err != nil {
				t.Fatalf("failed to remove org: %v", err)
			}
		}
	}
}
