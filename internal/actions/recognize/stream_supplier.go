package recognize

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/hekt/voice-recognition/pkg/speech"
)

type StreamSupplier interface {
	// Supply はストリームを一度だけ提供する。
	Supply(ctx context.Context) error
	// Start は一定間隔でのストリームの提供を開始する。
	Start(ctx context.Context) error
}

type streamSupplier struct {
	client speech.Client
	// sendStreamCh は送信用の stream を受け渡しする channel。
	sendStreamCh chan<- speechpb.Speech_StreamingRecognizeClient
	// receiveStreamCh は受信用の stream を受け渡しする channel。
	receiveStreamCh chan<- speechpb.Speech_StreamingRecognizeClient

	// recognizerFullName は recognizer のフルネーム。
	recognizerFullName string
	// reconnectInterval はストリームの提供間隔を表す。
	supplyInterval time.Duration
}

func NewStreamSupplier(
	client speech.Client,
	sendStreamCh chan<- speechpb.Speech_StreamingRecognizeClient,
	receiveStreamCh chan<- speechpb.Speech_StreamingRecognizeClient,
	recognizerFullName string,
	supplyInterval time.Duration,
) StreamSupplier {
	return &streamSupplier{
		client:             client,
		sendStreamCh:       sendStreamCh,
		receiveStreamCh:    receiveStreamCh,
		recognizerFullName: recognizerFullName,
		supplyInterval:     supplyInterval,
	}
}

func (s *streamSupplier) Start(ctx context.Context) error {
	timer := time.NewTimer(s.supplyInterval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			newStream, err := s.initializeStream(ctx)
			if err != nil {
				return fmt.Errorf("failed to initialize stream: %w", err)
			}

			timer.Reset(s.supplyInterval)

			select {
			case s.sendStreamCh <- newStream:
			case <-ctx.Done():
				return nil
			}

			select {
			case s.receiveStreamCh <- newStream:
			case <-ctx.Done():
				return nil
			}
		}
	}
}

func (s *streamSupplier) Supply(ctx context.Context) error {
	stream, err := s.initializeStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize stream: %w", err)
	}

	select {
	case s.sendStreamCh <- stream:
	case <-ctx.Done():
		return fmt.Errorf("context is done: %w", context.Canceled)
	}

	select {
	case s.receiveStreamCh <- stream:
	case <-ctx.Done():
		return fmt.Errorf("context is done: %w", context.Canceled)
	}

	return nil
}

func (s *streamSupplier) initializeStream(
	ctx context.Context,
) (speechpb.Speech_StreamingRecognizeClient, error) {
	stream, err := s.client.StreamingRecognize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	if err := stream.Send(&speechpb.StreamingRecognizeRequest{
		Recognizer: s.recognizerFullName,
		StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
			StreamingConfig: &speechpb.StreamingRecognitionConfig{
				StreamingFeatures: &speechpb.StreamingRecognitionFeatures{
					InterimResults: true,
				},
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to send initial request: %w", err)
	}

	return stream, nil
}
