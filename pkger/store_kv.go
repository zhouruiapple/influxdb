package pkger

import (
	"context"

	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/kv"
)

type StoreKV struct {
	store kv.Store
}

var _ Store = (*StoreKV)(nil)

func NewStoreKV(store kv.Store) *StoreKV {
	return &StoreKV{
		store: store,
	}
}

func (s *StoreKV) Create(ctx context.Context, orgID influxdb.ID, pkg *Pkg) error {
	panic("not implemented")
}

func (s *StoreKV) Read(ctx context.Context, orgID influxdb.ID, pkgName string) error {
	panic("not implemented")
}

func (s *StoreKV) Update(ctx context.Context, orgID influxdb.ID, pkg *Pkg) error {
	panic("not implemented")
}

func (s *StoreKV) Delete(ctx context.Context, orgID influxdb.ID, pkgName string) error {
	panic("not implemented")
}
