package recognize

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/speech/apiv2/speechpb"
)

//go:generate moq -out audio_sender_mock.go . AudioSender
type AudioSender interface {
	Start(ctx context.Context) error
}

type audioSender struct {
	audioReader  io.Reader
	sendStreamCh <-chan speechpb.Speech_StreamingRecognizeClient
	bufferSize   int
}

func NewAudioSender(
	audioReader io.Reader,
	sendStreamCh <-chan speechpb.Speech_StreamingRecognizeClient,
	bufferSize int,
) AudioSender {
	return &audioSender{
		audioReader:  audioReader,
		sendStreamCh: sendStreamCh,
		bufferSize:   bufferSize,
	}
}

func (s *audioSender) Start(ctx context.Context) error {
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
			// 新しい stream が来たら古い stream を閉じて新しい stream に切り替える。
			if err := stream.CloseSend(); err != nil {
				return fmt.Errorf("failed to close send direction of stream: %w", err)
			}
			stream = newStream
		default:
			n, err := s.audioReader.Read(buf)
			if n > 0 {
				err := stream.Send(&speechpb.StreamingRecognizeRequest{
					StreamingRequest: &speechpb.StreamingRecognizeRequest_Audio{
						Audio: buf[:n],
					},
				})
				if err != nil {
					return fmt.Errorf("failed to send audio data: %w", err)
				}
			}
			if err == io.EOF {
				if e := stream.CloseSend(); e != nil {
					return fmt.Errorf("failed to close send direction of stream: %w", e)
				}
				return nil
			}
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
		}
	}
}
