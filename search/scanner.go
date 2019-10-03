package search

import (
	"context"

	"github.com/influxdata/influxdb"
)

// Scanner is the collection of list service.
type Scanner struct {
	Service             FindService
	BucketService       influxdb.BucketService
	OrganizationService influxdb.OrganizationService
	UserService         influxdb.UserService
}

// Scan different resources in search index.
func (sc Scanner) Scan(ctx context.Context) error {
	scanFns := []func(context.Context) error{
		sc.scanBuckets,
		sc.scanOrganizations,
		sc.scanUsers,
	}

	for _, scanFn := range scanFns {
		if err := scanFn(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (sc Scanner) scanBuckets(ctx context.Context) error {
	bkts, _, err := sc.BucketService.FindBuckets(ctx, influxdb.BucketFilter{})
	if err != nil {
		return err
	}
	for _, bkt := range bkts {
		b := ConvertBucket(*bkt)
		if err := sc.Service.Index(ctx, b); err != nil {
			return err
		}
	}
	return nil
}

func (sc Scanner) scanOrganizations(ctx context.Context) error {
	orgs, _, err := sc.OrganizationService.FindOrganizations(ctx, influxdb.OrganizationFilter{})
	if err != nil {
		return err
	}

	for _, o := range orgs {
		org := ConvertOrganization(*o)
		if err := sc.Service.Index(ctx, org); err != nil {
			return err
		}
	}
	return nil
}

func (sc Scanner) scanUsers(ctx context.Context) error {
	usrs, _, err := sc.UserService.FindUsers(ctx, influxdb.UserFilter{})
	if err != nil {
		return &influxdb.Error{
			Msg: "err scanning user",
			Err: err,
		}
	}
	for _, usr := range usrs {
		u := ConvertUser(*usr)
		if err := sc.Service.Index(ctx, u); err != nil {
			return &influxdb.Error{
				Msg: "err converting user",
				Err: err,
			}
		}
	}
	return nil
}
