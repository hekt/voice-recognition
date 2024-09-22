package file

import (
	"fmt"
	"io"
	"os"
)

// OpenCloseFileWriter はファイルへの書き込みを行う構造体。
// io.Writer インターフェースを実装している。
// Write() メソッドが実行されるたびにファイルを開き、書き込みを行う。
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
