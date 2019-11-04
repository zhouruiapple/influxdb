package meta

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/bolt"
	"github.com/influxdata/influxdb/internal/fs"
	"github.com/influxdata/influxdb/kv"
	"github.com/spf13/cobra"
)

// TODO(edd): future batch API:
// The command also supports a batch generation, where JSON
// objects can be used to bulk generate buckets with retention periods.

// The schema should look like:

// {"org_id": "064d942e6469e221", "bucket_id": "042d986d64622123", "retention_period": "0"}
// {"org_id": "064d942e6469e221", "org": "edd's org", "bucket_id": "033d986d64629292", "bucket": "edd's bucket", "retention_period": "3h"}

// "org_id" is required. "bucket_id" is only required if "retention_period" is
// present. "org" and "bucket" are optional.

var Command = &cobra.Command{
	Use:   "meta",
	Short: "Generate meta data such as organizations and buckets",
	Long: `
This command can be used to generate meta data such as organizations and buckets.

Assuming that you already have a user, org, bucket and token locally created, e.g.,
through influx setup, then you can use this tool to add another org and bucket
that can be written to or read from using your existing token.

Example:

    $ influxd generate meta --org another-org --org_id 042e986e6469e321 --bucket b2 --bucket_id 123f386e6469e444     

This command is only for development use. It should not be run against any 
production servers. influxd should not be running when this tool is used.
`,
	RunE: run,
}

var flags struct {
	org       string
	orgID     string
	bucket    string
	bucketID  string
	retention time.Duration

	boltPath string
}

var kvService *kv.Service

func init() {
	pfs := Command.PersistentFlags()
	pfs.StringVar(&flags.org, "org", "", "optional organization name")
	pfs.StringVar(&flags.orgID, "org_id", "", "required organization ID")
	pfs.StringVar(&flags.bucket, "bucket", "", "optional bucket name. org_id required")
	pfs.StringVar(&flags.bucketID, "bucket_id", "", "bucket ID. org_id required")
	pfs.DurationVar(&flags.retention, "retention", time.Duration(0), "optional retention period for bucket. org_id and bucket_id required")

	bolt, err := fs.BoltFile()
	if err != nil {
		fmt.Printf("Error initializing bolt file path: %v", err)
		os.Exit(1)
	}
	pfs.StringVar(&flags.boltPath, "bolt-path", bolt, "path to the Bolt database file")
}

func verifyFlags() error {
	if flags.orgID == "" {
		return errors.New("a base-16 formatted org_id is required")
	} else if flags.retention != time.Duration(0) && flags.bucketID == "" {
		return errors.New("a base-16 formatted bucket_id is required if setting a retention duration")
	}
	return nil
}

func run(_ *cobra.Command, args []string) error {
	ctx := context.Background()

	if err := verifyFlags(); err != nil {
		return err
	}

	store := bolt.NewKVStore(flags.boltPath)
	if err := store.Open(ctx); err != nil {
		return err
	}
	kvService = kv.NewService(store)

	o, err := createOrg(ctx, flags.orgID, flags.org)
	if err != nil {
		return err
	}

	_, err = createBucket(ctx, o, flags.bucketID, flags.bucket)
	if err != nil {
		return err
	}

	return nil
}

func createOrg(ctx context.Context, id, name string) (*influxdb.Organization, error) {
	oid, err := influxdb.IDFromString(id)
	if err != nil {
		return nil, err
	}

	o := &influxdb.Organization{
		Name: name,
		ID:   *oid,
	}

	if err := kvService.PutOrganization(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func createBucket(ctx context.Context, o *influxdb.Organization, id, name string) (*influxdb.Bucket, error) {
	bid, err := influxdb.IDFromString(id)
	if err != nil {
		return nil, err
	}

	b := &influxdb.Bucket{
		Name:  name,
		ID:    *bid,
		OrgID: o.ID,
	}

	if err := kvService.PutBucket(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}
