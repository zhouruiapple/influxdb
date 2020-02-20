package pkger

import (
	"context"

	"github.com/influxdata/influxdb"
)

type StoreCombined struct {
	sql StoreSQL
	kv  StoreKV
}

var _ Store = (*StoreCombined)(nil)

func NewStoreCombined(sqlStore StoreSQL, kvStore StoreKV) *StoreCombined {
	return &StoreCombined{
		sql: sqlStore,
		kv:  kvStore,
	}
}

func (s *StoreCombined) Create(ctx context.Context, orgID influxdb.ID, pkg *Pkg) error {
	panic("not implemented")
}

func (s *StoreCombined) Read(ctx context.Context, orgID influxdb.ID, pkgName string) error {
	panic("not implemented")
}

func (s *StoreCombined) Update(ctx context.Context, orgID influxdb.ID, pkg *Pkg) error {
	panic("not implemented")
}

func (s *StoreCombined) Delete(ctx context.Context, orgID influxdb.ID, pkgName string) error {
	panic("not implemented")
}
