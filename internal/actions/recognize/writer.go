package recognize

import (
	"bytes"
	"io"
)

var (
	clearScreen = []byte("\033[H\033[2J")
	greenColor  = []byte("\033[32m")
	resetColor  = []byte("\033[0m")
	newLine     = []byte("\n")
)

var _ io.Writer = (*DecoratedInterimWriter)(nil)

type DecoratedInterimWriter struct {
	Writer io.Writer
}

func (w *DecoratedInterimWriter) Write(p []byte) (n int, err error) {
	buf := bytes.Buffer{}
	buf.Write(clearScreen)
	buf.Write(greenColor)
	buf.Write(p)
	buf.Write(resetColor)

	return w.Writer.Write(buf.Bytes())
}

var _ io.Writer = (*DecoratedResultWriter)(nil)

type DecoratedResultWriter struct {
	Writer io.Writer
}

func (w *DecoratedResultWriter) Write(p []byte) (n int, err error) {
	buf := bytes.NewBuffer(newLine)
	buf.Write(p)

	return w.Writer.Write(buf.Bytes())
}
