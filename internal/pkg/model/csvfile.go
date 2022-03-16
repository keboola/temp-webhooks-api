package model

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type CsvFile struct {
	lock   *sync.Mutex
	file   *os.File
	writer *csv.Writer
	rows   int
}

func NewCsvFile(header []string) (*CsvFile, error) {
	file, err := ioutil.TempFile(os.TempDir(), "keboola-csv")
	if err != nil {
		return nil, err
	}

	// Create file
	csvFile := &CsvFile{
		lock:   &sync.Mutex{},
		file:   file,
		writer: csv.NewWriter(file),
	}

	// Write header
	if err := csvFile.writer.Write(header); err != nil {
		return nil, err
	}

	return csvFile, nil
}

func (f *CsvFile) Rows() int {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.rows
}

func (f *CsvFile) Size() (int64, error) {
	info, err := f.file.Stat()
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

func (f *CsvFile) Write(record []string) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	if err := f.writer.Write(record); err != nil {
		return fmt.Errorf("cannot write to CSV file: %w", err)
	}

	f.rows++
	return nil
}
