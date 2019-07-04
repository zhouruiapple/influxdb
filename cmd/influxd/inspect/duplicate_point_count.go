package inspect

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/influxdata/influxdb/tsdb/tsm1"
	"github.com/spf13/cobra"
)

func NewDuplicatePointCountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `duplicate-point-count`,
		Short: "Counts the number of duplicate points",
		Long: `
This command will return a count of the number of duplicate points
across a set of TSM1 files.`,
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("tsm1 file path required")
		} else if len(args) > 1 {
			return fmt.Errorf("only one path allowed")
		}

		f, err := os.OpenFile(args[0], os.O_RDONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()

		r, err := tsm1.NewTSMReader(f)
		if err != nil {
			return err
		}
		defer r.Close()

		itr := r.BlockIterator()
		if itr == nil {
			return errors.New("invalid TSM file, no block iterator")
		}

		var vals tsm1.Values
		var totalN, duplicateN int

		// Counts duplicate points and clears points on key change.
		var prevKey []byte
		check := func(key []byte) {
			if bytes.Equal(prevKey, key) {
				return
			}

			n, dedupN := len(vals), len(vals.Deduplicate())
			duplicateN += (n - dedupN)
			totalN += n

			prevKey, vals = key, nil
		}

		// Iterate over every block.
		for itr.Next() {
			// Read next block and check duplicates if key changes.
			key, _, _, _, _, buf, err := itr.Read()
			if err != nil {
				return err
			}
			check(key)

			// Read all points for this block.
			tmp, err := tsm1.DecodeBlock(buf, nil)
			if err != nil {
				return err
			}
			vals = append(vals, tmp...)
		}

		// Final duplicate point check after last block.
		check(nil)

		// Report total duplicate point count.
		fmt.Printf("total: %d\n", totalN)
		fmt.Printf("duplicate: %d (%0.02f%%)\n", duplicateN, (float64(duplicateN)/float64(totalN))*100)

		return nil
	}

	return cmd
}
