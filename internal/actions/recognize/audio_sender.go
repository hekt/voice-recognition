package recognize

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"cloud.google.com/go/speech/apiv2/speechpb"
)

//go:generate moq -rm -out audio_sender_mock.go . AudioSenderInterface
type AudioSenderInterface interface {
	Start(ctx context.Context) error
}

var _ AudioSenderInterface = &AudioSender{}

type AudioSender struct {
	audioReader  io.Reader
	sendStreamCh <-chan speechpb.Speech_StreamingRecognizeClient
	bufferSize   int
}

func NewAudioSender(
	audioReader io.Reader,
	sendStreamCh <-chan speechpb.Speech_StreamingRecognizeClient,
	bufferSize int,
) *AudioSender {
	return &AudioSender{
		audioReader:  audioReader,
		sendStreamCh: sendStreamCh,
		bufferSize:   bufferSize,
	}
}

func (s *AudioSender) Start(ctx context.Context) error {
	slog.Debug("AudioSender: start")

	stream, ok := <-s.sendStreamCh
	if !ok {
		return fmt.Errorf("failed to get send stream from channel")
	}

	buf := make([]byte, s.bufferSize)
	for {
		select {
		case <-ctx.Done():
			return nil
		case newStream, ok := <-s.sendStreamCh:
			if !ok {
				return fmt.Errorf("failed to get new send stream from channel")
			}
			slog.Debug("AudioSender: new stream received")

			// when the new stream is received, close the current stream and switch to the new stream.
			if err := stream.CloseSend(); err != nil {
				return fmt.Errorf("failed to close send direction of stream on reconnect: %w", err)
			}

			stream = newStream
			slog.Debug("AudioSender: stream switched")
		default:
			n, err := s.audioReader.Read(buf)
			if errors.Is(err, io.EOF) {
				slog.Debug("AudioSender: EOF received")
				if err := stream.CloseSend(); err != nil {
					return fmt.Errorf("failed to close send direction of stream on EOF: %w", err)
				}
				return nil
			}
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}

			if n == 0 {
				continue
			}

			if err := stream.Send(&speechpb.StreamingRecognizeRequest{
				StreamingRequest: &speechpb.StreamingRecognizeRequest_Audio{
					Audio: buf[:n],
				},
			}); err != nil {
				return fmt.Errorf("failed to send audio data: %w", err)
			}
		}
	}
}
