package recognizer

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

//go:generate moq -rm -out audio_receiver_mock.go . AudioReceiverInterface
type AudioReceiverInterface interface {
	Start(ctx context.Context) error
}

var _ AudioReceiverInterface = &AudioReceiver{}

type AudioReceiver struct {
	audioReader io.Reader
	audioCh     chan<- []byte
	bufferSize  int
}

func NewAudioReceiver(
	audioReader io.Reader,
	audioCh chan<- []byte,
	bufferSize int,
) *AudioReceiver {
	return &AudioReceiver{
		audioReader: audioReader,
		audioCh:     audioCh,
		bufferSize:  bufferSize,
	}
}

func (r *AudioReceiver) Start(ctx context.Context) error {
	slog.Debug("AudioReceiver: start")

	buf := make([]byte, r.bufferSize)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			n, err := r.audioReader.Read(buf)
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

			dst := make([]byte, 0, n)
			r.audioCh <- append(dst, buf[:n]...)
		}
	}
}
