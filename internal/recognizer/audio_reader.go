package recognizer

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

//go:generate moq -rm -out audio_reader_mock.go . AudioReaderInterface
type AudioReaderInterface interface {
	Start(ctx context.Context) error
}

var _ AudioReaderInterface = &AudioReader{}

type AudioReader struct {
	reader     io.Reader
	audioCh    chan<- []byte
	bufferSize int
}

func NewAudioReceiver(
	reader io.Reader,
	audioCh chan<- []byte,
	bufferSize int,
) *AudioReader {
	return &AudioReader{
		reader:     reader,
		audioCh:    audioCh,
		bufferSize: bufferSize,
	}
}

func (r *AudioReader) Start(ctx context.Context) error {
	slog.Debug("AudioReceiver: start")

	buf := make([]byte, r.bufferSize)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			n, err := r.reader.Read(buf)
			if err == io.EOF {
				slog.Debug("AudioReceiver: EOF received")
				return nil
			}
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}

			if n == 0 {
				continue
			}

			// Send copied buffer to audio channel.
			r.audioCh <- append(make([]byte, 0, n), buf[:n]...)
		}
	}
}
