package recognize

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/hekt/voice-recognition/internal/interfaces/speech"
)

//go:generate moq -rm -out stream_supplier_mock.go . StreamSupplierInterface
type StreamSupplierInterface interface {
	// Supply is supplies the streams once.
	Supply(ctx context.Context) error
	// Start is starts supplying the streams at regular intervals.
	Start(ctx context.Context) error
}

var _ StreamSupplierInterface = (*StreamSupplier)(nil)

type StreamSupplier struct {
	// client is a client of the Speech-to-Text API.
	client speech.Client
	// sendStreamCh is a channel to pass the sending stream.
	sendStreamCh chan<- speechpb.Speech_StreamingRecognizeClient
	// receiveStreamCh is a channel to pass the receiving stream.
	receiveStreamCh chan<- speechpb.Speech_StreamingRecognizeClient

	// recognizerFullName is the full name of the recognizer.
	recognizerFullName string
	// supplyInterval is the interval of stream supply.
	supplyInterval time.Duration
}

func NewStreamSupplier(
	client speech.Client,
	sendStreamCh chan<- speechpb.Speech_StreamingRecognizeClient,
	receiveStreamCh chan<- speechpb.Speech_StreamingRecognizeClient,
	recognizerFullName string,
	supplyInterval time.Duration,
) *StreamSupplier {
	return &StreamSupplier{
		client:             client,
		sendStreamCh:       sendStreamCh,
		receiveStreamCh:    receiveStreamCh,
		recognizerFullName: recognizerFullName,
		supplyInterval:     supplyInterval,
	}
}

func (s *StreamSupplier) Start(ctx context.Context) error {
	slog.Debug("StreamSupplier: start")

	timer := time.NewTimer(s.supplyInterval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			slog.Debug("StreamSupplier: timer fired")

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

			slog.Debug("StreamSupplier: stream supplied")
		}
	}
}

func (s *StreamSupplier) Supply(ctx context.Context) error {
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

func (s *StreamSupplier) initializeStream(
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
