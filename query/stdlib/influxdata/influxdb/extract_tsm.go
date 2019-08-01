package influxdb

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/influxdata/flux/execute"

	"github.com/influxdata/influxdb/models"

	"github.com/influxdata/influxdb/tsdb"
	"github.com/influxdata/influxdb/tsdb/tsm1"

	"github.com/influxdata/influxdb"

	"github.com/influxdata/flux"
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
		return nil, nil
	}, false,
)

type TSMFilter struct {
	Org, Bucket *influxdb.ID
	Bounds      execute.Bounds
	source      values.Stream
	pipeReader  *io.PipeReader
	pipeWriter  *io.PipeWriter
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

	reader, writer := io.Pipe()

	return &TSMFilter{
		Org:        orgID,
		Bucket:     bucketID,
		Bounds:     bounds,
		source:     src,
		pipeReader: reader,
		pipeWriter: writer,
	}, nil
}

func (t *TSMFilter) FilteredTSMStream() (io.Reader, error) {
	entries, err := filterBlocks(t.source, t.Org, t.Bucket, t.Bounds)
	if err != nil {
		return nil, err
	}

	fmt.Println("entries: ", entries)

	fmt.Println("filtered blocks...")

	tsmWriter, err := tsm1.NewTSMWriter(t.pipeWriter)
	if err != nil {
		return nil, err
	}
	fmt.Println("created tsm writer...")
	iter := NewBlockIterator(entries, t.source)

	fmt.Println("created block iterator...")

	for iter.HasNext() {
		fmt.Println("about to get next")
		if err := iter.Next(); err != nil {
			return nil, err
		}

		fmt.Println("got next block...")

		entryBytes := iter.BlockBytes()
		entryKey := iter.SeriesKey()
		minTime, maxTime := iter.BlockMinMaxTime()

		if err := tsmWriter.WriteBlock(entryKey, minTime, maxTime, entryBytes); err != nil {
			return nil, err
		}
		fmt.Println("wrote block")
	}

	if err := tsmWriter.WriteIndex(); err != nil {
		return nil, err
	}

	tsmWriter.Close()

	return t.pipeReader, nil
}

func init() {
	flux.RegisterPackageValue("influxdata/influxdb", ExtractTSMKind, extractTSM)
}

type BlockIterator struct {
	entries           []ReadEntry
	currentBlockInfo  ReadEntry
	currentTags       []models.Tag
	currentValues     []tsm1.Value
	currentBlockBytes []byte
	pos               int
	stream            io.ReadSeeker
}

func NewBlockIterator(entries []ReadEntry, stream io.ReadSeeker) *BlockIterator {
	return &BlockIterator{
		entries: entries,
		stream:  stream,
	}
}

func (b *BlockIterator) HasNext() bool {
	fmt.Println("HasNext()")
	return b.pos < len(b.entries)
}

func (b *BlockIterator) Next() error {
	fmt.Println("Next()")
	b.currentBlockInfo = b.entries[b.pos]
	//vals, err := decodeBlockForEntry(b.stream, b.currentBlockInfo)
	blockBytes, err := readBytesForEntry(b.stream, b.currentBlockInfo)
	if err != nil {
		return err
	}
	fmt.Println("read bytes for entry...")
	//_, tags := tsdb.ParseSeriesKey(b.currentBlockInfo.seriesKey)

	//b.currentTags = tags
	//b.currentValues = vals
	b.currentBlockBytes = blockBytes
	b.pos++

	return nil
}

func (b *BlockIterator) Tags() []models.Tag {
	return b.currentTags
}

func (b *BlockIterator) Values() []tsm1.Value {
	return b.currentValues
}

func (b *BlockIterator) BlockBytes() []byte {
	return b.currentBlockBytes
}

func (b *BlockIterator) SeriesKey() []byte {
	return b.currentBlockInfo.seriesKey
}

func (b *BlockIterator) BlockMinMaxTime() (int64, int64) {
	return b.currentBlockInfo.blockData.MinTime, b.currentBlockInfo.blockData.MaxTime
}

type ReadEntry struct {
	blockData *tsm1.IndexEntry
	seriesKey []byte
	bounds    execute.Bounds
}

func filterBlocks(stream io.ReadSeeker, targetOrg, targetBucket *influxdb.ID, bounds execute.Bounds) ([]ReadEntry, error) {
	fmt.Println("filterBlocks()")
	if targetOrg == nil {
		return nil, errors.New("must provide org")
	}

	start, end, err := getFileIndexPos(stream)
	if err != nil {
		return nil, err
	}

	fmt.Println("got file index pos")

	indexBytes, err := readIndexBytes(stream, start, end)
	if err != nil {
		return nil, err
	}

	iter := NewIndexIterator(indexBytes)

	var blockEntries []ReadEntry
	for iter.HasNext() {
		if err := iter.Next(); err != nil {
			return nil, err
		}

		key := iter.Key()
		entry := iter.Entry()

		var a [16]byte
		copy(a[:], key[:16])
		org, bucket := tsdb.DecodeName(a)

		// filtering
		fmt.Printf("org: %s, bucket %s\n", org.String(), bucket.String())
		fmt.Printf("key: %s, entry: %v\n", string(key), entry)

		if org == *targetOrg {
			if targetBucket == nil || bucket == *targetBucket {
				b := execute.Bounds{
					Start: values.ConvertTime(time.Unix(0, entry.MinTime)),
					Stop:  values.ConvertTime(time.Unix(0, entry.MaxTime)),
				}

				// only add the block to our list if there is a non-empty overlap
				if b.Overlaps(bounds) {
					overlapping := b.Intersect(bounds)
					fmt.Println("bounds for block: ", b)

					// get the overlap between the bounds we're interested in and the
					// bounds for this particular block

					fmt.Println("b.Intersect(bounds): ", overlapping)
					fmt.Println("overlapping.IsEmpty(): ", overlapping.IsEmpty())
					blockEntries = append(blockEntries, ReadEntry{
						seriesKey: key[16:],
						blockData: entry,
						bounds:    overlapping,
					})
				}
			}
		}
	}

	return blockEntries, nil
}

func decodeBlockForEntry(stream io.ReadSeeker, entry ReadEntry) ([]tsm1.Value, error) {
	blockBytes, err := readBytesForEntry(stream, entry)
	if err != nil {
		return nil, err
	}
	values := []tsm1.Value{}
	if values, err = tsm1.DecodeBlock(blockBytes[4:], values); err != nil {
		return nil, err
	}

	return values, nil
}

func readBytesForEntry(stream io.ReadSeeker, entry ReadEntry) ([]byte, error) {
	fmt.Println("readBytesForEntry() start")
	if _, err := stream.Seek(entry.blockData.Offset, 0); err != nil {
		return nil, err
	}

	fmt.Println("seek succeeded")

	var blockBytes = make([]byte, entry.blockData.Size)
	n, err := stream.Read(blockBytes)
	if err != nil {
		return nil, err
	} else if n != int(entry.blockData.Size) {
		return nil, errors.New("could not read full block")
	}

	if len(blockBytes) < 4 {
		return nil, errors.New("block too short to read")
	}

	fmt.Println("readBytesForEntry() end")

	return blockBytes, nil
}

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

func getFileIndexPos(stream io.ReadSeeker) (int64, int64, error) {
	var footerStartPos int64
	footerStartPos, err := stream.Seek(-8, 2)
	if err != nil {
		return 0, 0, err
	}

	fmt.Println("footerStartPos: ", footerStartPos)

	var footer [8]byte
	if n, err := stream.Read(footer[:]); err != nil {
		return 0, 0, err
	} else if n != 8 {
		return 0, 0, errors.New("failed to read full footer")
	}

	indexStartPos := binary.BigEndian.Uint64(footer[:])
	fmt.Println("read footer: ", indexStartPos)

	return int64(indexStartPos), footerStartPos, nil
}

type IndexIterator struct {
	buffer []byte
	offset int

	currentEntry *tsm1.IndexEntry
	currentKey   []byte
}

func NewIndexIterator(b []byte) *IndexIterator {
	return &IndexIterator{
		buffer: b,
	}
}

func (iter *IndexIterator) HasNext() bool {
	return iter.offset < len(iter.buffer)
}

func (iter *IndexIterator) Key() []byte {
	return iter.currentKey
}

func (iter *IndexIterator) Entry() *tsm1.IndexEntry {
	return iter.currentEntry
}

func (iter *IndexIterator) Next() error {
	keyLenBytes := iter.buffer[iter.offset : iter.offset+2]

	keyLen := binary.BigEndian.Uint16(keyLenBytes)
	iter.offset += 2
	seriesKey := iter.buffer[iter.offset : iter.offset+int(keyLen)]

	iter.currentKey = seriesKey
	iter.offset += int(keyLen) + 3

	entry := &tsm1.IndexEntry{}
	if err := entry.UnmarshalBinary(iter.buffer[iter.offset : iter.offset+28]); err != nil {
		return err
	}

	iter.offset += 28

	iter.currentEntry = entry
	iter.currentKey = seriesKey

	return nil
}
