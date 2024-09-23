package file

import (
	"fmt"
	"io"
	"os"
)

// OpenCloseFileWrite is a impelmentation of io.Writer.
// It opens the file and writes to it each time the Write() method is executed.
type OpenCloseFileWriter struct {
	path string
	flag int
	perm os.FileMode
}

var _ io.Writer = (*OpenCloseFileWriter)(nil)

func NewOpenCloseFileWriter(path string, flag int, perm os.FileMode) *OpenCloseFileWriter {
	return &OpenCloseFileWriter{path, flag, perm}
}

func (w *OpenCloseFileWriter) Write(p []byte) (int, error) {
	file, err := os.OpenFile(w.path, w.flag, w.perm)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	n, err := file.Write(p)
	if err != nil {
		return n, fmt.Errorf("failed to write to file: %w", err)
	}

	return n, nil
}
