package recognizer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"golang.org/x/sync/errgroup"

	"github.com/hekt/voice-recognition/internal/interfaces/speech"
	"github.com/hekt/voice-recognition/internal/recognizer/google"
	"github.com/hekt/voice-recognition/internal/recognizer/model"
	"github.com/hekt/voice-recognition/internal/resource"
)

type Recognizer struct {
	streamSupplier    google.StreamSupplierInterface
	audioReceiver     AudioReaderInterface
	audioSender       google.AudioSenderInterface
	reseponseReceiver google.ResponseReceiverInterface
	responseProcessor google.ResponseProcessorInterface
	resultWriter      ResultWriterInterface
	processMonitor    ProcessMonitorInterface

	// client は Speech-to-Text API のクライアント。
	client speech.Client
	// audioCh は音声データの受け渡しをする channel。
	audioCh chan []byte
	// responseCh はレスポンスデータの受け渡しをする channel。
	responseCh chan *speechpb.StreamingRecognizeResponse
	// sendStreamCh は送信用の stream を受け渡しする channel。
	sendStreamCh chan speechpb.Speech_StreamingRecognizeClient
	// receiveStreamCh は受信用の stream を受け渡しする channel。
	receiveStreamCh chan speechpb.Speech_StreamingRecognizeClient
	// processCh はレスポンスデータの処理を通知する channel。
	processCh chan struct{}
}

func New(
	projectID string,
	recognizerName string,
	reconnectInterval time.Duration,
	bufferSize int,
	inactiveTimeout time.Duration,
	client speech.Client,
	audioReader io.Reader,
	ioResultWriter io.Writer,
	ioInterimWriter io.Writer,
) (*Recognizer, error) {
	if projectID == "" {
		return nil, errors.New("project ID must be specified")
	}
	if recognizerName == "" {
		return nil, errors.New("recognizer name must be specified")
	}
	if bufferSize < 1024 {
		return nil, errors.New("buffer size must be greater than or equal to 1024")
	}
	if reconnectInterval < time.Minute {
		return nil, errors.New("reconnect interval must be greater than or equal to 1 minute")
	}
	if inactiveTimeout == 0 {
		return nil, errors.New("inactive timeout must be specified")
	}
	if client == nil {
		return nil, errors.New("client must be specified")
	}
	if audioReader == nil {
		return nil, errors.New("audio reader must be specified")
	}
	if ioResultWriter == nil {
		return nil, errors.New("result writer must be specified")
	}
	if ioInterimWriter == nil {
		return nil, errors.New("interim writer must be specified")
	}

	// not sure what is the appropriate buffer size.
	audioCh := make(chan []byte, 10)
	responseCh := make(chan *speechpb.StreamingRecognizeResponse, 10)
	resultCh := make(chan []*model.Result, 10)
	// if the stream is not taken out, there is no need to create new stream.
	// so buffer size is set to 1.
	sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)
	receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)
	processCh := make(chan struct{}, 1)

	streamSupplier := google.NewStreamSupplier(
		client,
		sendStreamCh,
		receiveStreamCh,
		resource.RecognizerFullname(projectID, recognizerName),
		reconnectInterval,
	)
	audioReceiver := NewAudioReceiver(audioReader, audioCh, bufferSize)
	audioSender := google.NewAudioSender(audioCh, sendStreamCh)
	responseReceiver := google.NewResponseReceiver(responseCh, receiveStreamCh)
	responseProcessor := google.NewResponseProcessor(
		responseCh,
		resultCh,
	)
	resultWriter := NewResultWriter(
		resultCh,
		&NotifyingWriter{
			Writer:   &DecoratedResultWriter{Writer: ioResultWriter},
			NotifyCh: processCh,
		},
		&NotifyingWriter{
			Writer:   &DecoratedInterimWriter{Writer: ioInterimWriter},
			NotifyCh: processCh,
		},
	)
	processMonitor := NewProcessMonitor(processCh, inactiveTimeout)

	return &Recognizer{
		streamSupplier:    streamSupplier,
		audioReceiver:     audioReceiver,
		audioSender:       audioSender,
		reseponseReceiver: responseReceiver,
		responseProcessor: responseProcessor,
		resultWriter:      resultWriter,
		processMonitor:    processMonitor,

		client:          client,
		audioCh:         audioCh,
		responseCh:      responseCh,
		sendStreamCh:    sendStreamCh,
		receiveStreamCh: receiveStreamCh,
		processCh:       processCh,
	}, nil
}

func (r *Recognizer) Start(ctx context.Context) error {
	slog.Debug("recognizer started")

	defer func() {
		close(r.audioCh)
		close(r.sendStreamCh)
		close(r.receiveStreamCh)
		close(r.responseCh)
		close(r.processCh)
		if err := r.client.Close(); err != nil {
			slog.Error(fmt.Sprintf("failed to close client: %v", err))
		}
	}()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	if err := r.streamSupplier.Supply(ctx); err != nil {
		return fmt.Errorf("failed to supply stream: %w", err)
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if err := r.processMonitor.Start(ctx); err != nil {
			return fmt.Errorf("error occured in process monitor: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		if err := r.audioReceiver.Start(ctx); err != nil {
			return fmt.Errorf("error occured in audio receiver: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		if err := r.streamSupplier.Start(ctx); err != nil {
			return fmt.Errorf("error occured in stream supplier: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		if err := r.reseponseReceiver.Start(ctx); err != nil {
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
	eg.Go(func() error {
		if err := r.resultWriter.Start(ctx); err != nil {
			return fmt.Errorf("error occured in result writer: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	slog.Debug("recognizer stopped")

	return nil
}
