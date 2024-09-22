package recognize

import (
	"context"
	"fmt"
	"os"
	"time"

	speech "cloud.google.com/go/speech/apiv2"
	"github.com/hekt/voice-recognition/internal/file"
)

const (
	// streamTimeout はストリームのタイムアウト時間を表す。
	// この時間を超えるとサーバー側からストリームが切断される。
	streamTimeout = 5 * time.Minute

	// streamTimeoutOffset はストリームのタイムアウト時間のオフセットを表す。
	// この時間だけ短く設定することで、ストリームが切断される前に再接続を試みる。
	streamTimeoutOffset = 10 * time.Second
)

type Args struct {
	ProjectID      string
	RecognizerName string
}

func Run(ctx context.Context, arg Args, opts ...Option) error {
	options := &options{
		outputFilePath:    fmt.Sprintf("output/%d.txt", time.Now().Unix()),
		bufferSize:        1024,
		reconnectInterval: streamTimeout - streamTimeoutOffset,
	}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return fmt.Errorf("failed to apply option: %w", err)
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

	recognizer, err := newRecognizer(
		arg.ProjectID,
		arg.RecognizerName,
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

	return recognizer.Start(ctx)
}
