package readback_test

// 1. Compile as a standalone:
//
// go test -c query/stdlib/readback
//
// 2. Run many in parallel:
//
// # How many test case processes to run in parallel.
// PARALLEL=16
// 
// # How many iterations of of parallel runs to make.
// ITERATIONS=16
// 
// logfn() {
// 	printf "log-%02d-%02d.txt" $1 $2
// }
// 
// for run in `seq 1 $ITERATIONS`; do
// 		for i in `seq 1 $PARALLEL`; do
// 				LOG=`logfn $run $i`
// 				./readback.test >$LOG &
// 				sleep 0.5
// 		done
// 
// 		for i in `seq 1 $PARALLEL`; do
// 				wait
// 		done
// 
// 		for i in `seq 1 $PARALLEL`; do
// 			LOG=`logfn $run $i`
// 			grep -q ^FAIL $LOG || rm $LOG 
// 		done
// done

import (
	"fmt"
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib"

	platform "github.com/influxdata/influxdb/v2"
	"github.com/influxdata/influxdb/v2/cmd/influxd/launcher"
	influxdbcontext "github.com/influxdata/influxdb/v2/context"
	"github.com/influxdata/influxdb/v2/mock"
	"github.com/influxdata/influxdb/v2/query"
	"github.com/influxdata/influxdb/v2/http"
	_ "github.com/influxdata/influxdb/v2/query/stdlib"
	itesting "github.com/influxdata/influxdb/v2/query/stdlib/readback"
	"github.com/influxdata/influxdb/v2/kit/feature"
)

func genCalls(pkg *ast.Package, fn string) *ast.File {
	callFile := new(ast.File)
	callFile.Imports = []*ast.ImportDeclaration{{
		Path: &ast.StringLiteral{Value: "testing"},
	}}
	visitor := testStmtVisitor{
		fn: func(tc *ast.TestStatement) {
			callFile.Body = append(callFile.Body, &ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{
                         Name:     fn,
                    },
					Arguments: []ast.Expression{
						&ast.ObjectExpression{
							Properties: []*ast.Property{{
								Key:   &ast.Identifier{Name: "case"},
								Value: tc.Assignment.ID,
							}},
						},
					},
				},
			})
		},
	}
	ast.Walk(visitor, pkg)
	return callFile
}

type testStmtVisitor struct {
	fn func(*ast.TestStatement)
}

func (v testStmtVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.TestStatement:
		v.fn(n)
		return nil
	}
	return v
}

func (v testStmtVisitor) Done(node ast.Node) {}

// TestingInspectCalls constructs an ast.File that calls testing.inspect for each test case within the package.
func TestingTcCall(pkg *ast.Package) *ast.File {
	return genCalls(pkg, "tc_call")
}

// Default context.
var ctx = influxdbcontext.SetAuthorizer(context.Background(), mock.NewMockAuthorizer(true, nil))

func init() {
	runtime.FinalizeBuiltIns()
}

func TestFluxEndToEnd(t *testing.T) {
	runEndToEnd(t, stdlib.FluxTestPackages)
}

func runEndToEnd(t *testing.T, pkgs []*ast.Package) {
	flagger := feature.DefaultFlagger()
	l := launcher.RunTestLauncherOrFail(t, ctx, flagger)
	l.SetupOrFail(t)
	bs := l.BucketService(t)
	defer l.ShutdownOrFail(t, ctx)
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
					if reason, ok := itesting.FluxEndToEndSkipList[pkg.Path][name]; ok {
						t.Skip(reason)
					}

					testFlux(t, l, file, bs)
				})
			}
		})
	}
}

func makeTestPackage(file *ast.File) *ast.Package {
	file = file.Copy().(*ast.File)
	file.Package.Name.Name = "main"
	pkg := &ast.Package{
		Package: "main",
		Files:   []*ast.File{file},
	}
	return pkg
}

var optionsSource1 = `
import "testing"
import c "csv"
import "experimental"

// Options bucket and org are defined dynamically per test

option testing.loadStorage = (csv) => {
	return c.from(csv: csv) |> to(bucket: bucket, org: org)
}

tc_call = (case) => {
    tc = case()
    return tc.input
}

`

var optionsSource2 = `
import "testing"
import c "csv"
import "experimental"

// Options bucket and org are defined dynamically per test

option testing.loadStorage = (csv) => {
	return from(bucket:bucket)
}

tc_call = (case) => {
    tc = case()
    return tc.input |> tc.fn() |> yield(name: "fresh")
}
`

func testFluxWrite(t testing.TB, l *launcher.TestLauncher, file *ast.File, b *platform.Bucket ) {

	//fmt.Println("test case A: ", t.Name())

	optionsPkg := parser.ParseSource(optionsSource1)
	if ast.Check(optionsPkg) > 0 {
		panic(ast.GetError(optionsPkg))
	}
	optionsAST := optionsPkg.Files[0]

	// Define bucket and org options
	bucketOpt := &ast.OptionStatement{
		Assignment: &ast.VariableAssignment{
			ID:   &ast.Identifier{Name: "bucket"},
			Init: &ast.StringLiteral{Value: b.Name},
		},
	}
	orgOpt := &ast.OptionStatement{
		Assignment: &ast.VariableAssignment{
			ID:   &ast.Identifier{Name: "org"},
			Init: &ast.StringLiteral{Value: l.Org.Name},
		},
	}
	options := optionsAST.Copy().(*ast.File)
	options.Body = append([]ast.Statement{bucketOpt, orgOpt}, options.Body...)

	// Add options to pkg
	pkg := makeTestPackage(file)
	pkg.Files = append(pkg.Files, options)

	// Use testing.inspect call to get all of diff, want, and got
	inspectCalls := TestingTcCall(pkg)
	pkg.Files = append(pkg.Files, inspectCalls)

	bs, err := json.Marshal(pkg)
	if err != nil {
		t.Fatal(err)
	}

	req := &query.Request{
		OrganizationID: l.Org.ID,
		Compiler:       lang.ASTCompiler{AST: bs},
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

//		logFormatted := func( name string, results map[string]*bytes.Buffer ) {
//			if _, ok := results[name]; ok {
//				scanner := bufio.NewScanner(results[name])
//				for scanner.Scan() {
//					t.Log( scanner.Text())
//				}
//			} else {
//				t.Log( "table ", name, " not present in results" )
//			}
//		}
//		if _, ok := results["diff"]; ok {
//			fmt.Println("FAILURE: ", t.Name())
//			t.Error("diff table was not empty")
//			logFormatted("diff", results)
//			logFormatted("want", results)
//			logFormatted("got", results)
//		}
	}
}

func testFluxRead(t testing.TB, l *launcher.TestLauncher, file *ast.File, b *platform.Bucket ) bool {
	pass := false
	//fmt.Println("test case B: ", t.Name())

	optionsPkg := parser.ParseSource(optionsSource2)
	if ast.Check(optionsPkg) > 0 {
		panic(ast.GetError(optionsPkg))
	}
	optionsAST := optionsPkg.Files[0]

	// Define bucket and org options
	bucketOpt := &ast.OptionStatement{
		Assignment: &ast.VariableAssignment{
			ID:   &ast.Identifier{Name: "bucket"},
			Init: &ast.StringLiteral{Value: b.Name},
		},
	}
	orgOpt := &ast.OptionStatement{
		Assignment: &ast.VariableAssignment{
			ID:   &ast.Identifier{Name: "org"},
			Init: &ast.StringLiteral{Value: l.Org.Name},
		},
	}
	options := optionsAST.Copy().(*ast.File)
	options.Body = append([]ast.Statement{bucketOpt, orgOpt}, options.Body...)

	// Add options to pkg
	pkg := makeTestPackage(file)
	pkg.Files = append(pkg.Files, options)

	// Use testing.inspect call to get all of diff, want, and got
	inspectCalls := TestingTcCall(pkg)
	pkg.Files = append(pkg.Files, inspectCalls)

	bs, err := json.Marshal(pkg)
	if err != nil {
		t.Fatal(err)
	}

	req := &query.Request{
		OrganizationID: l.Org.ID,
		Compiler:       lang.ASTCompiler{AST: bs},
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

		if _, ok := results["fresh"]; ok {
			pass = true
		} else {
			fmt.Println("fresh was empty")
			t.Error("fresh was empty")
		}
	}
	return pass
}

func testFlux(t testing.TB, l *launcher.TestLauncher, file *ast.File, bs *http.BucketService) {
	b := &platform.Bucket{
		OrgID:           l.Org.ID,
		Name:            t.Name(),
		RetentionPeriod: 0,
	}

	if err := bs.CreateBucket(context.Background(), b); err != nil {
		t.Fatal(err)
	}

	// time.Sleep( 10 * time.Millisecond )

	if t.Name() == "TestFluxEndToEnd/difference_panic" {
		return
	}
	if t.Name() == "TestFluxEndToEnd/merge_filter_flag_on" {
		return
	}
	if t.Name() == "TestFluxEndToEnd/merge_filter_flag_off" {
		return
	}

	testFluxWrite(t, l, file, b)
	pass := testFluxRead(t, l, file, b)

	if !pass {
		retry := func() {
			retries := 0
			for !pass && retries < 5 {
				fmt.Println( "retrying with delays", t.Name() )

				time.Sleep( 1000 * time.Millisecond )
				testFluxWrite(t, l, file, b)

				time.Sleep( 1000 * time.Millisecond )
				pass = testFluxRead(t, l, file, b)
				retries += 1
			}
		}

		retry()

		// Remake the bucket.
		fmt.Println("remaking the bucket")
		if err := bs.CreateBucket(context.Background(), b); err != nil {
			t.Error(err)
			fmt.Println( "bucket creation failed: ", err )
		}
		time.Sleep( 1000 * time.Millisecond )

		retry()

	}
}
