package main

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/influxdb/search"

	"github.com/blevesearch/bleve"
	"github.com/influxdata/influxdb"
)

func main() {
	index, err := bleve.Open("example.bleve")
	if err != nil {
		panic(err)
	}

	svc := &search.Service{
		Core: index,
	}

	for i := 1; i <= 50; i++ {
		bucket := influxdb.Bucket{
			ID:                  influxdb.ID(i),
			OrgID:               influxdb.ID(51 + i),
			Type:                influxdb.BucketTypeUser,
			Name:                "tom",
			Description:         "description of bucket 1",
			RetentionPeriod:     1000 * time.Hour,
			RetentionPolicyName: "retention name",
			CRUDLog: influxdb.CRUDLog{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		if i%2 == 0 {
			bucket.Type = influxdb.BucketTypeSystem
		}

		err = svc.Index(context.Background(), search.ConvertBucket(bucket))
		if err != nil {
			panic(err)
		}
	}

	docs, err := svc.SimpleQuery(influxdb.ID(11).String(), "bucket")
	if err != nil {
		panic(err)
	}

	fmt.Println("hits: ", len(docs))
	for _, doc := range docs {
		fmt.Printf("%#v\n", doc)
	}
}
