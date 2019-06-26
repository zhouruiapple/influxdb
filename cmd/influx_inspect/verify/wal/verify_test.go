package wal

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/golang/snappy"
	"github.com/influxdata/influxdb/tsdb/engine/tsm1"
)

const numTestFiles = 1
const numTestEntries = 100

func TestReadValidEntries(t *testing.T) {
	test := CreateTest(t)
	defer test.Close()

	verifier := &verifyWAL{}
	verifier.Run(os.Stdout, test.dir)

	expectedEntries := numTestFiles * numTestEntries
	if verifier.totalEntries != expectedEntries {
		t.Fatalf("Error: expected %d entries, checked only %d entries", expectedEntries, verifier.totalEntries)
	}
}

type Test struct {
	dir string
}

func CreateTest(t *testing.T) *Test {
	t.Helper()

	dir, err := ioutil.TempDir(".", "wal")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(dir)
	f, err := ioutil.TempFile(dir, "test.wal")
	if err != nil {
		t.Fatal(err)
	}

	w := tsm1.NewWALSegmentWriter(f)

	p1 := tsm1.NewValue(1, int64(1))
	p2 := tsm1.NewValue(1, int64(2))

	exp := []struct {
		key    string
		values []tsm1.Value
	}{
		{"cpu,host=A#!~#value", []tsm1.Value{p1}},
		{"cpu,host=B#!~#value", []tsm1.Value{p2}},
	}

	for _, v := range exp {
		entry := &tsm1.WriteWALEntry{
			Values: map[string][]tsm1.Value{v.key: v.values},
		}

		if err := w.Write(mustMarshalEntry(entry)); err != nil {
			t.Fatal("failed to write points", err)
		}
		if err := w.Flush(); err != nil {
			t.Fatal("Failed to flush points", err)
		}
	}

	return &Test{
		dir: dir,
	}
}

// note: helper function for writing WAL data copied from internal WAL tests
func mustMarshalEntry(entry tsm1.WALEntry) (tsm1.WalEntryType, []byte) {
	bytes := make([]byte, 1024<<2)

	b, err := entry.Encode(bytes)
	if err != nil {
		panic(fmt.Sprintf("error encoding: %v", err))
	}

	return entry.Type(), snappy.Encode(b, b)
}

func (t *Test) Close() {
	os.RemoveAll(t.dir)
}
