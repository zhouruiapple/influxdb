package readback_test

// 1. Compile as a standalone:
//
// go test -c query/stdlib/readback
//
// 2. Run until it fails:
//
// while true; do ./readback.test; done
//
//    OR if you are less patient:
//
// bash ./query/stdlib/readback/load.sh

import (
	"fmt"
	"bytes"
	"context"
	"strings"
	"testing"

	"encoding/binary"
	"github.com/influxdata/influxdb/v2"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib"

	platform "github.com/influxdata/influxdb/v2"
	"github.com/influxdata/influxdb/v2/cmd/influxd/launcher"
	influxdbcontext "github.com/influxdata/influxdb/v2/context"
	"github.com/influxdata/influxdb/v2/mock"
	"github.com/influxdata/influxdb/v2/query"
	"github.com/influxdata/influxdb/v2/http"
	_ "github.com/influxdata/influxdb/v2/query/stdlib"
	"github.com/influxdata/influxdb/v2/query/stdlib/readback"
	"github.com/influxdata/influxdb/v2/kit/feature"
)

// Default context.
var ctx = influxdbcontext.SetAuthorizer(context.Background(), mock.NewMockAuthorizer(true, nil))

func init() {
	runtime.FinalizeBuiltIns()
}

type variableAssignmentVisitor struct {
	fn func(*ast.VariableAssignment)
}

func (v variableAssignmentVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.VariableAssignment:
		v.fn(n)
		return nil
	}
	return v
}

func (v variableAssignmentVisitor) Done(node ast.Node) {}

func makeFluxTest(t testing.TB, file *ast.File ) {

	if t.Name() == "TestFluxEndToEnd/difference_panic" {
		return
	}
	if t.Name() == "TestFluxEndToEnd/merge_filter_flag_on" {
		return
	}
	if t.Name() == "TestFluxEndToEnd/merge_filter_flag_off" {
		return
	}
	if t.Name() == "TestFluxEndToEnd/join_use_previous" {
		return
	}
	if t.Name() == "TestFluxEndToEnd/fill_previous" {
		return
	}
	if t.Name() == "TestFluxEndToEnd/table_fns" {
		return
	}
	if t.Name() == "TestFluxEndToEnd/to_float" {
		return
	}

	found := false
	inData := ""
	visitor := variableAssignmentVisitor{
		fn: func(va *ast.VariableAssignment) {
			if !found && va.ID.Name == "inData" {
				inData = "inData = " + string(ast.Format(va.Init)) + "\n"
				found = true
			}
		},
	}

	ast.Walk(visitor, file)

	if found {
		fmt.Println( "[]string {" )
		fmt.Println( fmt.Sprintf("\"%s\",", t.Name() ) )
	    fmt.Println( "`", inData, "`," )
		fmt.Println( "}," )
	}
}

func makeTestCases(t *testing.T, pkgs []*ast.Package) {
	fmt.Println( "package readback" )
	fmt.Println( "var Cases [][]string = [][]string {" )

	for _, pkg := range pkgs {
		test := func(t *testing.T, f func(t *testing.T)) {
			t.Run(pkg.Path, f)
		}
		if pkg.Path == "universe" {
			test = func(t *testing.T, f func(t *testing.T)) {
				f(t)
			}
		}

		test(t, func(t *testing.T) {
			for _, file := range pkg.Files {
				name := strings.TrimSuffix(file.Name, "_test.flux")
				t.Run(name, func(t *testing.T) {
					if reason, ok := readback.FluxEndToEndSkipList[pkg.Path][name]; ok {
						t.Skip(reason)
					}

					makeFluxTest(t, file)
				})
			}
		})
	}

	// Little hack to ignore the PASS printed by the test. Could just make a command, but meh.
	fmt.Println( "}" )
	fmt.Print( "//" )
}

func checkID(id influxdb.ID, check byte) bool {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	for i := range b {
		if b[i] == check {
			return false
		}
	}
	return true
}

func testFluxWrite(t testing.TB, l *launcher.TestLauncher, name string, write string, b *platform.Bucket ) {

	//fmt.Println("test case A: ", name )

	req := &query.Request{
		OrganizationID: l.Org.ID,
		Compiler:       lang.FluxCompiler{
			Query: write,
		},
	}

	if r, err := l.FluxQueryService().Query(ctx, req); err != nil {
		t.Fatal(err)
	} else {
		results := make( map[string]*bytes.Buffer )

		for r.More() {
			v := r.Next()

			//fmt.Println("e2e results: ", v.Name())
			if _, ok := results[v.Name()]; !ok {
				results[v.Name()] = &bytes.Buffer{}
			}
			err := execute.FormatResult(results[v.Name()], v)
			if err != nil {
				t.Error(err)
			}
		}
		if err := r.Err(); err != nil {
			t.Error(err)
		}
	}
}

func testFluxRead(t testing.TB, l *launcher.TestLauncher, name string, read string, b *platform.Bucket ) bool {
	pass := false

	req := &query.Request{
		OrganizationID: l.Org.ID,
		Compiler:       lang.FluxCompiler{Query: read},
	}

	if r, err := l.FluxQueryService().Query(ctx, req); err != nil {
		t.Fatal(err)
	} else {
		results := make( map[string]*bytes.Buffer )

		for r.More() {
			v := r.Next()

			if _, ok := results[v.Name()]; !ok {
				results[v.Name()] = &bytes.Buffer{}
			}
			err := execute.FormatResult(results[v.Name()], v)
			if err != nil {
				t.Error(err)
			}
		}
		if err := r.Err(); err != nil {
			t.Error(err)
		}

		if _, ok := results["readback"]; ok {
			pass = true
		} else {
			t.Error("readback was empty")
		}
	}
	return pass
}

func testFlux(t testing.TB, l *launcher.TestLauncher, bs *http.BucketService, name string, inData string ) {
	b := &platform.Bucket{
		OrgID:           l.Org.ID,
		Name:            name,
		RetentionPeriod: 0,
	}

	println( fmt.Sprint( "test name: ", name ) )

	// fmt.Println("creating bucket", name)
	if err := bs.CreateBucket(ctx, b); err != nil {
		t.Fatal(err)
	}

	found, err := bs.FindBucketByName( ctx, l.Org.ID, name )
	if err != nil || found == nil {
		fmt.Println( "immediate bucket find failed" )
		t.Error("immediate bucket find failed")
	}

	bucket := name
	org := l.Org.Name

	write :=
		"import \"csv\"" +
		inData +
		"csv.from( csv: inData ) |> to( bucket: \"" + bucket + "\", org: \"" + org + "\" )\n"

	read :=
		"from( bucket: \"" + bucket + "\" ) |> range( start: 0 ) |> yield( name: \"readback\" )\n"

	testFluxWrite(t, l, name, write, b)
	pass := testFluxRead(t, l, name, read, b)

	if !pass {
		fmt.Println( "readback was empty:", b.ID.String() )
	}

	if pass && !checkID( b.ID, 0x20 ) {
		fmt.Println( "test passed but checkid 20 failed:", b.ID.String() )
	}

	if pass && !checkID( b.ID, 0x2c ) {
		fmt.Println( "test passed but checkid 2c failed:", b.ID.String() )
	}

	if pass && !checkID( b.ID, 0x5c ) {
		fmt.Println( "test passed but checkid 5c failed:", b.ID.String() )
	}

	if !pass && ( checkID( b.ID, 0x20 ) && checkID( b.ID, 0x2c ) && checkID( b.ID, 0x5c ) ) {
		fmt.Println( "test failed but checkid passed:", b.ID.String() )
	}
}
func runReadBack(t *testing.T, pkgs []*ast.Package) {
	flagger := feature.DefaultFlagger()
	l := launcher.RunTestLauncherOrFail(t, ctx, flagger)
	l.SetupOrFail(t)
	bs := l.BucketService(t)
	defer l.ShutdownOrFail(t, ctx)

	for i, _ := range readback.Cases {
		testFlux(t, l, bs, readback.Cases[i][0], readback.Cases[i][1])
	}
}

func TestFluxEndToEnd(t *testing.T) {
	runReadBack(t, stdlib.FluxTestPackages)
}
