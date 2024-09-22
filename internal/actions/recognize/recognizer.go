package recognize

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"golang.org/x/sync/errgroup"

	"github.com/hekt/voice-recognition/internal/util"
	"github.com/hekt/voice-recognition/pkg/speech"
)

type recognizer struct {
	streamSupplier    StreamSupplier
	audioSender       AudioSender
	reseponseReceiver ResponseReceiver
	responseProcessor ResponseProcessor

	// client は Speech-to-Text API のクライアント。
	client speech.Client
	// responseCh はレスポンスデータの受け渡しをする channel。
	responseCh chan *speechpb.StreamingRecognizeResponse
	// sendStreamCh は送信用の stream を受け渡しする channel。
	sendStreamCh chan speechpb.Speech_StreamingRecognizeClient
	// receiveStreamCh は受信用の stream を受け渡しする channel。
	receiveStreamCh chan speechpb.Speech_StreamingRecognizeClient
}

func newRecognizer(
	projectID string,
	recognizerName string,
	reconnectInterval time.Duration,
	bufferSize int,
	client speech.Client,
	audioReader io.Reader,
	resultWriter io.Writer,
	interimWriter io.Writer,
) (*recognizer, error) {
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
	if client == nil {
		return nil, errors.New("client must be specified")
	}
	if audioReader == nil {
		return nil, errors.New("audio reader must be specified")
	}
	if resultWriter == nil {
		return nil, errors.New("result writer must be specified")
	}
	if interimWriter == nil {
		return nil, errors.New("interim writer must be specified")
	}

	// どの程度のバッファサイズが適切かは不明なため、バッファサイズは 10 にしている。
	responseCh := make(chan *speechpb.StreamingRecognizeResponse, 10)
	// stream が取り出されていないということはまだ使われていないということで、
	// その状態で新しい stream を作成する必要はないため、バッファサイズは 1 にしている。
	sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)
	receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)

	streamSupplier := NewStreamSupplier(
		client,
		sendStreamCh,
		receiveStreamCh,
		util.RecognizerFullname(projectID, recognizerName),
		reconnectInterval,
	)
	audioSender := NewAudioSender(audioReader, sendStreamCh, bufferSize)
	responseReceiver := NewResponseReceiver(responseCh, receiveStreamCh)
	responseProcessor := NewResponseProcessor(
		&DecoratedResultWriter{Writer: resultWriter},
		&DecoratedInterimWriter{Writer: interimWriter},
		responseCh,
	)

	return &recognizer{
		streamSupplier:    streamSupplier,
		audioSender:       audioSender,
		reseponseReceiver: responseReceiver,
		responseProcessor: responseProcessor,

		client:          client,
		responseCh:      responseCh,
		sendStreamCh:    sendStreamCh,
		receiveStreamCh: receiveStreamCh,
	}, nil
}

func (r *recognizer) Start(ctx context.Context) error {
	defer func() {
		r.client.Close()
		close(r.responseCh)
		close(r.sendStreamCh)
		close(r.receiveStreamCh)
	}()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

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

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}
