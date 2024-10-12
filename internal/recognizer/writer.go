package recognizer

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
	buf    bytes.Buffer
}

func (w *DecoratedInterimWriter) Write(p []byte) (n int, err error) {
	w.buf.Reset()
	w.buf.Write(clearScreen)
	w.buf.Write(greenColor)
	w.buf.Write(p)
	w.buf.Write(resetColor)

	return w.Writer.Write(w.buf.Bytes())
}

var _ io.Writer = (*DecoratedResultWriter)(nil)

type DecoratedResultWriter struct {
	Writer io.Writer
	buf    bytes.Buffer
}

func (w *DecoratedResultWriter) Write(p []byte) (n int, err error) {
	w.buf.Reset()
	w.buf.Write(newLine)
	w.buf.Write(p)

	return w.Writer.Write(w.buf.Bytes())
}

var _ io.Writer = (*NotifyingWriter)(nil)

type NotifyingWriter struct {
	Writer   io.Writer
	NotifyCh chan<- struct{}
}

func (w *NotifyingWriter) Write(p []byte) (n int, err error) {
	defer func() {
		w.NotifyCh <- struct{}{}
	}()
	return w.Writer.Write(p)
}
