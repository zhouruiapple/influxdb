package influxdb

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

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
	Bounds      flux.Bounds
	source      values.Stream
}

func NewTSMFilter(org, bucket string, bounds flux.Bounds, src values.Stream) (*TSMFilter, error) {
	orgID, err := influxdb.IDFromString(org)
	if err != nil {
		return nil, err
	}

	bucketID, err := influxdb.IDFromString(bucket)
	if err != nil {
		return nil, err
	}

	return &TSMFilter{
		Org:    orgID,
		Bucket: bucketID,
		Bounds: bounds,
		source: src,
	}, nil
}

func (t *TSMFilter) FilteredTSMStream(filter TSMFilter) (values.Stream, error) {
	entries, err := filterBlocks(t.source, t.Org, t.Bucket)
	if err != nil {
		return nil, err
	}

	iter := NewBlockIterator(entries, t.source)

	for iter.HasNext() {
		if err := iter.Next(); err != nil {
			return nil, err
		}

		tags := iter.Tags()
		vals := iter.Values()
	}

	return nil
}

func init() {
	flux.RegisterPackageValue("influxdata/influxdb", ExtractTSMKind, extractTSM)
}

type BlockIterator struct {
	entries          []ReadEntry
	currentBlockInfo ReadEntry
	currentTags      []models.Tag
	currentValues    []tsm1.Value
	pos              int
	stream           io.ReadSeeker
}

func NewBlockIterator(entries []ReadEntry, stream io.ReadSeeker) *BlockIterator {
	return &BlockIterator{
		entries: entries,
		stream:  stream,
	}
}

func (b *BlockIterator) HasNext() bool {
	return b.pos < len(b.entries)
}

func (b *BlockIterator) Next() error {
	b.currentBlockInfo = b.entries[b.pos]
	vals, err := decodeBlockForEntry(b.stream, b.currentBlockInfo)
	if err != nil {
		return err
	}
	_, tags := tsdb.ParseSeriesKey(b.currentBlockInfo.seriesKey)

	b.currentTags = tags
	b.currentValues = vals

	return nil
}

func (b *BlockIterator) Tags() []models.Tag {
	return b.currentTags
}

func (b *BlockIterator) Values() []tsm1.Value {
	return b.currentValues
}

type ReadEntry struct {
	blockData *tsm1.IndexEntry
	seriesKey []byte
}

func filterBlocks(stream io.ReadSeeker, targetOrg, targetBucket influxdb.ID) ([]ReadEntry, error) {
	start, end, err := getFileIndexPos(stream)
	if err != nil {
		return nil, err
	}

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

		if org == targetOrg {
			if targetBucket == 0 || bucket == targetBucket {
				blockEntries = append(blockEntries, ReadEntry{
					seriesKey: key[16:],
					blockData: entry,
				})
			}
		}
	}

	return blockEntries, nil
}

func decodeBlockForEntry(stream io.ReadSeeker, entry ReadEntry) ([]tsm1.Value, error) {
	if _, err := stream.Seek(entry.blockData.Offset, 0); err != nil {
		return nil, err
	}

	var blockBytes = make([]byte, entry.blockData.Size)
	n, err := stream.Read(blockBytes)
	if err != nil {
		return nil, err
	} else if n != int(entry.blockData.Size) {
		return nil, errors.New("could not read full block")
	}

	values := []tsm1.Value{}
	if len(blockBytes) < 4 {
		return nil, errors.New("block too short to read")
	}
	if values, err = tsm1.DecodeBlock(blockBytes[4:], values); err != nil {
		return nil, err
	}

	return values, nil
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
