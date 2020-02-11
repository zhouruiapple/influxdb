package main

import (
	"context"
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/repl"
	_ "github.com/influxdata/flux/stdlib"
	"github.com/influxdata/influxdb/http"
	"github.com/influxdata/influxdb/query"
	_ "github.com/influxdata/influxdb/query/stdlib"
	"github.com/spf13/cobra"
	"os"
)

var queryFlags struct {
	org             organization
	enableProfiling bool
}

func cmdQuery() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query [query literal or @/path/to/query.flux]",
		Short: "Execute a Flux query",
		Long: `Execute a literal Flux query provided as a string,
or execute a literal Flux query contained in a file by specifying the file prefixed with an @ sign.`,
		Args: cobra.ExactArgs(1),
		RunE: wrapCheckSetup(fluxQueryF),
	}
	cmd.Flags().BoolVar(&queryFlags.enableProfiling, "profile", false, "enable query profiling")
	queryFlags.org.register(cmd, true)

	return cmd
}

func fluxQueryF(cmd *cobra.Command, args []string) error {
	if flags.local {
		return fmt.Errorf("local flag not supported for query command")
	}

	if err := queryFlags.org.validOrgFlags(); err != nil {
		return err
	}

	q, err := repl.LoadQuery(args[0])
	if err != nil {
		return fmt.Errorf("failed to load query: %v", err)
	}

	orgSvc, err := newOrganizationService()
	if err != nil {
		return fmt.Errorf("failed to initialized organization service client: %v", err)
	}

	orgID, err := queryFlags.org.getID(orgSvc)
	if err != nil {
		return err
	}

	flux.FinalizeBuiltIns()

	qs := &http.FluxQueryService{
		Addr:               flags.host,
		Token:              flags.token,
		InsecureSkipVerify: flags.skipVerify,
	}
	req := &query.Request{
		OrganizationID: orgID,
		Compiler: lang.FluxCompiler{
			Query: q,
		},
	}
	if queryFlags.enableProfiling {
		req.WithProfiling()
	}

	ri, err := qs.Query(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}
	defer ri.Release()
	for ri.More() {
		if err := execute.FormatResult(os.Stdout, ri.Next()); err != nil {
			return fmt.Errorf("error in formatting result: %v", err)
		}
	}
	ri.Release()
	if err := ri.Err(); err != nil {
		return fmt.Errorf("error in executing query: %v", err)
	}
	if queryFlags.enableProfiling {
		return renderStats(ri.Statistics())
	}
	return nil
}

func renderStats(stats flux.Statistics) error {
	fmt.Println()
	fmt.Println("Profiling stats:")
	if prof, ok := stats.Metadata["profiling"]; ok {
		p := prof[0].(map[string]interface{})
		for k, v := range p {
			fmt.Printf("%s:\n", k)
			mv := v.(map[string]interface{})
			for k, v := range mv {
				if k == "key" {
					continue
				}
				fmt.Printf("\t%s -> %v\n", k, v)
			}
		}
	}
	fmt.Println()
	/*
	fmt.Println("Profiling stats - detail:")
	for k, v := range stats.Metadata {
		fmt.Printf("%s -> %v\n", k, v)
	}
	 */
	return nil
}
