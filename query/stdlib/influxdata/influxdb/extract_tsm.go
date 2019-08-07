package influxdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"sort"
	"time"

	"github.com/influxdata/flux"

	"github.com/influxdata/flux/execute"

	"github.com/influxdata/influxdb/tsdb"
	"github.com/influxdata/influxdb/tsdb/tsm1"

	"github.com/influxdata/influxdb"

	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	streamArg      = "stream"
	bucketArg      = "bucket"
	orgArg         = "org"
	startArg       = "start"
	stopArg        = "stop"
	ExtractTSMKind = "extractTSM"
)

func init() {
	flux.RegisterPackageValue("influxdata/influxdb", ExtractTSMKind, extractTSM)
}

var extractTSM = values.NewFunction(
	"extractTSM",
	semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			streamArg: semantic.Stream,
			startArg:  semantic.Time,
			stopArg:   semantic.Time,
			orgArg:    semantic.String,
			bucketArg: semantic.String,
		},
		PipeArgument: streamArg,
		Required:     semantic.LabelSet{streamArg, orgArg},
		Return:       semantic.Stream,
	}),
	func(args values.Object) (values.Value, error) {
		stream, exists := args.Get(streamArg)
		if !exists {
			return nil, fmt.Errorf("missing argument %s", streamArg)
		} else if stream.Type().Nature() != semantic.Stream {
			return nil, errors.New("stream must be of type stream")
		}

		var startTime, stopTime values.Time
		start, exists := args.Get(startArg)
		if !exists {
			startTime = values.ConvertTime(flux.MinTime.Absolute)
		} else if start.Type().Nature() != semantic.Time {
			return nil, errors.New("start must be of type time")
		} else {
			startTime = start.Time()
		}

		stop, exists := args.Get(stopArg)
		if !exists {
			stopTime = values.ConvertTime(flux.MaxTime.Absolute)
		} else if stop.Type().Nature() != semantic.Time {
			return nil, errors.New("stop must be of type time")
		} else {
			stopTime = stop.Time()
		}

		var org, bucket string
		orgArg, exists := args.Get(orgArg)
		if !exists {
			return nil, fmt.Errorf("missing argument %s", orgArg)
		} else if orgArg.Type().Nature() != semantic.String {
			return nil, fmt.Errorf("org id must be string")
		}
		org = orgArg.Str()

		bucketArg, exists := args.Get(bucketArg)
		if !exists {
			bucket = ""
		} else if bucketArg.Type().Nature() != semantic.String {
			return nil, fmt.Errorf("bucket id must be string")
		} else {
			bucket = bucketArg.Str()
		}

		timeBounds := execute.Bounds{
			Start: startTime,
			Stop:  stopTime,
		}

		tsmFilter, err := NewTSMFilter(org, bucket, timeBounds, stream)
		if err != nil {
			return nil, err
		}

		outStream, err := tsmFilter.FilteredTSMStream()
		if err != nil {
			return nil, err
		}

		return values.NewReadStream(outStream), nil
	}, false,
)

// TSMFilter contains parameters and state for
// filtering a remote TSM file based on certain conditions,
// and creating a new resulting TSM file
type TSMFilter struct {
	Org, Bucket *influxdb.ID
	Bounds      execute.Bounds
	source      values.Stream

	buf io.ReadWriter

	dataSize uint32
}

func NewTSMFilter(org, bucket string, bounds execute.Bounds, src values.Stream) (*TSMFilter, error) {
	orgID, err := influxdb.IDFromString(org)
	if err != nil {
		return nil, err
	}

	var bucketID *influxdb.ID
	if bucket != "" {
		bucketID, err = influxdb.IDFromString(bucket)
		if err != nil {
			return nil, err
		}
	}

	return &TSMFilter{
		Org:    orgID,
		Bucket: bucketID,
		Bounds: bounds,
		source: src,
		buf:    &bytes.Buffer{},
	}, nil
}

// FilteredTSMStream returns a read stream to a tsm file
// containing only data which matches predicates in the TSMFilter struct.
func (t *TSMFilter) FilteredTSMStream() (io.Reader, error) {
	s, _ := t.source.(io.ReadSeeker)
	entries, err := t.filterBlocks(s, t.Org, t.Bucket, t.Bounds)
	if err != nil {
		return nil, err
	}

	chunks := segmentToChunks(entries)
	if err := t.BuildOutputTSM(chunks, s); err != nil {
		return nil, err
	}

	return t.buf, nil
}

//BuildOutputTSM downloads all relevant tsm data and writes the blocks into
// a new tsm file (currently an in-memory buffer)
func (t *TSMFilter) BuildOutputTSM(chunks []*DownloadChunk, stream io.ReadSeeker) error {
	tsmWriter, err := tsm1.NewTSMWriter(t.buf)
	if err != nil {
		return err
	}

	for _, chunk := range chunks {
		iter, err := chunk.BlockIterator(stream)
		if err != nil {
			return err
		}

		var read uint32 = 0
		last := time.Now()
		for iter.HasNext() {
			if err := iter.Next(); err != nil {
				return err
			}

			entryBytes := iter.BlockBytes()
			entryKey := iter.SeriesKey()
			minTime, maxTime := iter.BlockMinMaxTime()

			delta := time.Since(last)
			rate := float64(len(entryBytes)) / 1000000 / delta.Seconds()
			last = time.Now()
			fmt.Printf("\r %f MiB/s wrote %d out of total %d; %3.2f%% finished", rate, read, t.dataSize, float32(read)/float32(t.dataSize)*100)

			if err := tsmWriter.WriteBlock(entryKey, minTime, maxTime, entryBytes); err != nil {
				return err
			}
			read += uint32(len(entryBytes))
		}
	}

	if err := tsmWriter.WriteIndex(); err != nil {
		return err
	}

	if err := tsmWriter.Close(); err != nil {
		return err
	}

	return nil
}

// BlockIterator represents state for iterating over
// the blocks in a downloaded chunk of a tsm file
type BlockIterator struct {
	offset int64

	n       int
	blocks  []byte
	entries []*ReadEntry

	block     []byte
	blockInfo *ReadEntry
}

// HasNext reports whether or not the iterator has another entry to return
func (b *BlockIterator) HasNext() bool {
	return b.n < len(b.entries)
}

// Next points the block iterator at the next block data and information
func (b *BlockIterator) Next() error {
	info, contents := b.blockAt(b.n)

	contents, err := stripAndVerifyChecksum(contents)
	if err != nil {
		return err
	}
	b.block = contents
	b.blockInfo = info
	b.n++

	return nil
}

func (b *BlockIterator) blockAt(i int) (*ReadEntry, []byte) {
	entry := b.entries[i]
	start := entry.blockData.Offset - b.offset
	end := start + int64(entry.blockData.Size)

	return entry, b.blocks[start:end]
}

// stripAndVerifyChecksum removes the first 4 bytes from a downloaded block (the checksum), and
// verifies the checksum
func stripAndVerifyChecksum(block []byte) ([]byte, error) {
	if len(block) < 4 {
		return nil, errors.New("block too short")
	}

	sum := block[:4]
	block = block[4:]

	if err := verifyChecksum(sum, block); err != nil {
		return nil, err
	}

	return block, nil
}

// verifyChecksum verifies whether or not the checksum in a block matches
// the actual block data
func verifyChecksum(want []byte, data []byte) error {
	var checksum [crc32.Size]byte
	binary.BigEndian.PutUint32(checksum[:], crc32.ChecksumIEEE(data))

	if bytes.Compare(want, checksum[:]) != 0 {
		return errors.New("invalid checksum for block")
	}

	return nil
}

// Reports the raw, compressed bytes for a given block
func (b *BlockIterator) BlockBytes() []byte {
	return b.block
}

// Reports the series key for a given block
func (b *BlockIterator) SeriesKey() []byte {
	return b.blockInfo.seriesKey
}

// Reports the min and max time for the current block
func (b *BlockIterator) BlockMinMaxTime() (int64, int64) {
	bnds := b.blockInfo.bounds
	return bnds.Start.Time().UnixNano(), bnds.Stop.Time().UnixNano()
}

// ReadEntry represents information needed to
// copy a block from a remote tsm file to the destination
// TODO: store information about data that needs to be tombstoned
type ReadEntry struct {
	blockData tsm1.IndexEntry // index information for associated block
	seriesKey []byte          // the block's series key
	bounds    execute.Bounds  // the bounds for data included within this block
}

func (t *TSMFilter) filterBlocks(stream io.ReadSeeker, targetOrg, targetBucket *influxdb.ID, bounds execute.Bounds) ([]*ReadEntry, error) {
	if targetOrg == nil {
		return nil, errors.New("must provide org")
	}

	start, end, err := getFileIndexPos(stream)
	if err != nil {
		fmt.Println("error getting file index pos")
		return nil, err
	}

	indexBytes, err := readIndexBytes(stream, start, end)
	if err != nil {
		return nil, err
	}

	idx := tsm1.NewIndirectIndex()
	if err := idx.UnmarshalBinary(indexBytes); err != nil {
		return nil, err
	}

	iter := idx.IteratorFullIndex()

	var blockEntries []*ReadEntry
	for iter.Next() {
		key := iter.Key()
		var a [16]byte
		copy(a[:], key[:16])
		org, bucket := tsdb.DecodeName(a)

		if org == *targetOrg {
			if targetBucket == nil || bucket == *targetBucket {
				var e []tsm1.IndexEntry
				entries, err := idx.ReadEntries(key, e)
				if err != nil {
					return nil, err
				}

				for _, entry := range entries {
					b := execute.Bounds{
						Start: values.ConvertTime(time.Unix(0, entry.MinTime)),
						Stop:  values.ConvertTime(time.Unix(0, entry.MaxTime)),
					}

					// only add the block to our list if there is a non-empty overlap
					if b.Overlaps(bounds) {
						overlapping := b.Intersect(bounds)
						// get the overlap between the bounds we're interested in and the
						// bounds for this particular block
						re := &ReadEntry{
							seriesKey: key,
							blockData: entry,
							bounds:    overlapping,
						}

						blockEntries = append(blockEntries, re)
						t.dataSize += entry.Size - 4
					}
				}
			}
		}
	}
	return blockEntries, nil
}

//DownloadChunk represents information to handle a single s3 download request
type DownloadChunk struct {
	// Start represents the offset of the first block in the chunk in this tsm file
	Start int64
	// End represents the end offset of the last block in the chunk
	End int64
	// Entries contains all the read entries (series key+index info) that are encompassed
	// by this chunk
	Entries []*ReadEntry
	// Buffer represents the internal download buffer for this chunk
	Buffer []byte
}

// segmentToChunks splits up read entries into Download
// chunks based on the position of their blocks in the tsm file,
// to optimize aws read requests
func segmentToChunks(entries []*ReadEntry) []*DownloadChunk {
	var endOffsetToChunk = make(map[int64]*DownloadChunk)

	for _, entry := range entries {
		startOff := entry.blockData.Offset
		endOff := startOff + int64(entry.blockData.Size)

		// if we already have a chunk ending at this start offset...
		if chunk, ok := endOffsetToChunk[startOff]; ok {
			// add this entry to the chunk
			chunk.End = endOff
			chunk.Entries = append(chunk.Entries, entry)

			delete(endOffsetToChunk, startOff)
			endOffsetToChunk[endOff] = chunk
		} else if chunk, ok := endOffsetToChunk[endOff]; ok && startOff == chunk.Start { // handles case where
			// two series keys point to the same block
			chunk.Entries = append(chunk.Entries, entry)
		} else {
			// add a new download chunk
			c := &DownloadChunk{
				Start:   startOff,
				End:     endOff,
				Entries: []*ReadEntry{entry},
			}
			endOffsetToChunk[endOff] = c
		}
	}

	return sortByOffset(endOffsetToChunk)
}

// sorts entries by start offset within the tsm file
// this is necessary, since blocks must be written
// into the resulting tsm file in the same order
// that they were read to maintain ordering
func sortByOffset(m map[int64]*DownloadChunk) []*DownloadChunk {
	offsets := make([]int, 0, len(m))
	chunks := make([]*DownloadChunk, len(m))

	for off := range m {
		offsets = append(offsets, int(off))
	}
	sort.Ints(offsets)

	for i, off := range offsets {
		chunks[i] = m[int64(off)]
	}

	return chunks
}

// Download populates the internal buffer of a DownloadChunk from
// a given read stream
func (c *DownloadChunk) Download(stream io.ReadSeeker) error {
	if _, err := stream.Seek(c.Start, 0); err != nil {
		return err
	}

	bufSiz := c.End - c.Start
	c.Buffer = make([]byte, bufSiz)

	if n, err := stream.Read(c.Buffer); err != nil {
		return err
	} else if int64(n) != bufSiz {
		return errors.New("could not read all block data for chunk")
	}

	return nil
}

// BlockIterator returns an iterator which can be used to iterate over individual
// blocks in a chunk
func (c *DownloadChunk) BlockIterator(stream io.ReadSeeker) (*BlockIterator, error) {
	if err := c.Download(stream); err != nil {
		return nil, err
	}
	return &BlockIterator{
		entries: c.Entries,
		blocks:  c.Buffer,
		offset:  c.Start,
	}, nil
}

// readIndexBytes downloads the index portion of a tsm file
func readIndexBytes(stream io.ReadSeeker, start, end int64) ([]byte, error) {
	if _, err := stream.Seek(start, 0); err != nil {
		return nil, err
	}

	indexSize := end - start
	indexBytes := make([]byte, indexSize)
	n, err := stream.Read(indexBytes)
	if err != nil {
		return nil, err
	} else if int64(n) != indexSize {
		return nil, errors.New("failed to read index")
	}

	return indexBytes, nil
}

// getFileIndexPos reports the start and end offsets of the
// index portion of a tsm file
func getFileIndexPos(stream io.ReadSeeker) (int64, int64, error) {
	var footerStartPos int64
	footerStartPos, err := stream.Seek(-8, 2)
	if err != nil {
		return 0, 0, err
	}

	var footer [8]byte
	if n, err := stream.Read(footer[:]); err != nil {
		return 0, 0, err
	} else if n != 8 {
		return 0, 0, errors.New("failed to read full footer")
	}

	indexStartPos := binary.BigEndian.Uint64(footer[:])
	return int64(indexStartPos), footerStartPos, nil
}
