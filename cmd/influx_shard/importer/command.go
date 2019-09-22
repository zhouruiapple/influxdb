package importer

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb/services/meta"
	"github.com/influxdata/influxdb/tsdb/engine/tsm1"
	"go.uber.org/zap"
)

// Command represents the program execution for "store query".
type Command struct {
	// Standard input/output, overridden for testing.
	Stderr io.Writer
	Stdin  io.Reader
	Logger *zap.Logger

	dataDir         string
	metadb          string
	database        string
	retentionPolicy string
	srcShardPath    string
	replace         bool
}

// NewCommand returns a new instance of Command.
func NewCommand() *Command {
	return &Command{
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}
}

// Run executes the import command using the specified args.
func (cmd *Command) Run(args []string) (err error) {
	// Parse command line flags.
	err = cmd.parseFlags(args)
	if err != nil {
		return err
	}

	// Load the destination InfluxDB's meta data.
	mdata, err := ReadMetaData(cmd.metadb)
	if err != nil {
		return err
	}

	// Find or create the destination database.
	db, rp := cmd.database, cmd.retentionPolicy

	dbi, rpi, err := CreateDBAndRP(mdata, db, rp)
	if err != nil {
		return err
	}

	_, _ = dbi, rpi

	// Parse the shard ID from the source shard's path.
	shardID, err := ShardIDFromPath(cmd.srcShardPath)
	if err != nil {
		return err
	}

	// Open the source shard.
	shard := NewShard(shardID, cmd.srcShardPath)
	if err := shard.Open(); err != nil {
		return err
	}
	defer shard.Close()

	// Get the min & max time from the source shard.
	min, max := shard.TimeRange()

	// Find all the shard groups in the destination database and retention policy
	// that could contain the source shard. Hopefully there's just one.
	sgs, err := mdata.ShardGroupsByTimeRange(db, rp, min, max)
	if err != nil {
		return err
	}

	if len(sgs) == 0 {
		if err := mdata.CreateShardGroup(db, rp, min); err != nil {
			return err
		}
		sgs, err = mdata.ShardGroupsByTimeRange(db, rp, min, max)
		if err != nil {
			return err
		}
	}

	if len(sgs) > 1 {
		return fmt.Errorf("expected 1 shard group that could contain this shard but found %d", len(sgs))
	}

	if err := mdata.CreateShardGroup(db, rp, min); err != nil {
		return err
	}

	sgi, err := mdata.ShardGroupByTimestamp(db, rp, min)
	if err != nil {
		return err
	}

	if sgi == nil {
		return fmt.Errorf("error finding shard group for shard")
	}

	// Copy the shard data to the destination.
	newPath := filepath.Join(cmd.dataDir, db, rp, strconv.FormatUint(shardID, 10))
	shard, err = shard.Copy(newPath)
	if err != nil {
		return err
	}

	// Add the shard to the shard group.
	sgi.Shards = append(sgi.Shards, meta.ShardInfo{
		ID:     shardID,
		Owners: []meta.ShardOwner{}, // TODO: set owners?
	})

	// Write meta data back to disk.
	if err := WriteMetaData(cmd.metadb, mdata); err != nil {
		os.RemoveAll(newPath)
		return err
	}

	return err
}

// ShardIDFromPath parses an int64 shard ID from a path to a shard directory, which
// looks something like /path/to/1234.
func ShardIDFromPath(path string) (uint64, error) {
	path = strings.Trim(path, "/")
	_, idstr := filepath.Split(path)
	return strconv.ParseUint(idstr, 10, 64)
}

// CreateDBAndRP is idempotent and creates a DB and RP if needed.
func CreateDBAndRP(data *meta.Data, dbname, rpname string) (*meta.DatabaseInfo, *meta.RetentionPolicyInfo, error) {
	dbi := data.Database(dbname)
	if dbi == nil {
		if err := data.CreateDatabase(dbname); err != nil {
			return nil, nil, err
		}
		dbi = data.Database(dbname)
		if dbi == nil {
			return nil, nil, errors.New("database not found after sucessfully creating it")
		}
	}

	rpi, err := data.RetentionPolicy(dbname, rpname)
	if err != nil {
		return nil, nil, err
	}

	if rpi == nil {
		rpi = meta.DefaultRetentionPolicyInfo()
		if err := data.CreateRetentionPolicy(dbname, rpi, true); err != nil {
			return nil, nil, err
		}
	}

	return dbi, rpi, nil
}

// ReadMetaData reads InfluxDB OSS meta data from disk.
func ReadMetaData(metadb string) (*meta.Data, error) {
	b, err := ioutil.ReadFile(metadb)
	if err != nil {
		return nil, fmt.Errorf("ReadMetaData: ReadFile: %s", err)
	}

	data := &meta.Data{}
	if err := data.UnmarshalBinary(b); err != nil {
		return nil, fmt.Errorf("ReadMetaData: UnmarshalBinary: %s", err)
	}

	return data, nil
}

// WriteMetaData writes InfluxDB OSS meta data to disk.
func WriteMetaData(metadb string, data *meta.Data) error {
	b, err := data.MarshalBinary()
	if err != nil {
		return fmt.Errorf("WriteMetaData: MarshalBinary: %s", err)
	}

	if err = ioutil.WriteFile(metadb, b, 0666); err != nil {
		return fmt.Errorf("WriteMetaData: Write: %s", err)
	}

	return nil
}

// parseFlags parses the import command's command line flags.
func (cmd *Command) parseFlags(args []string) error {
	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	fs.StringVar(&cmd.dataDir, "data", "", "Destination data directory")
	fs.StringVar(&cmd.metadb, "meta", "", "Destination meta.db")
	fs.StringVar(&cmd.database, "db", "", "Destination database name")
	fs.StringVar(&cmd.retentionPolicy, "rp", "", "Destination retention policy")
	fs.StringVar(&cmd.srcShardPath, "shard", "", "Source shard directory")
	fs.BoolVar(&cmd.replace, "replace", false, "Enables replacing an existing shard")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if cmd.dataDir == "" {
		return errors.New("destination data directory required")
	}

	if cmd.metadb == "" {
		return errors.New("destination path/to/meta.db required")
	}

	if cmd.database == "" {
		return errors.New("destination database required")
	}

	if cmd.retentionPolicy == "" {
		return errors.New("destination retention policy required")
	}

	if cmd.srcShardPath == "" {
		return errors.New("source shard directory required")
	}

	return nil
}

// Shard represents a set of TSM files belonging to a single shard.
type Shard struct {
	ID    uint64
	Path  string
	Files []*tsm1.TSMReader
}

// NewShard returns a new *Shard with the specified shard's info.
func NewShard(id uint64, path string) *Shard {
	return &Shard{
		ID:    id,
		Path:  path,
		Files: []*tsm1.TSMReader{},
	}
}

// Open opens all the TSM files in the shard.
func (s *Shard) Open() error {
	// Get a list of TSM files in the path.
	pattern := filepath.Join(s.Path, "*.tsm")
	filenames, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// Open readers for the TSM files.
	for _, name := range filenames {
		r, err := OpenTSMReader(name)
		if err != nil {
			return fmt.Errorf("error opening TSMReader: %s", err)
		}
		s.Files = append(s.Files, r)
	}

	return nil
}

// Close closes all the TSM files in the shard.
func (s *Shard) Close() error {
	var err error
	for _, r := range s.Files {
		tmperr := r.Close()
		if tmperr != nil && err == nil {
			err = tmperr
		}
	}
	s.Files = []*tsm1.TSMReader{}
	return err
}

// TimeRange returns the time range for the shard.
func (s *Shard) TimeRange() (time.Time, time.Time) {
	var min, max int64 = math.MaxInt64, math.MinInt64

	for _, r := range s.Files {
		s, e := r.TimeRange()
		if s < min {
			min = s
		}
		if e > max {
			max = e
		}
	}

	return time.Unix(0, min), time.Unix(0, max)
}

// Copy copies the shard to another directory, opens it, and returns the new shard.
func (s *Shard) Copy(dest string) (*Shard, error) {
	fmt.Printf("copying: %q -> %q\n", s.Path, dest)
	// Error if destination already exists.
	_, err := os.Stat(dest)
	if err == nil {
		return nil, fmt.Errorf("copy shard: %q already exists", dest)
	}

	// Close source TSM files so we can copy.
	if err := s.Close(); err != nil {
		return nil, fmt.Errorf("copy shard: close: %s", err)
	}

	// Create the destination directory.
	if err := os.MkdirAll(dest, 0777); err != nil {
		return nil, fmt.Errorf("copy shard: mkdir: %s", err)
	}

	// If commit == false when this function exists, destination will be deleted.
	commit := false
	defer func() {
		if !commit {
			os.RemoveAll(dest)
		}
	}()

	// Get a list of TSM files in the path.
	pattern := filepath.Join(s.Path, "*")
	filenames, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Copy each file.
	for _, from := range filenames {
		_, filename := filepath.Split(from)
		to := filepath.Join(dest, filename)
		if err := CopyFile(from, to); err != nil {
			return nil, err
		}
	}

	// Open the new shard.
	shard := NewShard(s.ID, dest)
	if err := shard.Open(); err != nil {
		return nil, err
	}

	commit = true

	return shard, nil
}

// CopyFile copies a file.
func CopyFile(src, dest string) error {
	from, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("CopyFile: open src: %s", err)
	}
	defer from.Close()

	to, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("CopyFile: create dest: %s", err)
	}
	defer to.Close()

	_, err = io.Copy(to, from)

	return err
}

// OpenTSMReader opens a .tsm file and returns a TSMReader for it.
func OpenTSMReader(path string) (*tsm1.TSMReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return tsm1.NewTSMReader(f)
}
