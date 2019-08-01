package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/influxdata/flux/values"

	"github.com/influxdata/flux/execute"

	influx_stdlib "github.com/influxdata/influxdb/query/stdlib/influxdata/influxdb"

	"os"
)

var bucket string
var org string
var filename string
var outfile string

func main() {
	flag.StringVar(&filename, "file", "test.tsm", "test tsm file")
	flag.StringVar(&org, "org", "", "organization for filtering")
	flag.StringVar(&bucket, "bucket", "", "bucket for filtering")
	flag.StringVar(&outfile, "out", "out.tsm", "optional name for output .tsm file")
	flag.Parse()

	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open test file")
		return
	}
	fmt.Println("filename ", filename)
	info, _ := f.Stat()
	fmt.Println("tsm file size: ", info.Size())

	//orgID, err := influxdb.IDFromString(org)
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "invalid org id")
	//	return
	//}
	//
	//var bucketID *influxdb.ID
	//if bucket != "" {
	//	bucketID, err = influxdb.IDFromString(bucket)
	//	if err != nil {
	//		fmt.Fprintf(os.Stderr, "invalid bucket id")
	//		return
	//	}
	//} else {
	//	i := influxdb.ID(0)
	//	bucketID = &i
	//}

	//if entries, err := processFile(f, *orgID, *bucketID); err != nil {
	//	fmt.Fprintf(os.Stderr, err.Error())
	//} else if err := printBlockData(f, entries); err != nil {
	//	fmt.Fprintf(os.Stderr, err.Error())
	//}
	tsmFilter, err := influx_stdlib.NewTSMFilter(org, bucket, execute.AllTime, values.NewStream(f))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error filtering tsm: %v\n", err)
		return
	}

	out, err := tsmFilter.FilteredTSMStream()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
	outBytes, err := ioutil.ReadAll(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}

	if err := ioutil.WriteFile(outfile, outBytes, 0644); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
}

//type ReadEntry struct {
//	blockData *tsm1.IndexEntry
//	seriesKey []byte
//}
//
//func processFile(stream io.ReadSeeker, targetOrg, targetBucket influxdb.ID) ([]ReadEntry, error) {
//	start, end, err := getFileIndexPos(stream)
//	if err != nil {
//		return nil, err
//	}
//
//	indexBytes, err := readIndexBytes(stream, start, end)
//	if err != nil {
//		return nil, err
//	}
//
//	iter := NewIndexIterator(indexBytes)
//
//	var blockEntries []ReadEntry
//	for iter.HasNext() {
//		if err := iter.Next(); err != nil {
//			return nil, err
//		}
//
//		key := iter.Key()
//		entry := iter.Entry()
//
//		var a [16]byte
//		copy(a[:], key[:16])
//		org, bucket := tsdb.DecodeName(a)
//
//		// filtering
//		fmt.Printf("org: %s, bucket %s\n", org.String(), bucket.String())
//		fmt.Printf("key: %s, entry: %v\n", string(key), entry)
//
//		if org == targetOrg {
//			if targetBucket == 0 || bucket == targetBucket {
//				blockEntries = append(blockEntries, ReadEntry{
//					seriesKey: key[16:],
//					blockData: entry,
//				})
//			}
//		}
//	}
//
//	return blockEntries, nil
//}
//
//func decodeBlockForEntry(stream io.ReadSeeker, entry ReadEntry) ([]tsm1.Value, error) {
//	if _, err := stream.Seek(entry.blockData.Offset, 0); err != nil {
//		return nil, err
//	}
//
//	var blockBytes = make([]byte, entry.blockData.Size)
//	n, err := stream.Read(blockBytes)
//	if err != nil {
//		return nil, err
//	} else if n != int(entry.blockData.Size) {
//		return nil, errors.New("could not read full block")
//	}
//
//	values := []tsm1.Value{}
//	if len(blockBytes) < 4 {
//		return nil, errors.New("block too short to read")
//	}
//	if values, err = tsm1.DecodeBlock(blockBytes[4:], values); err != nil {
//		return nil, err
//	}
//
//	return values, nil
//}
//
//func printBlockData(stream io.ReadSeeker, entries []ReadEntry) error {
//	for _, entry := range entries {
//		values, err := decodeBlockForEntry(stream, entry)
//		if err != nil {
//			return err
//		}
//
//		fmt.Println(values)
//	}
//
//	return nil
//}
//
//func readIndexBytes(stream io.ReadSeeker, start, end int64) ([]byte, error) {
//	if _, err := stream.Seek(start, 0); err != nil {
//		return nil, err
//	}
//
//	indexSize := end - start
//	indexBytes := make([]byte, indexSize)
//
//	n, err := stream.Read(indexBytes)
//	if err != nil {
//		return nil, err
//	} else if int64(n) != indexSize {
//		return nil, errors.New("failed to read index")
//	}
//
//	return indexBytes, nil
//}
//
//func getFileIndexPos(stream io.ReadSeeker) (int64, int64, error) {
//	var footerStartPos int64
//	footerStartPos, err := stream.Seek(-8, 2)
//	if err != nil {
//		return 0, 0, err
//	}
//
//	fmt.Println("footerStartPos: ", footerStartPos)
//
//	var footer [8]byte
//	if n, err := stream.Read(footer[:]); err != nil {
//		return 0, 0, err
//	} else if n != 8 {
//		return 0, 0, errors.New("failed to read full footer")
//	}
//
//	indexStartPos := binary.BigEndian.Uint64(footer[:])
//	fmt.Println("read footer: ", indexStartPos)
//
//	return int64(indexStartPos), footerStartPos, nil
//}
//
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
