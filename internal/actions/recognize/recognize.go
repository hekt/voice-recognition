package recognize

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	speech "cloud.google.com/go/speech/apiv2"
	"github.com/hekt/voice-recognition/internal/file"
)

const (
	// maxStreamDuration is the maximum duration for which the stream remains connected
	maxStreamDuration = 5 * time.Minute

	// reconnectLeadTime is the lead time before the stream timeout to initiate reconnection attempts.
	reconnectLeadTime = 10 * time.Second
)

type Args struct {
	ProjectID      string
	RecognizerName string
}

func Run(ctx context.Context, args Args, opts ...Option) error {
	options := &options{
		outputFilePath:    fmt.Sprintf("output/%d.txt", time.Now().Unix()),
		bufferSize:        1024,
		reconnectInterval: maxStreamDuration - reconnectLeadTime,
	}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// This behavior ensures the output file is created early,
	// making it easier to use with tools like `tail -f`.
	{
		f, err := os.OpenFile(options.outputFilePath, os.O_CREATE, os.FileMode(0o644))
		if err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("failed to close output file: %w", err)
		}
	}

	audioReader := os.Stdin
	resultWriter := file.NewOpenCloseFileWriter(
		options.outputFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		os.FileMode(0o644),
	)
	interimWriter := os.Stdout

	client, err := speech.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create speech client: %w", err)
	}

	recognizer, err := NewRecognizer(
		args.ProjectID,
		args.RecognizerName,
		options.reconnectInterval,
		options.bufferSize,
		client,
		audioReader,
		resultWriter,
		interimWriter,
	)
	if err != nil {
		return err
	}

	if err := recognizer.Start(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
	return nil
}
