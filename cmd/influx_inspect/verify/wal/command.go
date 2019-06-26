// // exportWAL reads every WAL entry from r and exports it to w.
// FOR REFERNCE on checking wal entry corruption:
// func (cmd *Command) exportWALFile(walFilePath string, w io.Writer, warnDelete func()) error {
// 	f, err := os.Open(walFilePath)
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			fmt.Fprintf(w, "skipped missing file: %s", walFilePath)
// 			return nil
// 		}
// 		return err
// 	}
// 	defer f.Close()

// 	r := tsm1.NewWALSegmentReader(f)
// 	defer r.Close()

// 	for r.Next() {
// 		entry, err := r.Read()
// 		if err != nil {
// 			n := r.Count()
// 			fmt.Fprintf(cmd.Stderr, "file %s corrupt at position %d: %v", walFilePath, n, err)
// 			break
// 		}

// 		switch t := entry.(type) {
// 		case *tsm1.DeleteWALEntry, *tsm1.DeleteRangeWALEntry:
// 			warnDelete()
// 			continue
// 		case *tsm1.WriteWALEntry:
// 			for key, values := range t.Values {
// 				measurement, field := tsm1.SeriesAndFieldFromCompositeKey([]byte(key))
// 				// measurements are stored escaped, field names are not
// 				field = escape.Bytes(field)

// 				if err := cmd.writeValues(w, measurement, string(field), values); err != nil {
// 					// An error from writeValues indicates an IO error, which should be returned.
// 					return err
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

package wal

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"text/tabwriter"
)

type Command struct {
	Stdout io.Writer
	Stderr io.Writer

	dir string
}

func NewCommand() *Command {
	return &Command{
		Stderr: os.Stderr,
		Stdout: os.Stdout,
	}
}

func (cmd *Command) Run(args ...string) error {
	fmt.Println("running")
	var path string
	fs := flag.NewFlagSet("verify-wal", flag.ExitOnError)
	fs.StringVar(&path, "dir", os.Getenv("HOME")+"/.influxdb", "Root storage path. [$HOME/.influxdb]")

	fs.SetOutput(cmd.Stdout)
	fs.Usage = cmd.printUsage

	if err := fs.Parse(args); err != nil {
		return err
	}

	// "wal" should be the directory where all wal files are stored
	dataPath := filepath.Join(path, "wal")
	tw := tabwriter.NewWriter(cmd.Stdout, 16, 8, 0, '\t', 0)

	verifier := &verifyWAL{}
	err := verifier.Run(tw, dataPath)
	tw.Flush()
	return err
}

func (cmd *Command) printUsage() {
	usage := `Verifies the integrity of WAL (Write-ahead log) files.

Usage: influx_inspect verify-wal [flags]

    -dir <path>
            Root data path.
            Defaults to "%[1]s/.influxdb/data".
`

	fmt.Printf(usage, os.Getenv("HOME"), runtime.GOMAXPROCS(0))
}
