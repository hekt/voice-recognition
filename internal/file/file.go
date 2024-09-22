package file

import (
	"fmt"
	"io"
	"os"
)

// FileWriter はファイルへの書き込みを行う構造体。
// io.Writer インターフェースを実装している。
// Write() メソッドが実行されるたびにファイルを開き、書き込みを行う。
type FileWriter struct {
	path string
	flag int
	perm os.FileMode
}

var _ io.Writer = (*FileWriter)(nil)

func NewFileWriter(path string, flag int, perm os.FileMode) *FileWriter {
	return &FileWriter{path, flag, perm}
}

func (w *FileWriter) Write(p []byte) (int, error) {
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
