package tsi1

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"sync/atomic"
	"text/tabwriter"

	"github.com/influxdata/influxdb/tsdb"
	"go.uber.org/zap"
)

const (
	// Number of series IDs to stored in slice before we convert to a roaring
	// bitmap. Roaring bitmaps have a non-trivial initial cost to construct.
	useBitmapN = 25
)

// TsiReport represents the program execution for "influxd reporttsi".
type ReportTsi struct {
	// Standard input/output, overridden for testing.
	Stderr io.Writer
	Stdout io.Writer
	Logger *zap.Logger

	Path          string
	shardPaths    map[uint64]string
	shardIdxs     map[uint64]*Index
	cardinalities map[uint64]map[string]*cardinality

	seriesFilePath string // optional. Defaults to dbPath/_series
	sfile          *tsdb.SeriesFile

	topN          int
	byMeasurement bool
	byTagKey      bool

	// How many goroutines to dedicate to calculating cardinality.
	Concurrency int
}

// NewCommand returns a new instance of Command with default setting applied.
func NewReportTsi() *ReportTsi {
	return &ReportTsi{
		Logger:        zap.NewNop(),
		shardIdxs:     map[uint64]*Index{},
		cardinalities: map[uint64]map[string]*cardinality{},
		topN:          0,
		byMeasurement: true,
		byTagKey:      false,
		Concurrency:   runtime.GOMAXPROCS(0),
	}
}

func (report *ReportTsi) RunTsiReport() error {
	// Open all the indexes.
	// Walk engine to find first each series file, then each index file
	seriesFiles := make([]*tsdb.SeriesFile, 0)
	seriesDir := filepath.Join(report.Path, "_series")
	seriesFileInfos, err := ioutil.ReadDir(seriesDir)
	if err != nil {
		return err
	}
	report.Logger.Error("searching series: " + seriesDir)
	for _, seriesFolder := range seriesFileInfos {
		folder := filepath.Join(seriesDir, seriesFolder.Name())
		//folderInfos, err := ioutil.ReadDir(folder)
		// if err != nil {
		// 	return err
		// }

		sFile := tsdb.NewSeriesFile(folder)
		sFile.WithLogger(report.Logger)
		if err := sFile.Open(context.Background()); err != nil {
			report.Logger.Error("failed")
			return err
		}
		defer sFile.Close()
		seriesFiles = append(seriesFiles, sFile)

		// for _, seriesFile := range folderInfos {
		// 	path := filepath.Join(folder, seriesFile.Name())
		// 	report.Logger.Error("adding seriesFile: " + path)
		// 	sFile := tsdb.NewSeriesFile(path)
		// 	sFile.WithLogger(report.Logger)
		// 	if err := sFile.Open(context.Background()); err != nil {
		// 		report.Logger.Error("failed")
		// 		return err
		// 	}
		// 	defer sFile.Close()
		// 	seriesFiles = append(seriesFiles, sFile)
		// }
	}

	indexFiles := make([]*IndexFile, 0)
	indexDir := filepath.Join(report.Path, "index")
	indexFileInfos, err := ioutil.ReadDir(indexDir)
	report.Logger.Error("searching index: " + indexDir)
	for _, indexFile := range indexFileInfos {
		path := filepath.Join(indexDir, indexFile.Name())
		report.Logger.Error("adding: " + path)
		indexFile := NewIndexFile(seriesFiles[len(indexFiles)])
		if err := indexFile.Open(); err != nil {
			return err
		}
		indexFiles = append(indexFiles, indexFile)
	}

	// Open all the indexes.
	// Walk engine to find first each series file, then each index file
	// TODO (me) we do not want manual entry of paths. We should be able to find all indexes
	// start path at: ./influxdbv2/engine
	for id, pth := range report.shardPaths {
		pth = path.Join(pth, "index")
		// Verify directory is an index before opening it.
		if ok, err := IsIndexDir(pth); err != nil {
			return err
		} else if !ok {
			return fmt.Errorf("not a TSI index directory: %q", pth)
		}

		report.shardIdxs[id] = NewIndex(report.sfile,
			NewConfig(),
			WithPath(pth),
			DisableCompactions(),
		)
		if err := report.shardIdxs[id].Open(context.Background()); err != nil {
			return err
		}
		defer report.shardIdxs[id].Close()

		// Initialise cardinality set to store cardinalities for this shard.
		report.cardinalities[id] = map[string]*cardinality{}
	}

	// Calculate cardinalities of shards.
	fn := report.cardinalityByMeasurement
	// if cmd.byTagKey {
	// TODO(edd)
	// }

	// Blocks until all work done.
	report.calculateCardinalities(fn)

	// Print summary.
	if err := report.printSummaryByMeasurement(); err != nil {
		return err
	}

	allIDs := make([]uint64, 0, len(report.shardIdxs))
	for id := range report.shardIdxs {
		allIDs = append(allIDs, id)
	}
	sort.Slice(allIDs, func(i int, j int) bool { return allIDs[i] < allIDs[j] })

	for _, id := range allIDs {
		if err := report.printShardByMeasurement(id); err != nil {
			return err
		}
	}
	return nil
}

// calculateCardinalities calculates the cardinalities of the set of shard being
// worked on concurrently. The provided function determines how cardinality is
// calculated and broken down.
func (report *ReportTsi) calculateCardinalities(fn func(id uint64) error) error {
	// Get list of shards to work on.
	shardIDs := make([]uint64, 0, len(report.shardIdxs))
	for id := range report.shardIdxs {
		shardIDs = append(shardIDs, id)
	}

	errC := make(chan error, len(shardIDs))
	var maxi uint32 // index of maximumm shard being worked on.
	for k := 0; k < report.Concurrency; k++ {
		go func() {
			for {
				i := int(atomic.AddUint32(&maxi, 1) - 1) // Get next partition to work on.
				if i >= len(shardIDs) {
					return // No more work.
				}
				errC <- fn(shardIDs[i])
			}
		}()
	}

	// Check for error
	for i := 0; i < cap(errC); i++ {
		if err := <-errC; err != nil {
			return err
		}
	}
	return nil
}

type cardinality struct {
	name  []byte
	short []uint32
	set   *tsdb.SeriesIDSet
}

func (c *cardinality) add(x uint64) {
	if c.set != nil {
		c.set.AddNoLock(tsdb.NewSeriesID(x))
		return
	}

	c.short = append(c.short, uint32(x)) // Series IDs never get beyond 2^32

	// Cheaper to store in bitmap.
	if len(c.short) > useBitmapN {
		c.set = tsdb.NewSeriesIDSet()
		for i := 0; i < len(c.short); i++ {
			c.set.AddNoLock(tsdb.NewSeriesID(uint64(c.short[i])))
		}
		c.short = nil
		return
	}
}

func (c *cardinality) cardinality() int64 {
	if c == nil || (c.short == nil && c.set == nil) {
		return 0
	}

	if c.short != nil {
		return int64(len(c.short))
	}
	return int64(c.set.Cardinality())
}

type cardinalities []*cardinality

func (a cardinalities) Len() int           { return len(a) }
func (a cardinalities) Less(i, j int) bool { return a[i].cardinality() < a[j].cardinality() }
func (a cardinalities) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (report *ReportTsi) cardinalityByMeasurement(shardID uint64) error {
	idx := report.shardIdxs[shardID]
	itr, err := idx.MeasurementIterator()
	if err != nil {
		return err
	} else if itr == nil {
		return nil
	}
	defer itr.Close()

OUTER:
	for {
		name, err := itr.Next()
		if err != nil {
			return err
		} else if name == nil {
			break OUTER
		}

		// Get series ID set to track cardinality under measurement.
		c, ok := report.cardinalities[shardID][string(name)]
		if !ok {
			c = &cardinality{name: name}
			report.cardinalities[shardID][string(name)] = c
		}

		sitr, err := idx.MeasurementSeriesIDIterator(name)
		if err != nil {
			return err
		} else if sitr == nil {
			continue
		}

		var e tsdb.SeriesIDElem
		for e, err = sitr.Next(); err == nil && e.SeriesID.ID != 0; e, err = sitr.Next() {
			if e.SeriesID.ID > math.MaxUint32 {
				panic(fmt.Sprintf("series ID is too large: %d (max %d). Corrupted series file?", e.SeriesID, uint32(math.MaxUint32)))
			}
			c.add(e.SeriesID.ID)
		}
		sitr.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

type result struct {
	name  []byte
	count int64

	// For low cardinality measurements just track series using map
	lowCardinality map[uint32]struct{}

	// For higher cardinality measurements track using bitmap.
	set *tsdb.SeriesIDSet
}

func (r *result) addShort(ids []uint32) {
	// There is already a bitset of this result.
	if r.set != nil {
		for _, id := range ids {
			r.set.AddNoLock(tsdb.NewSeriesID(uint64(id)))
		}
		return
	}

	// Still tracking low cardinality sets
	if r.lowCardinality == nil {
		r.lowCardinality = map[uint32]struct{}{}
	}

	for _, id := range ids {
		r.lowCardinality[id] = struct{}{}
	}

	// Cardinality is large enough that we will benefit from using a bitmap
	if len(r.lowCardinality) > useBitmapN {
		r.set = tsdb.NewSeriesIDSet()
		for id := range r.lowCardinality {
			r.set.AddNoLock(tsdb.NewSeriesID(uint64(id)))
		}
		r.lowCardinality = nil
	}
}

func (r *result) merge(other *tsdb.SeriesIDSet) {
	if r.set == nil {
		r.set = tsdb.NewSeriesIDSet()
		for id := range r.lowCardinality {
			r.set.AddNoLock(tsdb.NewSeriesID(uint64(id)))
		}
		r.lowCardinality = nil
	}
	r.set.Merge(other)
}

type results []*result

func (a results) Len() int           { return len(a) }
func (a results) Less(i, j int) bool { return a[i].count < a[j].count }
func (a results) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (report *ReportTsi) printSummaryByMeasurement() error {
	// Get global set of measurement names across shards.
	// idxs := &tsdb.IndexSet{SeriesFile: report.sfile}
	// for _, idx := range report.shardIdxs {
	// 	idxs.Indexes = append(idxs.Indexes, idx)
	// }

	// mitr, err := idxs.MeasurementIterator()
	// if err != nil {
	// 	return err
	// } else if mitr == nil {
	// 	return errors.New("got nil measurement iterator for index set")
	// }
	// defer mitr.Close()

	// var name []byte
	// var totalCardinality int64
	// measurements := results{}
	// for name, err = mitr.Next(); err == nil && name != nil; name, err = mitr.Next() {
	// 	res := &result{name: name}
	// 	for _, shardCards := range report.cardinalities {
	// 		other, ok := shardCards[string(name)]
	// 		if !ok {
	// 			continue // this shard doesn't have anything for this measurement.
	// 		}

	// 		if other.short != nil && other.set != nil {
	// 			panic("cardinality stored incorrectly")
	// 		}

	// 		if other.short != nil { // low cardinality case
	// 			res.addShort(other.short)
	// 		} else if other.set != nil { // High cardinality case
	// 			res.merge(other.set)
	// 		}

	// 		// Shard does not have any series for this measurement.
	// 	}

	// 	// Determine final cardinality and allow intermediate structures to be
	// 	// GCd.
	// 	if res.lowCardinality != nil {
	// 		res.count = int64(len(res.lowCardinality))
	// 	} else {
	// 		res.count = int64(res.set.Cardinality())
	// 	}
	// 	totalCardinality += res.count
	// 	res.set = nil
	// 	res.lowCardinality = nil
	// 	measurements = append(measurements, res)
	// }

	// if err != nil {
	// 	return err
	// }

	// // sort measurements by cardinality.
	// sort.Sort(sort.Reverse(measurements))

	// if report.topN > 0 {
	// 	// There may not be "topN" measurement cardinality to sub-slice.
	// 	n := int(math.Min(float64(report.topN), float64(len(measurements))))
	// 	measurements = measurements[:n]
	// }

	// tw := tabwriter.NewWriter(report.Stdout, 4, 4, 1, '\t', 0)
	// fmt.Fprintf(tw, "Summary\nDatabase Path: %s\nCardinality (exact): %d\n\n", report.dbPath, totalCardinality)
	// fmt.Fprint(tw, "Measurement\tCardinality (exact)\n\n")
	// for _, res := range measurements {
	// 	fmt.Fprintf(tw, "%q\t\t%d\n", res.name, res.count)
	// }

	// if err := tw.Flush(); err != nil {
	// 	return err
	// }
	// fmt.Fprint(report.Stdout, "\n\n")
	// return nil
	return nil
}

func (report *ReportTsi) printShardByMeasurement(id uint64) error {
	allMap, ok := report.cardinalities[id]
	if !ok {
		return nil
	}

	var totalCardinality int64
	all := make(cardinalities, 0, len(allMap))
	for _, card := range allMap {
		n := card.cardinality()
		if n == 0 {
			continue
		}

		totalCardinality += n
		all = append(all, card)
	}

	sort.Sort(sort.Reverse(all))

	// Trim to top-n
	if report.topN > 0 {
		// There may not be "topN" measurement cardinality to sub-slice.
		n := int(math.Min(float64(report.topN), float64(len(all))))
		all = all[:n]
	}

	tw := tabwriter.NewWriter(report.Stdout, 4, 4, 1, '\t', 0)
	fmt.Fprintf(tw, "===============\nShard ID: %d\nPath: %s\nCardinality (exact): %d\n\n", id, report.shardPaths[id], totalCardinality)
	fmt.Fprint(tw, "Measurement\tCardinality (exact)\n\n")
	for _, card := range all {
		fmt.Fprintf(tw, "%q\t\t%d\n", card.name, card.cardinality())
	}
	fmt.Fprint(tw, "===============\n\n")
	if err := tw.Flush(); err != nil {
		return err
	}
	fmt.Fprint(report.Stdout, "\n\n")
	return nil
}
