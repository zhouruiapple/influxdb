package launcher_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	nethttp "net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/cmd/influxd/launcher"
	phttp "github.com/influxdata/influxdb/http"
	"github.com/influxdata/influxdb/query"
	_ "github.com/influxdata/influxdb/query/builtin"
)

func TestPipeline_Write_Query_FieldKey(t *testing.T) {
	be := launcher.RunTestLauncherOrFail(t, ctx)
	be.SetupOrFail(t)
	defer be.ShutdownOrFail(t, ctx)

	resp, err := nethttp.DefaultClient.Do(
		be.MustNewHTTPRequest(
			"POST",
			fmt.Sprintf("/api/v2/write?org=%s&bucket=%s", be.Org.ID, be.Bucket.ID),
			`cpu,region=west,server=a v0=1.2
cpu,region=west,server=b v0=33.2
cpu,region=east,server=b,area=z v1=100.0
disk,regions=north,server=b v1=101.2
mem,server=b value=45.2`))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Error(err)
		}
	}()
	if resp.StatusCode != 204 {
		t.Fatal("failed call to write points")
	}

	rawQ := fmt.Sprintf(`from(bucket:"%s")
	|> range(start:-1m)
	|> filter(fn: (r) => r._measurement ==  "cpu" and (r._field == "v1" or r._field == "v0"))
	|> group(columns:["_time", "_value"], mode:"except")
	`, be.Bucket.Name)

	// Expected keys:
	//
	// _measurement=cpu,region=west,server=a,_field=v0
	// _measurement=cpu,region=west,server=b,_field=v0
	// _measurement=cpu,region=east,server=b,area=z,_field=v1
	//
	results := be.MustExecuteQuery(rawQ)
	defer results.Done()
	results.First(t).HasTablesWithCols([]int{4, 4, 5})
}

// This test initialises a default launcher writes some data,
// and checks that the queried results contain the expected number of tables
// and expected number of columns.
func TestPipeline_WriteV2_Query(t *testing.T) {
	be := launcher.RunTestLauncherOrFail(t, ctx)
	be.SetupOrFail(t)
	defer be.ShutdownOrFail(t, ctx)

	// The default gateway instance inserts some values directly such that ID lookups seem to break,
	// so go the roundabout way to insert things correctly.
	req := be.MustNewHTTPRequest(
		"POST",
		fmt.Sprintf("/api/v2/write?org=%s&bucket=%s", be.Org.ID, be.Bucket.ID),
		fmt.Sprintf("ctr n=1i %d", time.Now().UnixNano()),
	)
	phttp.SetToken(be.Auth.Token, req)

	resp, err := nethttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Error(err)
		}
	}()

	if resp.StatusCode != nethttp.StatusNoContent {
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, resp.Body); err != nil {
			t.Fatalf("Could not read body: %s", err)
		}
		t.Fatalf("exp status %d; got %d, body: %s", nethttp.StatusNoContent, resp.StatusCode, buf.String())
	}

	res := be.MustExecuteQuery(fmt.Sprintf(`from(bucket:"%s") |> range(start:-5m)`, be.Bucket.Name))
	defer res.Done()
	res.HasTableCount(t, 1)
}

// This test initializes a default launcher; writes some data; queries the data (success);
// sets memory limits to the same read query; checks that the query fails because limits are exceeded.
func TestPipeline_QueryMemoryLimits(t *testing.T) {
	t.Skip("setting memory limits in the client is not implemented yet")

	l := launcher.RunTestLauncherOrFail(t, ctx)
	l.SetupOrFail(t)
	defer l.ShutdownOrFail(t, ctx)

	// write some points
	for i := 0; i < 100; i++ {
		l.WritePointsOrFail(t, fmt.Sprintf(`m,k=v1 f=%di %d`, i*100, time.Now().UnixNano()))
	}

	// compile a from query and get the spec
	qs := fmt.Sprintf(`from(bucket:"%s") |> range(start:-5m)`, l.Bucket.Name)
	pkg, err := flux.Parse(qs)
	if err != nil {
		t.Fatal(err)
	}

	// we expect this request to succeed
	req := &query.Request{
		Authorization:  l.Auth,
		OrganizationID: l.Org.ID,
		Compiler: lang.ASTCompiler{
			AST: pkg,
		},
	}
	if err := l.QueryAndNopConsume(context.Background(), req); err != nil {
		t.Fatal(err)
	}

	// ok, the first request went well, let's add memory limits:
	// this query should error.
	// spec.Resources = flux.ResourceManagement{
	// 	MemoryBytesQuota: 100,
	// }

	if err := l.QueryAndNopConsume(context.Background(), req); err != nil {
		if !strings.Contains(err.Error(), "allocation limit reached") {
			t.Fatalf("query errored with unexpected error: %v", err)
		}
	} else {
		t.Fatal("expected error, got successful query execution")
	}
}

func TestPipeline_Query_Auth_Buckets(t *testing.T) {
	l := launcher.RunTestLauncherOrFail(t, ctx)
	l.SetupOrFail(t)
	defer l.ShutdownOrFail(t, ctx)

	// create a few buckets.
	want := []string{l.Bucket.Name}
	for i := 0; i < 3; i++ {
		b := &influxdb.Bucket{
			OrgID: l.Org.ID,
			Name:  fmt.Sprintf("testbucket%d", i),
		}
		if err := l.BucketService().CreateBucket(ctx, b); err != nil {
			t.Fatalf("unexpected create bucket error: %s", err)
		}
		want = append(want, b.Name)
	}
	sort.Strings(want)

	// construct the request we will continue to use for the tests.
	req := &query.Request{
		Authorization:  l.Auth,
		OrganizationID: l.Org.ID,
		Compiler: lang.FluxCompiler{
			Query: `
buckets()
`,
		},
	}

	getBucketNames := func(buckets *[]string) func(res flux.Result) error {
		return func(res flux.Result) error {
			start := len(*buckets)
			if err := res.Tables().Do(func(tbl flux.Table) error {
				return tbl.Do(func(cr flux.ColReader) error {
					j := execute.ColIdx("name", cr.Cols())
					if j == -1 {
						return errors.New("no column named \"name\"")
					}

					vs := cr.Strings(j)
					for i := 0; i < vs.Len(); i++ {
						*buckets = append(*buckets, vs.ValueString(i))
					}
					return nil
				})
			}); err != nil {
				return err
			}
			sort.Strings((*buckets)[start:])
			return nil
		}
	}

	// All permissions.
	t.Run("All", func(t *testing.T) {
		var got []string
		if err := l.QueryAndConsume(ctx, req, getBucketNames(&got)); err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		if !cmp.Equal(want, got) {
			t.Errorf("unexpected bucket names -want/+got:\n%s", cmp.Diff(want, got))
		}
	})

	t.Run("Partial", func(t *testing.T) {
		l := *l

		auth := &influxdb.Authorization{
			OrgID:  l.Org.ID,
			UserID: l.User.ID,
			Permissions: []influxdb.Permission{
				{
					Action: influxdb.ReadAction,
					Resource: influxdb.Resource{
						Type:  influxdb.BucketsResourceType,
						ID:    &l.Bucket.ID,
						OrgID: &l.Org.ID,
					},
				},
			},
		}
		if err := l.AuthorizationService().CreateAuthorization(ctx, auth); err != nil {
			t.Fatalf("unexpected error creating authorization: %s", err)
		}
		l.Auth = auth

		var got []string
		if err := l.QueryAndConsume(ctx, req, getBucketNames(&got)); err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		want := []string{l.Bucket.Name}
		if !cmp.Equal(want, got) {
			t.Errorf("unexpected bucket names -want/+got:\n%s", cmp.Diff(want, got))
		}
	})

	t.Run("None", func(t *testing.T) {
		l := *l

		auth := &influxdb.Authorization{
			OrgID:  l.Org.ID,
			UserID: l.User.ID,
			Permissions: []influxdb.Permission{
				{
					// No read permission on any bucket.
					Action: influxdb.WriteAction,
					Resource: influxdb.Resource{
						Type:  influxdb.BucketsResourceType,
						ID:    &l.Bucket.ID,
						OrgID: &l.Org.ID,
					},
				},
			},
		}
		if err := l.AuthorizationService().CreateAuthorization(ctx, auth); err != nil {
			t.Fatalf("unexpected error creating authorization: %s", err)
		}
		l.Auth = auth

		if err := l.QueryAndNopConsume(ctx, req); err == nil {
			t.Error("expected error")
		} else if got, want := influxdb.ErrorCode(err), influxdb.ENotFound; got != want {
			t.Errorf("unexpected error code -want/+got:\n\t- %v\n\t+ %v", got, want)
		}
	})
}

func TestPipeline_Query_Auth_From(t *testing.T) {
	l := launcher.RunTestLauncherOrFail(t, ctx)
	l.SetupOrFail(t)
	defer l.ShutdownOrFail(t, ctx)

	// write one point so we can read something.
	l.WritePointsOrFail(t, fmt.Sprintf(`m,k=v1 f=%di %d`, 0, time.Now().UnixNano()))

	t.Run("Success_Bucket", func(t *testing.T) {
		// read from the bucket.
		req := &query.Request{
			Authorization:  l.Auth,
			OrganizationID: l.Org.ID,
			Compiler: lang.FluxCompiler{
				Query: fmt.Sprintf(`
from(bucket: "%s")
	|> range(start: -5m)
`, l.Bucket.Name),
			},
		}

		if err := l.QueryAndNopConsume(ctx, req); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("Success_BucketID", func(t *testing.T) {
		// read from the bucket.
		req := &query.Request{
			Authorization:  l.Auth,
			OrganizationID: l.Org.ID,
			Compiler: lang.FluxCompiler{
				Query: fmt.Sprintf(`
from(bucketID: "%s")
	|> range(start: -5m)
`, l.Bucket.ID),
			},
		}

		if err := l.QueryAndNopConsume(ctx, req); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	auth := &influxdb.Authorization{
		OrgID:  l.Org.ID,
		UserID: l.User.ID,
		Permissions: []influxdb.Permission{
			{
				Action: influxdb.WriteAction,
				Resource: influxdb.Resource{
					Type:  influxdb.BucketsResourceType,
					ID:    &l.Bucket.ID,
					OrgID: &l.Org.ID,
				},
			},
		},
	}
	if err := l.AuthorizationService().CreateAuthorization(ctx, auth); err != nil {
		t.Fatalf("unexpected error creating authorization: %s", err)
	}
	l.Auth = auth

	t.Run("Forbidden_Bucket", func(t *testing.T) {
		l := *l

		// read from the bucket.
		req := &query.Request{
			Authorization:  l.Auth,
			OrganizationID: l.Org.ID,
			Compiler: lang.FluxCompiler{
				Query: fmt.Sprintf(`
from(bucket: "%s")
	|> range(start: -5m)
`, l.Bucket.Name),
			},
		}

		if err := l.QueryAndNopConsume(ctx, req); err == nil {
			t.Fatal("expected error")
		} else if got, want := influxdb.ErrorCode(err), influxdb.ENotFound; got != want {
			t.Errorf("unexpected error code -want/+got:\n\t- %v\n\t+ %v", got, want)
		}
	})

	t.Run("Forbidden_BucketID", func(t *testing.T) {
		l := *l

		// read from the bucket.
		req := &query.Request{
			Authorization:  l.Auth,
			OrganizationID: l.Org.ID,
			Compiler: lang.FluxCompiler{
				Query: fmt.Sprintf(`
from(bucketID: "%s")
	|> range(start: -5m)
`, l.Bucket.ID),
			},
		}

		if err := l.QueryAndNopConsume(ctx, req); err == nil {
			t.Fatal("expected error")
		} else if got, want := influxdb.ErrorCode(err), influxdb.ENotFound; got != want {
			t.Errorf("unexpected error code -want/+got:\n\t- %v\n\t+ %v", got, want)
		}
	})
}

func TestPipeline_Query_Auth_WritePoints(t *testing.T) {
	l := launcher.RunTestLauncherOrFail(t, ctx)
	l.SetupOrFail(t)
	defer l.ShutdownOrFail(t, ctx)

	// write one point so we can read something.
	l.WritePointsOrFail(t, fmt.Sprintf(`m,k=v1 f=%di %d`, 5, time.Now().UnixNano()))

	t.Run("Success_Bucket", func(t *testing.T) {
		req := &query.Request{
			Authorization:  l.Auth,
			OrganizationID: l.Org.ID,
			Compiler: lang.FluxCompiler{
				Query: fmt.Sprintf(`
from(bucket: "%s")
	|> range(start: -5m)
	|> filter(fn: (r) => r._measurement == "m" and r._field == "f")
	|> map(fn: (r) => ({r with _measurement: "w", _value: r._value * 2}))
	|> to(bucket: "%s")
`, l.Bucket.Name, l.Bucket.Name),
			},
		}

		if err := l.QueryAndNopConsume(ctx, req); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("Success_BucketID", func(t *testing.T) {
		req := &query.Request{
			Authorization:  l.Auth,
			OrganizationID: l.Org.ID,
			Compiler: lang.FluxCompiler{
				Query: fmt.Sprintf(`
from(bucket: "%s")
	|> range(start: -5m)
	|> filter(fn: (r) => r._measurement == "m" and r._field == "f")
	|> map(fn: (r) => ({r with _measurement: "w", _value: r._value * 2}))
	|> to(bucketID: "%s")
`, l.Bucket.Name, l.Bucket.ID),
			},
		}

		if err := l.QueryAndNopConsume(ctx, req); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	auth := &influxdb.Authorization{
		OrgID:  l.Org.ID,
		UserID: l.User.ID,
		Permissions: []influxdb.Permission{
			{
				// No write permission on the bucket.
				Action: influxdb.ReadAction,
				Resource: influxdb.Resource{
					Type:  influxdb.BucketsResourceType,
					ID:    &l.Bucket.ID,
					OrgID: &l.Org.ID,
				},
			},
		},
	}
	if err := l.AuthorizationService().CreateAuthorization(ctx, auth); err != nil {
		t.Fatalf("unexpected error creating authorization: %s", err)
	}
	l.Auth = auth

	t.Run("Forbidden_Bucket", func(t *testing.T) {
		l := *l

		req := &query.Request{
			Authorization:  l.Auth,
			OrganizationID: l.Org.ID,
			Compiler: lang.FluxCompiler{
				Query: fmt.Sprintf(`
from(bucket: "%s")
	|> range(start: -5m)
	|> filter(fn: (r) => r._measurement == "m" and r._field == "f")
	|> map(fn: (r) => ({r with _measurement: "forbidden", _value: r._value * 2}))
	|> to(bucket: "%s")
`, l.Bucket.Name, l.Bucket.Name),
			},
		}

		if err := l.QueryAndNopConsume(ctx, req); err == nil {
			t.Fatal("expected error")
		} else if got, want := influxdb.ErrorCode(err), influxdb.EUnauthorized; got != want {
			t.Errorf("unexpected error code -want/+got:\n\t- %v\n\t+ %v", got, want)
		}
	})

	t.Run("Forbidden_BucketID", func(t *testing.T) {
		l := *l

		req := &query.Request{
			Authorization:  l.Auth,
			OrganizationID: l.Org.ID,
			Compiler: lang.FluxCompiler{
				Query: fmt.Sprintf(`
from(bucket: "%s")
	|> range(start: -5m)
	|> filter(fn: (r) => r._measurement == "m" and r._field == "f")
	|> map(fn: (r) => ({r with _measurement: "forbidden", _value: r._value * 2}))
	|> to(bucketID: "%s")
`, l.Bucket.Name, l.Bucket.ID),
			},
		}

		if err := l.QueryAndNopConsume(ctx, req); err == nil {
			t.Fatal("expected error")
		} else if got, want := influxdb.ErrorCode(err), influxdb.EUnauthorized; got != want {
			t.Errorf("unexpected error code -want/+got:\n\t- %v\n\t+ %v", got, want)
		}
	})
}

func TestPipeline_Query_LoadSecret_Success(t *testing.T) {
	l := launcher.RunTestLauncherOrFail(t, ctx)
	l.SetupOrFail(t)
	defer l.ShutdownOrFail(t, ctx)

	const key, value = "mytoken", "secrettoken"
	if err := l.SecretService().PutSecret(ctx, l.Org.ID, key, value); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// write one point so we can use it
	l.WritePointsOrFail(t, fmt.Sprintf(`m,k=v1 f=%di %d`, 0, time.Now().UnixNano()))

	// we expect this request to succeed
	req := &query.Request{
		Authorization:  l.Auth,
		OrganizationID: l.Org.ID,
		Compiler: lang.FluxCompiler{
			Query: fmt.Sprintf(`
import "influxdata/influxdb/secrets"

token = secrets.get(key: "mytoken")
from(bucket: "%s")
	|> range(start: -5m)
	|> set(key: "token", value: token)
`, l.Bucket.Name),
		},
	}
	if err := l.QueryAndConsume(ctx, req, func(r flux.Result) error {
		return r.Tables().Do(func(tbl flux.Table) error {
			return tbl.Do(func(cr flux.ColReader) error {
				j := execute.ColIdx("token", cr.Cols())
				if j == -1 {
					return errors.New("cannot find table column \"token\"")
				}

				for i := 0; i < cr.Len(); i++ {
					v := execute.ValueForRow(cr, i, j)
					if got, want := v, values.NewString("secrettoken"); !got.Equal(want) {
						t.Errorf("unexpected value at row %d -want/+got:\n\t- %v\n\t+ %v", i, got, want)
					}
				}
				return nil
			})
		})
	}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestPipeline_Query_LoadSecret_Forbidden(t *testing.T) {
	l := launcher.RunTestLauncherOrFail(t, ctx)
	l.SetupOrFail(t)
	defer l.ShutdownOrFail(t, ctx)

	const key, value = "mytoken", "secrettoken"
	if err := l.SecretService().PutSecret(ctx, l.Org.ID, key, value); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// write one point so we can use it
	l.WritePointsOrFail(t, fmt.Sprintf(`m,k=v1 f=%di %d`, 0, time.Now().UnixNano()))

	auth := &influxdb.Authorization{
		OrgID:  l.Org.ID,
		UserID: l.User.ID,
		Permissions: []influxdb.Permission{
			{
				Action: influxdb.ReadAction,
				Resource: influxdb.Resource{
					Type:  influxdb.BucketsResourceType,
					ID:    &l.Bucket.ID,
					OrgID: &l.Org.ID,
				},
			},
		},
	}
	if err := l.AuthorizationService().CreateAuthorization(ctx, auth); err != nil {
		t.Fatalf("unexpected error creating authorization: %s", err)
	}
	l.Auth = auth

	// we expect this request to succeed
	req := &query.Request{
		Authorization:  l.Auth,
		OrganizationID: l.Org.ID,
		Compiler: lang.FluxCompiler{
			Query: fmt.Sprintf(`
import "influxdata/influxdb/secrets"

token = secrets.get(key: "mytoken")
from(bucket: "%s")
	|> range(start: -5m)
	|> set(key: "token", value: token)
`, l.Bucket.Name),
		},
	}
	if err := l.QueryAndNopConsume(ctx, req); err == nil {
		t.Error("expected error")
	} else if got, want := influxdb.ErrorCode(err), influxdb.EUnauthorized; got != want {
		t.Errorf("unexpected error code -want/+got:\n\t- %v\n\t+ %v", got, want)
	}
}
