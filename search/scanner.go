package search

import (
	"context"

	"github.com/influxdata/influxdb"
)

// Scanner is the collection of list service.
type Scanner struct {
	Service FindService
	influxdb.BucketService
}

// Scan different resources in search index.
func (sc Scanner) Scan(ctx context.Context) error {
	if err := sc.scanBuckets(ctx); err != nil {
		return err
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
