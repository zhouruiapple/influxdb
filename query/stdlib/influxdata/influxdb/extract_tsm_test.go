package influxdb

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/influxdata/influxdb/tsdb/tsm1"
)

func TestSegmentToChunks(t *testing.T) {
	tests := []struct {
		name    string
		entries []*ReadEntry
		want    []*DownloadChunk
	}{
		{
			name: "contiguous",
			entries: []*ReadEntry{
				{SeriesKey: []byte("key0"), BlockData: tsm1.IndexEntry{Offset: 100, Size: 200}},
				{SeriesKey: []byte("key1"), BlockData: tsm1.IndexEntry{Offset: 300, Size: 50}},
				{SeriesKey: []byte("key2"), BlockData: tsm1.IndexEntry{Offset: 350, Size: 20}},
			},
			want: []*DownloadChunk{
				{
					Start: 100,
					End:   370,
					Entries: []*ReadEntry{
						{SeriesKey: []byte("key0"), BlockData: tsm1.IndexEntry{Offset: 100, Size: 200}},
						{SeriesKey: []byte("key1"), BlockData: tsm1.IndexEntry{Offset: 300, Size: 50}},
						{SeriesKey: []byte("key2"), BlockData: tsm1.IndexEntry{Offset: 350, Size: 20}},
					},
				},
			},
		},
		{
			name: "two chunks",
			entries: []*ReadEntry{
				{SeriesKey: []byte("key0"), BlockData: tsm1.IndexEntry{Offset: 100, Size: 200}},
				{SeriesKey: []byte("key1"), BlockData: tsm1.IndexEntry{Offset: 300, Size: 50}},
				{SeriesKey: []byte("key2"), BlockData: tsm1.IndexEntry{Offset: 1024, Size: 76}},
				{SeriesKey: []byte("key3"), BlockData: tsm1.IndexEntry{Offset: 1100, Size: 500}},
			},
			want: []*DownloadChunk{
				{
					Start: 100,
					End:   350,
					Entries: []*ReadEntry{
						{SeriesKey: []byte("key0"), BlockData: tsm1.IndexEntry{Offset: 100, Size: 200}},
						{SeriesKey: []byte("key1"), BlockData: tsm1.IndexEntry{Offset: 300, Size: 50}},
					},
				},
				{
					Start: 1024,
					End:   1600,
					Entries: []*ReadEntry{
						{SeriesKey: []byte("key2"), BlockData: tsm1.IndexEntry{Offset: 1024, Size: 76}},
						{SeriesKey: []byte("key3"), BlockData: tsm1.IndexEntry{Offset: 1100, Size: 500}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SegmentToChunks(tt.entries)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

//func TestFilterTSMFile(t *testing.T) {
//	f, err := ioutil.TempFile(".", "test")
//	w, err := tsm1.NewTSMWriter(f)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	orgID, _ := influxdb.IDFromString("0434394f8ee69000")
//	bucketID1, _ := influxdb.IDFromString("1111111111111111")
//	bucketID2, _ := influxdb.IDFromString("2222222222222222")
//	orgBytes, _ := orgID.Encode()
//	bucketID1Bytes, _ := bucketID1.Encode()
//	m := make([]byte, 16)
//	copy(orgBytes, m[:8])
//	copy(bucketID1Bytes, m[8:])
//
//	var data = []struct {
//		key    string
//		values []tsm1.Value
//	}{
//		{string(m) + "t=tag1", []tsm1.Value{
//			tsm1.NewValue(0, 1.0),
//			tsm1.NewValue(1, 2.0)},
//		},
//		{string(m) + "t=tag2", []tsm1.Value{
//			tsm1.NewValue(2, 3.0),
//			tsm1.NewValue(3, 4.0)},
//		},
//	}
//
//	for _, pt := range data {
//		if err := w.Write([]byte(pt.key), pt.values); err != nil {
//			t.Fatal(err)
//		}
//	}
//
//	bnds := execute.AllTime
//
//	filter, err := NewTSMFilter("0434394f8ee69000", "", bnds, values.NewReadSeekStream(f))
//
//}
