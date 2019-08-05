package influxdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
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

type TSMFilter struct {
	Org, Bucket *influxdb.ID
	Bounds      execute.Bounds
	source      values.Stream

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
	}, nil
}

func (t *TSMFilter) FilteredTSMStream() (io.Reader, error) {
	s, _ := t.source.(io.ReadSeeker)
	entries, err := t.filterBlocks(s, t.Org, t.Bucket, t.Bounds)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		fmt.Println("entry: ", entry.blocksChunkLen)
	}

	buf := &bytes.Buffer{}
	tsmWriter, err := tsm1.NewTSMWriter(buf)
	if err != nil {
		return nil, err
	}

	fmt.Println("len(entries): ", len(entries))
	iter := NewBlockIterator(entries, s)

	var read uint32 = 0
	for iter.HasNext() {
		start := time.Now()
		if err := iter.Next(); err != nil {
			return nil, err
		}

		fmt.Println("time to fetch next block: ", time.Since(start))

		fmt.Printf("\rwrote %d out of total %d; %3.2f%% finished", read, t.dataSize, float32(read)/float32(t.dataSize)*100)

		entryBytes := iter.BlockBytes()
		entryKey := iter.SeriesKey()
		minTime, maxTime := iter.BlockMinMaxTime()

		fmt.Println("len(entryBytes): ", len(entryBytes))
		start = time.Now()
		if err := tsmWriter.WriteBlock(entryKey, minTime, maxTime, entryBytes); err != nil {
			return nil, err
		}
		fmt.Printf("time to write block: %v\n", time.Since(start))

		read += uint32(len(entryBytes))
	}

	if err := tsmWriter.WriteIndex(); err != nil {
		return nil, err
	}

	if err := tsmWriter.Close(); err != nil {
		return nil, err
	}

	return buf, nil
}

type BlockIterator struct {
	entries      []*ReadEntry // contains all relevant series keys and their associated blocks
	entry        *ReadEntry   // current series key/min max time
	nextEntryPos int          // current index into entries

	blockList    [][]byte // array of actual data for all blocks referenced in the given entry
	block        []byte
	blockListIdx int // the block currently referenced by the iterator

	stream io.ReadSeeker // our source for reading block data given read entries
}

func NewBlockIterator(entries []*ReadEntry, stream io.ReadSeeker) *BlockIterator {
	return &BlockIterator{
		entries: entries,
		stream:  stream,
	}
}

func (b *BlockIterator) HasNext() bool {
	return b.nextEntryPos < len(b.entries)
}

func (b *BlockIterator) Next() error {
	// if we're still processing the current
	// list of block data, simply increment the
	// block list index
	if b.blockListIdx < len(b.blockList) {
		fmt.Println("returning early after setting block...")
		b.block = b.blockList[b.blockListIdx]
		b.blockListIdx++
		return nil
	}

	// Otherwise, fetch list of block data for
	// next read entry
	if err := b.FetchBlockList(); err != nil {
		return err
	}

	b.entry = b.entries[b.nextEntryPos]

	// at this point, we should never have an empty block list
	if len(b.blockList) == 0 {
		return errors.New("empty block list")
	}

	// Reset the index into the block list and
	// move to the next read entry
	b.blockListIdx = 0
	b.nextEntryPos++

	return nil
}

func (b *BlockIterator) FetchBlockList() error {
	blockList, err := readBytesForEntry(b.stream, b.entries[b.nextEntryPos])
	if err != nil {
		return err
	}
	b.blockList = blockList

	return nil
}

//func (b *BlockIterator) Tags() []models.Tag {
//	return b.currentTags
//}
//
//func (b *BlockIterator) Values() []tsm1.Value {
//	return b.currentValues
//}

func (b *BlockIterator) BlockBytes() []byte {
	return b.block
}

func (b *BlockIterator) SeriesKey() []byte {
	return b.entry.seriesKey
}

func (b *BlockIterator) BlockMinMaxTime() (int64, int64) {
	bnds := b.entry.bounds
	return bnds.Start.Time().UnixNano(), bnds.Stop.Time().UnixNano()
}

type ReadEntry struct {
	blockData []tsm1.IndexEntry
	seriesKey []byte
	bounds    execute.Bounds

	blocksChunkStart int64
	blocksChunkLen   int64
}

func (t *TSMFilter) filterBlocks(stream io.ReadSeeker, targetOrg, targetBucket *influxdb.ID, bounds execute.Bounds) ([]*ReadEntry, error) {
	if targetOrg == nil {
		return nil, errors.New("must provide org")
	}

	start, end, err := getFileIndexPos(stream)
	if err != nil {
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

				minTime, maxTime := minMaxTime(entries)
				b := execute.Bounds{
					Start: values.ConvertTime(time.Unix(0, minTime)),
					Stop:  values.ConvertTime(time.Unix(0, maxTime)),
				}
				// only add the block to our list if there is a non-empty overlap
				if b.Overlaps(bounds) {
					overlapping := b.Intersect(bounds)
					// get the overlap between the bounds we're interested in and the
					// bounds for this particular block
					re := &ReadEntry{
						seriesKey: key,
						blockData: entries,
						bounds:    overlapping,
					}
					if err := re.setReadRange(); err != nil {
						return nil, err
					}

					blockEntries = append(blockEntries, re)

					for _, entry := range entries {
						t.dataSize += entry.Size
					}
				}
			}
		}
	}

	return blockEntries, nil
}

func minMaxTime(entries []tsm1.IndexEntry) (int64, int64) {
	if len(entries) == 0 {
		return flux.MinTime.Absolute.UnixNano(), flux.MaxTime.Absolute.UnixNano()
	}

	var minTime = entries[0].MinTime
	var maxTime = entries[0].MaxTime
	for _, entry := range entries {
		if entry.MinTime < minTime {
			minTime = entry.MinTime
		}

		if entry.MaxTime > maxTime {
			maxTime = entry.MaxTime
		}
	}

	return minTime, maxTime
}

//func decodeBlockForEntry(stream io.ReadSeeker, entry ReadEntry) ([]tsm1.Value, error) {
//	blockBytes, err := readBytesForEntry(stream, entry)
//	if err != nil {
//		return nil, err
//	}
//	values := []tsm1.Value{}
//	if values, err = tsm1.DecodeBlock(blockBytes[4:], values); err != nil {
//		return nil, err
//	}
//
//	return values, nil
//}

func readBytesForEntry(stream io.ReadSeeker, entry *ReadEntry) ([][]byte, error) {
	if entry == nil {
		return nil, nil
	}
	//
	//for _, entry := range entry.blockData {
	//	fmt.Println("entry offset: ", entry.Offset)
	//	fmt.Println("entry offset + entry size: ", entry.Offset+int64(entry.Size))
	//}
	if _, err := stream.Seek(entry.blocksChunkStart, 0); err != nil {
		return nil, err
	}

	var blockBytes = make([]byte, entry.blocksChunkLen)
	start := time.Now()
	n, err := stream.Read(blockBytes)
	fmt.Printf("read took: %v\n", time.Since(start))
	if err != nil {
		return nil, err
	} else if n != int(entry.blocksChunkLen) {
		return nil, errors.New("could not read full block")
	}

	blockList := make([][]byte, len(entry.blockData))

	fmt.Println("splitting block start")
	start = time.Now()
	for i, idxEntry := range entry.blockData {

		blockStart := idxEntry.Offset - entry.blocksChunkStart
		blockEnd := blockStart + int64(idxEntry.Size)
		blockData := blockBytes[blockStart:blockEnd]

		if len(blockData) < 4 {
			return nil, errors.New("block too short")
		}

		sum := blockData[:4]
		blockData = blockData[4:]

		if err := verifyChecksum(sum, blockData); err != nil {
			return nil, err
		}

		blockList[i] = blockData
	}
	fmt.Printf("splitting blocks took: %v\n", time.Since(start))

	//if len(blockBytes) < 4 {
	//	return nil, errors.New("block too short to read")
	//}
	//
	//oldSum := blockBytes[:4]
	//blockBytes = blockBytes[4:]
	//if err := verifyChecksum(oldSum, blockBytes); err != nil {
	//	return nil, err
	//}

	return blockList, nil
}

func (entry *ReadEntry) setReadRange() error {
	if len(entry.blockData) < 1 {
		return fmt.Errorf("no blocks for key: %s", entry.seriesKey)
	}
	startEntry, endEntry := entry.blockData[0], entry.blockData[len(entry.blockData)-1]

	entry.blocksChunkStart = startEntry.Offset
	entry.blocksChunkLen = endEntry.Offset + int64(endEntry.Size)

	return nil
}

func verifyChecksum(want []byte, data []byte) error {
	var checksum [crc32.Size]byte
	binary.BigEndian.PutUint32(checksum[:], crc32.ChecksumIEEE(data))

	if bytes.Compare(want, checksum[:]) != 0 {
		return errors.New("invalid checksum for block")
	}

	return nil
}

func readIndexBytes(stream io.ReadSeeker, start, end int64) ([]byte, error) {
	if _, err := stream.Seek(start, 0); err != nil {
		return nil, err
	}

	fmt.Printf("start: %d; end: %d\n", start, end)

	indexSize := end - start

	fmt.Println("indexSize is: ", indexSize)

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

	fmt.Println("footer start pos: ", footerStartPos)

	var footer [8]byte
	if n, err := stream.Read(footer[:]); err != nil {
		return 0, 0, err
	} else if n != 8 {
		return 0, 0, errors.New("failed to read full footer")
	}

	indexStartPos := binary.BigEndian.Uint64(footer[:])
	return int64(indexStartPos), footerStartPos, nil
}

//type IndexIterator struct {
//	buffer []byte
//	offset int
//
//	currentEntry *tsm1.IndexEntry
//	currentKey   []byte
//}
//
//func NewIndexIterator(b []byte) *IndexIterator {
//	return &IndexIterator{
//		buffer: b,
//	}
//}
//
//func (iter *IndexIterator) HasNext() bool {
//	fmt.Println("HasNext(): ")
//	fmt.Println("offset is: ", iter.offset)
//	fmt.Println("buffer is: ", len(iter.buffer))
//	return iter.offset < len(iter.buffer)
//}
//
//func (iter *IndexIterator) Key() []byte {
//	return iter.currentKey
//}
//
//func (iter *IndexIterator) Entry() *tsm1.IndexEntry {
//	return iter.currentEntry
//}
//
//func (iter *IndexIterator) Next() error {
//	keyLenBytes := iter.buffer[iter.offset : iter.offset+2]
//
//	keyLen := binary.BigEndian.Uint16(keyLenBytes)
//	iter.offset += 2
//	fmt.Println("iter.offset: ", iter.offset)
//	fmt.Println("len(iter.buffer): ", len(iter.buffer))
//	fmt.Println("keyLen: ", int(keyLen))
//	seriesKey := iter.buffer[iter.offset : iter.offset+int(keyLen)]
//
//	iter.currentKey = seriesKey
//	iter.offset += int(keyLen) + 3
//
//	entry := &tsm1.IndexEntry{}
//	if err := entry.UnmarshalBinary(iter.buffer[iter.offset : iter.offset+28]); err != nil {
//		return err
//	}
//
//	iter.offset += 28
//
//	iter.currentEntry = entry
//	iter.currentKey = seriesKey
//
//	return nil
//}
