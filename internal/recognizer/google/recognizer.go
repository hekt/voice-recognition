package google

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/speech/apiv2/speechpb"
	myspeech "github.com/hekt/voice-recognition/internal/interfaces/speech"
	"github.com/hekt/voice-recognition/internal/recognizer/model"
	"github.com/hekt/voice-recognition/internal/resource"
	"golang.org/x/sync/errgroup"
)

var _ model.RecognizerCoreInterface = (*Recognizer)(nil)

type Recognizer struct {
	streamSupplier    StreamSupplierInterface
	audioSender       AudioSenderInterface
	responseReceiver  ResponseReceiverInterface
	responseProcessor ResponseProcessorInterface

	client myspeech.Client

	responseCh      chan *speechpb.StreamingRecognizeResponse
	sendStreamCh    chan speechpb.Speech_StreamingRecognizeClient
	receiveStreamCh chan speechpb.Speech_StreamingRecognizeClient

	audioCh  <-chan []byte
	resultCh chan<- []*model.Result
}

func NewRecognizer(
	ctx context.Context,
	client myspeech.Client,
	audioCh <-chan []byte,
	resultCh chan<- []*model.Result,
	projectID string,
	recognizerName string,
	reconnectInterval time.Duration,
) (*Recognizer, error) {
	if projectID == "" {
		return nil, errors.New("project ID must be specified")
	}
	if recognizerName == "" {
		return nil, errors.New("recognizer name must be specified")
	}
	if reconnectInterval < time.Minute {
		return nil, errors.New("reconnect interval must be greater than or equal to 1 minute")
	}
	if client == nil {
		return nil, errors.New("client must be specified")
	}
	if audioCh == nil {
		return nil, errors.New("audio channel must be specified")
	}
	if resultCh == nil {
		return nil, errors.New("result channel must be specified")
	}

	sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)
	receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)
	responseCh := make(chan *speechpb.StreamingRecognizeResponse, 1)

	streamSupplier := NewStreamSupplier(
		client,
		sendStreamCh,
		receiveStreamCh,
		resource.RecognizerFullname(projectID, recognizerName),
		reconnectInterval,
	)
	audioSender := NewAudioSender(audioCh, sendStreamCh)
	responseReceiver := NewResponseReceiver(responseCh, receiveStreamCh)
	responseProcessor := NewResponseProcessor(responseCh, resultCh)

	return &Recognizer{
		streamSupplier:    streamSupplier,
		audioSender:       audioSender,
		responseReceiver:  responseReceiver,
		responseProcessor: responseProcessor,

		client: client,

		responseCh:      make(chan *speechpb.StreamingRecognizeResponse),
		sendStreamCh:    make(chan speechpb.Speech_StreamingRecognizeClient),
		receiveStreamCh: make(chan speechpb.Speech_StreamingRecognizeClient),

		audioCh:  audioCh,
		resultCh: resultCh,
	}, nil
}

func (r *Recognizer) Start(ctx context.Context) error {
	defer func() {
		close(r.responseCh)
		close(r.sendStreamCh)
		close(r.receiveStreamCh)
		if err := r.client.Close(); err != nil {
			slog.Error(fmt.Sprintf("failed to close client: %v", err))
		}
	}()

	if err := r.streamSupplier.Supply(ctx); err != nil {
		return fmt.Errorf("failed to supply stream: %w", err)
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if err := r.streamSupplier.Start(ctx); err != nil {
			return fmt.Errorf("error occured in stream supplier: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		if err := r.responseReceiver.Start(ctx); err != nil {
			return fmt.Errorf("error occured in response receiver: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		if err := r.audioSender.Start(ctx); err != nil {
			return fmt.Errorf("error occured in audio sender: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		if err := r.responseProcessor.Start(ctx); err != nil {
			return fmt.Errorf("error occured in response processor: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}
