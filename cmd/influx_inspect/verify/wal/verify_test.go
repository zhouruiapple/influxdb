package wal

import (
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/influxdata/influxdb/tsdb/engine/tsm1"
	"github.com/pkg/errors"
)

const numTestFiles = 1
const numTestEntries = 100

func TestVerifyValidEntries(t *testing.T) {
	test := CreateTest(t, func() (string, error) {
		dir := mustCreateTempDir(t)
		w := tsm1.NewWAL(dir)
		if err := w.Open(); err != nil {
			return "", errors.Wrap(err, "error opening wal")
		}

		for i := 0; i < numTestEntries; i++ {
			writeRandomEntry(w, t)
		}

		if err := w.Close(); err != nil {
			return "", errors.Wrap(err, "error closing wal")
		}

		return dir, nil
	})
	defer test.Close()

	verifier := &verifyWAL{}
	verifier.Run(os.Stdout, test.dir)

	expectedEntries := numTestFiles * numTestEntries
	if verifier.totalEntries != expectedEntries {
		t.Fatalf("Error: expected %d entries, checked only %d entries", expectedEntries, verifier.totalEntries)
	}
}

func TestVerifyCorruptEntries(t *testing.T) {
	test := CreateTest(t, func() (string, error) {
		dir := mustCreateTempDir(t)
		writeCorruptEntry(dir, t)
		return dir, nil
	})

	defer test.Close()

	verifier := &verifyWAL{}
	verifier.Run(os.Stdout, test.dir)
	expectedEntries := 1
	expectedErrors := 1

	if verifier.totalEntries != expectedEntries {
		t.Fatalf("Error: expected %d entries, found %d entries", expectedEntries, verifier.totalEntries)
	}

	if verifier.totalErrors != expectedErrors {
		t.Fatalf("Error: expected %d corrupt entries, found %d corrupt entries", expectedErrors, verifier.totalErrors)
	}
}

type Test struct {
	dir string
}

func CreateTest(t *testing.T, createFiles func() (string, error)) *Test {
	t.Helper()

	dir, err := createFiles()

	if err != nil {
		t.Fatal(err)
	}

	return &Test{
		dir: dir,
	}
}

func writeRandomEntry(w *tsm1.WAL, t *testing.T) {
	if _, err := w.WriteMulti(map[string][]tsm1.Value{
		"cpu,host=A#!~#value": []tsm1.Value{
			tsm1.NewValue(rand.Int63(), rand.Float64()),
		},
	}); err != nil {
		t.Fatalf("error writing points: %v", err)
	}
}

func writeCorruptEntry(walDir string, t *testing.T) {
	f := mustCreateTempFile(t, walDir)
	defer f.Close()
	// random byte sequence
	corrupt := []byte{0, 255, 0, 1, 3, 4, 8, 7}
	f.Write(corrupt)
}

func (t *Test) Close() {
	os.RemoveAll(t.dir)
}

func mustCreateTempDir(t *testing.T) string {
	name, err := ioutil.TempDir(".", "wal-test")
	if err != nil {
		t.Fatal(err)
	}

	return name
}

func mustCreateTempFile(t *testing.T, dir string) *os.File {
	file, err := ioutil.TempFile(dir, "corrupt*.wal")
	if err != nil {
		t.Fatal(err)
	}

	return file
}
