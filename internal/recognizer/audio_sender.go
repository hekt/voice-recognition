package recognizer

import (
	"context"
	"fmt"
	"log/slog"

	"cloud.google.com/go/speech/apiv2/speechpb"
)

//go:generate moq -rm -out audio_sender_mock.go . AudioSenderInterface
type AudioSenderInterface interface {
	Start(ctx context.Context) error
}

var _ AudioSenderInterface = &AudioSender{}

type AudioSender struct {
	audioCh      <-chan []byte
	sendStreamCh <-chan speechpb.Speech_StreamingRecognizeClient
}

func NewAudioSender(
	audioCh <-chan []byte,
	sendStreamCh <-chan speechpb.Speech_StreamingRecognizeClient,
) *AudioSender {
	return &AudioSender{
		audioCh:      audioCh,
		sendStreamCh: sendStreamCh,
	}
}

func (s *AudioSender) Start(ctx context.Context) error {
	slog.Debug("AudioSender: start")

	stream, ok := <-s.sendStreamCh
	if !ok {
		return fmt.Errorf("failed to get send stream from channel")
	}
	defer func() {
		if err := stream.CloseSend(); err != nil {
			slog.Error(fmt.Sprintf("failed to close send direction of stream: %v", err))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case newStream, ok := <-s.sendStreamCh:
			if !ok {
				return fmt.Errorf("send stream channel is closed")
			}
			slog.Debug("AudioSender: new stream received")

			// when the new stream is received, close the current stream and switch to the new stream.
			if err := stream.CloseSend(); err != nil {
				return fmt.Errorf("failed to close send direction of stream on reconnect: %w", err)
			}

			stream = newStream
			slog.Debug("AudioSender: stream switched")
		case audio, ok := <-s.audioCh:
			if !ok {
				return fmt.Errorf("audio channel is closed")
			}
			if err := stream.Send(&speechpb.StreamingRecognizeRequest{
				StreamingRequest: &speechpb.StreamingRecognizeRequest_Audio{
					Audio: audio,
				},
			}); err != nil {
				return fmt.Errorf("failed to send audio data: %w", err)
			}
		}
	}
}
