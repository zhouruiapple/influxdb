// inspects low-level details about tsi1 files.
package inspect

import (
	"errors"
	"io"
	"path/filepath"
	"regexp"

	"github.com/influxdata/influxdb/internal/fs"
	"github.com/influxdata/influxdb/tsdb/tsi1"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// Command represents the program execution for "influxd dumptsi".
var measurementFilter, tagKeyFilter, tagValueFilter string
var dumpTSIFlags = struct {
	// Standard input/output, overridden for testing.
	Stderr io.Writer
	Stdout io.Writer

	orgID    string
	bucketID string

	seriesFilePath string
	tsiPath        string

	showSeries         bool
	showMeasurements   bool
	showTagKeys        bool
	showTagValues      bool
	showTagValueSeries bool

	measurementPrefixHex string
	tagKeyFilter         *regexp.Regexp
	tagValueFilter       *regexp.Regexp
}{}

// NewCommand returns a new instance of Command.
func NewDumpTSICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dump-tsi",
		Short: "Dump low level tsi information",
		Long: `Dumps low-level details about tsi1 files.

This tool emits low-level details about the TSI index and Series file. If 
running on a machine being used for production workloads it should be used
with caution.

This tool lets you output series, measurement names, tag keys and values, and allows
for regex filters on those. Further, you can limit the output to and org and/or 
bucket.
		`,
		RunE: dumpTsi,
	}
	defaultDataDir, _ := fs.InfluxDir()
	defaultIndexDir := filepath.Join(defaultDataDir, "engine", "index")
	defaultSeriesDir := filepath.Join(defaultDataDir, "engine", "_series")

	cmd.Flags().StringVar(&dumpTSIFlags.seriesFilePath, "series-file", defaultSeriesDir, "path to series file")
	cmd.Flags().StringVar(&dumpTSIFlags.tsiPath, "tsi-index", defaultIndexDir, "path to the the TSI index")

	cmd.Flags().StringVar(&dumpTSIFlags.orgID, "org_id", "", "limit results to provided organisation")
	cmd.Flags().StringVar(&dumpTSIFlags.bucketID, "bucket_id", "", "limit results to provided bucket (org_id required)")

	cmd.Flags().BoolVar(&dumpTSIFlags.showSeries, "series", false, "emit raw series data")

	cmd.Flags().BoolVar(&dumpTSIFlags.showMeasurements, "measurements", false, "emit raw measurement data")
	cmd.Flags().StringVar(&dumpTSIFlags.measurementPrefixHex, "measurement-prefix", "", "filter measurements by provided base-16 prefix")

	cmd.Flags().BoolVar(&dumpTSIFlags.showTagKeys, "tag-keys", false, "emit raw tag key data")
	cmd.Flags().StringVar(&tagKeyFilter, "tag-key-filter", "", "filter tag keys by regex")

	cmd.Flags().BoolVar(&dumpTSIFlags.showTagValues, "tag-values", false, "emit raw tag value data")
	cmd.Flags().StringVar(&tagValueFilter, "tag-value-filter", "", "filter tag values by regex")
	cmd.Flags().BoolVar(&dumpTSIFlags.showTagValueSeries, "tag-value-series", false, "emit raw series for each tag value")

	return cmd
}

func dumpTsi(cmd *cobra.Command, args []string) error {
	logger := zap.NewNop()

	// Parse filters.
	if tagKeyFilter != "" {
		re, err := regexp.Compile(tagKeyFilter)
		if err != nil {
			return err
		}
		dumpTSIFlags.tagKeyFilter = re
	}
	if tagValueFilter != "" {
		re, err := regexp.Compile(tagValueFilter)
		if err != nil {
			return err
		}
		dumpTSIFlags.tagValueFilter = re
	}

	if dumpTSIFlags.tsiPath == "" {
		return errors.New("data path must be specified")
	}

	// Some flags imply other flags.
	if dumpTSIFlags.showTagValueSeries {
		dumpTSIFlags.showTagValues = true
	}
	if dumpTSIFlags.showTagValues {
		dumpTSIFlags.showTagKeys = true
	}
	if dumpTSIFlags.showTagKeys {
		dumpTSIFlags.showMeasurements = true
	}

	dump := tsi1.NewDumpTSI(logger)
	dump.SeriesFilePath = dumpTSIFlags.seriesFilePath
	dump.IndexPath = dumpTSIFlags.tsiPath
	dump.ShowSeries = dumpTSIFlags.showSeries
	dump.ShowMeasurements = dumpTSIFlags.showMeasurements
	dump.ShowTagKeys = dumpTSIFlags.showTagKeys
	dump.ShowTagValues = dumpTSIFlags.showTagValues
	dump.ShowTagValueSeries = dumpTSIFlags.showTagValueSeries
	dump.MeasurementPrefix = dumpTSIFlags.measurementPrefixHex
	dump.TagKeyFilter = dumpTSIFlags.tagKeyFilter
	dump.TagValueFilter = dumpTSIFlags.tagValueFilter

	return dump.Run()
}
