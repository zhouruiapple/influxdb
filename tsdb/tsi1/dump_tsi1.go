package tsi1

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"text/tabwriter"

	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/influxdb/tsdb"
	"go.uber.org/zap"
)

// Command represents the program execution for "influxd inspect dump-tsi".
type DumpTSI struct {
	// Standard input/output, overridden for testing.
	Stderr io.Writer
	Stdout io.Writer

	Logger *zap.Logger

	// Optional: defaults to DataPath/_series
	SeriesFilePath string

	// root dir of the engine
	IndexPath string

	// Optional filters on org and bucket.
	OrgID    *influxdb.ID
	BucketID *influxdb.ID

	ShowSeries         bool
	ShowMeasurements   bool
	ShowTagKeys        bool
	ShowTagValues      bool
	ShowTagValueSeries bool

	MeasurementPrefix      string
	measurementPrefixBytes []byte // reduce allocations. Set when Run called.
	TagKeyFilter           *regexp.Regexp
	TagValueFilter         *regexp.Regexp
}

// NewCommand returns a new instance of Command.
func NewDumpTSI(logger *zap.Logger) DumpTSI {
	dump := DumpTSI{
		Logger: logger,
		Stderr: os.Stderr,
		Stdout: os.Stdout,
	}
	return dump
}

// Run executes the command.
func (cmd *DumpTSI) Run() error {
	// Used for efficient prefix matching on measurement
	cmd.measurementPrefixBytes = []byte(cmd.MeasurementPrefix)

	sfile := tsdb.NewSeriesFile(cmd.SeriesFilePath)
	sfile.Logger = cmd.Logger
	if err := sfile.Open(context.Background()); err != nil {
		return err
	}
	defer sfile.Close()

	idx := NewIndex(sfile, NewConfig(), WithPath(cmd.IndexPath), DisableCompactions())
	if err := idx.Open(context.Background()); err != nil {
		return err
	}

	// Build a file set from the paths on the command line.
	// idx, fs, err := cmd.readFileSet(sfile)
	// if err != nil {
	// 	return err
	// }

	// if cmd.ShowSeries {
	// 	if err := cmd.printSeries(sfile); err != nil {
	// 		return err
	// 	}
	// }

	// If this is an ad-hoc fileset then process it and close afterward.
	// if fs != nil {
	// 	defer fs.Release()
	// 	if cmd.ShowSeries || cmd.ShowMeasurements {
	// 		return cmd.printMeasurements(sfile, fs)
	// 	}
	// 	return cmd.printFileSummaries(fs)
	// }

	// Otherwise iterate over each partition in the index.
	defer idx.Close()
	for i := 0; i < int(idx.PartitionN); i++ {
		if err := func() error {
			partition := idx.PartitionAt(i).fileSet
			// if err != nil {
			// 	return err
			// }
			defer partition.Release()

			if cmd.ShowSeries || cmd.ShowMeasurements {
				return cmd.printMeasurements(sfile, partition)
			}
			return cmd.printFileSummaries(partition)
		}(); err != nil {
			return err
		}
	}
	return nil
}

// func (cmd *DumpTSI) readFileSet(sfile *tsdb.SeriesFile) (*Index, *FileSet, error) {

// 	return index, nil, nil
// }

func (cmd *DumpTSI) printSeries(sfile *tsdb.SeriesFile) error {
	if !cmd.ShowSeries {
		return nil
	}

	// Print header.
	tw := tabwriter.NewWriter(cmd.Stdout, 8, 8, 1, '\t', 0)
	fmt.Fprintln(tw, "Series\t")

	// Iterate over each series.
	itr := sfile.SeriesIDIterator()
	for {
		e, err := itr.Next()
		if err != nil {
			return err
		} else if e.SeriesID.ID == 0 {
			break
		}
		name, tags := tsdb.ParseSeriesKey(sfile.SeriesKey(e.SeriesID))

		// The measurement prefix matches (or prefix matching is not enabled) and the series matches.
		if cmd.matchMeasurement(name) && cmd.matchSeries(name, tags) {
			deleted := sfile.IsDeleted(e.SeriesID)
			fmt.Fprintf(tw, "%x%s\t%v\n", name, tags.HashKey(), deletedString(deleted))
		}
	}

	// Flush & write footer spacing.
	if err := tw.Flush(); err != nil {
		return err
	}
	fmt.Fprint(cmd.Stdout, "\n\n")

	return nil
}

func (cmd *DumpTSI) printMeasurements(sfile *tsdb.SeriesFile, fs *FileSet) error {
	if !cmd.ShowMeasurements {
		return nil
	}

	tw := tabwriter.NewWriter(cmd.Stdout, 8, 8, 1, '\t', 0)
	fmt.Fprintln(tw, "Measurement\t")

	// Iterate over each series.
	if itr := fs.MeasurementIterator(); itr != nil {
		for e := itr.Next(); e != nil; e = itr.Next() {
			if !cmd.matchMeasurement(e.Name()) {
				continue
			}

			fmt.Fprintf(tw, "%x\t%v\n", e.Name(), deletedString(e.Deleted()))
			if err := tw.Flush(); err != nil {
				return err
			}

			if err := cmd.printTagKeys(sfile, fs, e.Name()); err != nil {
				return err
			}
		}
	}

	fmt.Fprint(cmd.Stdout, "\n\n")
	return nil
}

var measurementTagKey = []byte("_measurement (\\x00)")
var fieldTagKey = []byte("_field (\\xff)")

func (cmd *DumpTSI) printTagKeys(sfile *tsdb.SeriesFile, fs *FileSet, name []byte) error {
	if !cmd.ShowTagKeys {
		return nil
	}

	// Iterate over each key.
	tw := tabwriter.NewWriter(cmd.Stdout, 8, 8, 1, '\t', 0)
	itr := fs.TagKeyIterator(name)
	for e := itr.Next(); e != nil; e = itr.Next() {
		key := e.Key()
		if bytes.Equal(key, models.MeasurementTagKeyBytes) {
			key = measurementTagKey
		} else if bytes.Equal(key, models.FieldKeyTagKeyBytes) {
			key = fieldTagKey
		}

		if cmd.TagKeyFilter != nil && !cmd.TagKeyFilter.Match(key) {
			continue
		}

		fmt.Fprintf(tw, "    %s\t%v\n", key, deletedString(e.Deleted()))
		if err := tw.Flush(); err != nil {
			return err
		}

		if err := cmd.printTagValues(sfile, fs, name, key); err != nil {
			return err
		}
	}
	fmt.Fprint(cmd.Stdout, "\n")

	return nil
}

func (cmd *DumpTSI) printTagValues(sfile *tsdb.SeriesFile, fs *FileSet, name, key []byte) error {
	if !cmd.ShowTagValues {
		return nil
	}

	// Iterate over each value.
	tw := tabwriter.NewWriter(cmd.Stdout, 8, 8, 1, '\t', 0)
	itr := fs.TagValueIterator(name, key)
	for e := itr.Next(); e != nil; e = itr.Next() {
		if cmd.TagValueFilter != nil && !cmd.TagValueFilter.Match(e.Value()) {
			continue
		}

		fmt.Fprintf(tw, "        %s\t%v\n", e.Value(), deletedString(e.Deleted()))
		if err := tw.Flush(); err != nil {
			return err
		}

		if err := cmd.printTagValueSeries(sfile, fs, name, key, e.Value()); err != nil {
			return err
		}
	}
	fmt.Fprint(cmd.Stdout, "\n")

	return nil
}

func (cmd *DumpTSI) printTagValueSeries(sfile *tsdb.SeriesFile, fs *FileSet, name, key, value []byte) error {
	if !cmd.ShowTagValueSeries {
		return nil
	}

	// Iterate over each series.
	tw := tabwriter.NewWriter(cmd.Stdout, 8, 8, 1, '\t', 0)
	itr, err := fs.TagValueSeriesIDIterator(name, key, value)
	if err != nil {
		return err
	}
	for {
		e, err := itr.Next()
		if err != nil {
			return err
		} else if e.SeriesID.ID == 0 {
			break
		}

		name, tags := tsdb.ParseSeriesKey(sfile.SeriesKey(e.SeriesID))

		if !cmd.matchSeries(name, tags) {
			continue
		}

		fmt.Fprintf(tw, "            %s%s\n", name, tags.HashKey())
		if err := tw.Flush(); err != nil {
			return err
		}
	}
	fmt.Fprint(cmd.Stdout, "\n")

	return nil
}

func (cmd *DumpTSI) printFileSummaries(fs *FileSet) error {
	for _, f := range fs.Files() {
		switch f := f.(type) {
		case *LogFile:
			fmt.Printf("got an alleged LogFile: %v\n", f.Path())
			if err := cmd.printLogFileSummary(f); err != nil {
				return err
			}
		case *IndexFile:
			if err := cmd.printIndexFileSummary(f); err != nil {
				return err
			}
		default:
			panic("unreachable")
		}
		fmt.Fprintln(cmd.Stdout, "")
	}
	return nil
}

func (cmd *DumpTSI) printLogFileSummary(f *LogFile) error {
	fmt.Fprintf(cmd.Stdout, "[LOG FILE] %s\n", filepath.Base(f.Path()))
	tw := tabwriter.NewWriter(cmd.Stdout, 8, 8, 1, '\t', 0)
	fmt.Fprintf(tw, "Series:\t%d\n", f.SeriesN())
	fmt.Fprintf(tw, "Measurements:\t%d\n", f.MeasurementN())
	fmt.Fprintf(tw, "Tag Keys:\t%d\n", f.TagKeyN())
	fmt.Fprintf(tw, "Tag Values:\t%d\n", f.TagValueN())
	return tw.Flush()
}

func (cmd *DumpTSI) printIndexFileSummary(f *IndexFile) error {
	fmt.Fprintf(cmd.Stdout, "[INDEX FILE] %s\n", filepath.Base(f.Path()))

	// Calculate summary stats.
	var measurementN, measurementSeriesN, measurementSeriesSize uint64
	var keyN uint64
	var valueN, valueSeriesN, valueSeriesSize uint64

	if mitr := f.MeasurementIterator(); mitr != nil {
		for me, _ := mitr.Next().(*MeasurementBlockElem); me != nil; me, _ = mitr.Next().(*MeasurementBlockElem) {
			kitr := f.TagKeyIterator(me.Name())
			for ke, _ := kitr.Next().(*TagBlockKeyElem); ke != nil; ke, _ = kitr.Next().(*TagBlockKeyElem) {
				vitr := f.TagValueIterator(me.Name(), ke.Key())
				for ve, _ := vitr.Next().(*TagBlockValueElem); ve != nil; ve, _ = vitr.Next().(*TagBlockValueElem) {
					valueN++
					valueSeriesN += uint64(ve.SeriesN())
					valueSeriesSize += uint64(len(ve.SeriesData()))
				}
				keyN++
			}
			measurementN++
			measurementSeriesN += uint64(me.SeriesN())
			measurementSeriesSize += uint64(len(me.SeriesData()))
		}
	}

	// Write stats.
	tw := tabwriter.NewWriter(cmd.Stdout, 8, 8, 1, '\t', 0)
	fmt.Fprintf(tw, "Measurements:\t%d\n", measurementN)
	fmt.Fprintf(tw, "  Series data size:\t%d (%s)\n", measurementSeriesSize, formatSize(measurementSeriesSize))
	fmt.Fprintf(tw, "  Bytes per series:\t%.01fb\n", float64(measurementSeriesSize)/float64(measurementSeriesN))
	fmt.Fprintf(tw, "Tag Keys:\t%d\n", keyN)
	fmt.Fprintf(tw, "Tag Values:\t%d\n", valueN)
	fmt.Fprintf(tw, "  Series:\t%d\n", valueSeriesN)
	fmt.Fprintf(tw, "  Series data size:\t%d (%s)\n", valueSeriesSize, formatSize(valueSeriesSize))
	fmt.Fprintf(tw, "  Bytes per series:\t%.01fb\n", float64(valueSeriesSize)/float64(valueSeriesN))
	return tw.Flush()
}

// matchMeasurement returns true if measurement matching is disabled or if the measurement
// does match a measurement prefix.
func (cmd *DumpTSI) matchMeasurement(name []byte) bool {
	if cmd.MeasurementPrefix == "" {
		return true
	}

	// Convert name into base-16.
	nameHex := []byte(fmt.Sprintf("%x", name))
	return bytes.HasPrefix(nameHex, cmd.measurementPrefixBytes)
}

// matchSeries returns true if the command filters matches the series.
func (cmd *DumpTSI) matchSeries(name []byte, tags models.Tags) bool {
	// Filter by tag key/value.
	if cmd.TagKeyFilter != nil || cmd.TagValueFilter != nil {
		var matched bool
		for _, tag := range tags {
			if (cmd.TagKeyFilter == nil || cmd.TagKeyFilter.Match(tag.Key)) && (cmd.TagValueFilter == nil || cmd.TagValueFilter.Match(tag.Value)) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// deletedString returns "(deleted)" if v is true.
func deletedString(v bool) string {
	if v {
		return "(deleted)"
	}
	return ""
}

func formatSize(v uint64) string {
	denom := uint64(1)
	var uom string
	for _, uom = range []string{"b", "kb", "mb", "gb", "tb"} {
		if denom*1024 > v {
			break
		}
		denom *= 1024
	}
	return fmt.Sprintf("%0.01f%s", float64(v)/float64(denom), uom)
}
