package wal

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/influxdata/influxdb/tsdb/engine/tsm1"
	"github.com/pkg/errors"
)

type verifyWAL struct {
	files        []string
	currentFile  string
	err          error
	start        time.Time
	totalErrors  int
	totalEntries int
}

// add all files with the .wal extension to the current struct
func (v *verifyWAL) loadFiles(path string) error {
	fmt.Println("file path: ", path)
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == "."+tsm1.WALFileExtension {
			v.files = append(v.files, path)
		}

		return nil
	})

	if err != nil {
		return errors.Wrap(err, "failed to read wal files")
	}

	return nil
}

func (v *verifyWAL) WALReader() *tsm1.WALSegmentReader {
	file, err := os.OpenFile(v.currentFile, os.O_RDONLY, 0600)
	if err != nil {
		return nil
	}

	return tsm1.NewWALSegmentReader(file)
}

func (v *verifyWAL) Start() {
	v.start = time.Now()
}

func (v *verifyWAL) Elapsed() time.Duration {
	return time.Since(v.start)
}

func (v *verifyWAL) Next() bool {
	if len(v.files) == 0 {
		return false
	}

	v.currentFile = v.files[0]
	v.files = v.files[1:]

	return true
}

func (v *verifyWAL) Run(w io.Writer, dataPath string) error {
	if err := v.loadFiles(dataPath); err != nil {
		return err
	}

	log.Println("v.files: ", v.files)

	// iterate through all discovered wal files. Create a reader for each file, and
	// attempt to parse entries to verify their integrity.
	// if there are any unmarshalling errors, increase the error counts
	// and log a relevant message
	v.Start()
	for v.Next() {
		reader := v.WALReader()
		fileErrors := 0
		fileEntries := 0
		for reader.Next() {
			entry, err := reader.Read()
			if err != nil {
				fmt.Fprintf(w, "%s: corrupt wal entry at position %d; %v\n", v.currentFile, reader.Count(), err)
				fileErrors++
				v.totalErrors++
			} else {
				fmt.Println("entry of type: ", entry.Type())
			}
			fileEntries++
			v.totalEntries++
		}

		if fileErrors == 0 {
			fmt.Fprintf(w, "%s: healthy\n", v.currentFile)
		} else {
			fmt.Fprintf(w, "%s: %d corrupt entries found\n", v.currentFile, fileErrors)
		}

		reader.Close()
	}

	fmt.Fprintf(w, "Corrupt entries: %d; Total entries: %d; Time elapsed: %f", v.totalErrors, v.totalEntries, v.Elapsed().Seconds())

	return nil
}
