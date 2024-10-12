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

	"golang.org/x/sync/errgroup"

	"github.com/hekt/voice-recognition/internal/recognizer/google"
	"github.com/hekt/voice-recognition/internal/recognizer/model"
)

type Recognizer struct {
	audioReader    AudioReaderInterface
	recognizer     model.RecognizerCoreInterface
	resultWriter   ResultWriterInterface
	processMonitor ProcessMonitorInterface

	audioCh   chan []byte
	resultCh  chan []*model.Result
	processCh chan struct{}
}

func New(
	ctx context.Context,
	projectID string,
	recognizerName string,
	reconnectInterval time.Duration,
	bufferSize int,
	inactiveTimeout time.Duration,
	ioAudioReader io.Reader,
	ioResultWriter io.Writer,
	ioInterimWriter io.Writer,
) (*Recognizer, error) {
	if bufferSize < 1024 {
		return nil, errors.New("buffer size must be greater than or equal to 1024")
	}
	if inactiveTimeout == 0 {
		return nil, errors.New("inactive timeout must be specified")
	}
	if ioAudioReader == nil {
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
	resultCh := make(chan []*model.Result, 10)
	processCh := make(chan struct{}, 1)

	recognizer, err := google.NewRecognizer(
		ctx,
		audioCh,
		resultCh,
		projectID,
		recognizerName,
		reconnectInterval,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create google recognizer: %w", err)
	}

	audioReader := NewAudioReceiver(ioAudioReader, audioCh, bufferSize)
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
		recognizer:     recognizer,
		audioReader:    audioReader,
		resultWriter:   resultWriter,
		processMonitor: processMonitor,

		audioCh:   audioCh,
		resultCh:  resultCh,
		processCh: processCh,
	}, nil
}

func (r *Recognizer) Start(ctx context.Context) error {
	slog.Debug("recognizer started")

	defer func() {
		close(r.audioCh)
		close(r.resultCh)
		close(r.processCh)
	}()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if err := r.processMonitor.Start(ctx); err != nil {
			return fmt.Errorf("error occured in process monitor: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		if err := r.recognizer.Start(ctx); err != nil {
			return fmt.Errorf("error occured in recognizer: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		if err := r.audioReader.Start(ctx); err != nil {
			return fmt.Errorf("error occured in audio receiver: %w", err)
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
