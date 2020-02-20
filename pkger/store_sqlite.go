package pkger

import (
	"context"
	"database/sql"

	"github.com/influxdata/influxdb"
)

type StoreSQL struct {
	db *sql.DB
}

var _ Store = (*StoreSQL)(nil)

func NewStoreSQLite(db *sql.DB) *StoreSQL {
	return &StoreSQL{
		db: db,
	}
}

func (s *StoreSQL) Create(ctx context.Context, orgID influxdb.ID, pkg *Pkg) error {
	panic("not implemented")
}

func (s *StoreSQL) Read(ctx context.Context, orgID influxdb.ID, pkgName string) error {
	panic("not implemented")
}

func (s *StoreSQL) Update(ctx context.Context, orgID influxdb.ID, pkg *Pkg) error {
	panic("not implemented")
}

func (s *StoreSQL) Delete(ctx context.Context, orgID influxdb.ID, pkgName string) error {
	panic("not implemented")
}
