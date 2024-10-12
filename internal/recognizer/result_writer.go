package recognizer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/hekt/voice-recognition/internal/recognizer/model"
)

//go:generate moq -rm -out result_writer_mock.go . ResultWriterInterface
type ResultWriterInterface interface {
	Start(ctx context.Context) error
}

var _ ProcessMonitorInterface = &ProcessMonitor{}

type ResultWriter struct {
	resultCh      <-chan []*model.Result
	resultWriter  io.Writer
	interimWriter io.Writer
}

func NewResultWriter(
	resultCh <-chan []*model.Result,
	resultWriter io.Writer,
	interimWriter io.Writer,
) *ResultWriter {
	return &ResultWriter{
		resultCh:      resultCh,
		resultWriter:  resultWriter,
		interimWriter: interimWriter,
	}
}

func (w *ResultWriter) Start(ctx context.Context) error {
	buf := bytes.Buffer{}
	var interimResult []byte
	defer func() {
		if len(interimResult) == 0 {
			return
		}
		if _, err := w.resultWriter.Write(interimResult); err != nil {
			slog.Error(fmt.Sprintf("failed to write interim result: %v", err))
		}
		slog.Debug("ResponseProcessor: interim result written")
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case results, ok := <-w.resultCh:
			if !ok {
				return fmt.Errorf("result channel is closed")
			}

			buf.Reset()
			for _, result := range results {
				if !result.IsFinal {
					buf.WriteString(result.Transcript)
					continue
				}

				if _, err := w.resultWriter.Write([]byte(result.Transcript)); err != nil {
					return fmt.Errorf("failed to write result: %w", err)
				}
				interimResult = nil
				buf.Reset()
			}

			if buf.Len() == 0 {
				continue
			}

			interimResult = buf.Bytes()
			if _, err := w.interimWriter.Write(interimResult); err != nil {
				return fmt.Errorf("failed to write interim result: %w", err)
			}
		}
	}
}
