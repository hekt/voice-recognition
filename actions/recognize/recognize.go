package recognize

import (
	"context"
	"fmt"
	"os"
	"time"

	speech "cloud.google.com/go/speech/apiv2"
	"github.com/hekt/voice-recognition/file"
)

const (
	// streamTimeout はストリームのタイムアウト時間を表す。
	// この時間を超えるとサーバー側からストリームが切断される。
	streamTimeout = 5 * time.Minute

	// streamTimeoutOffset はストリームのタイムアウト時間のオフセットを表す。
	// この時間だけ短く設定することで、ストリームが切断される前に再接続を試みる。
	streamTimeoutOffset = 10 * time.Second
)

type Arg struct {
	ProjectID         string
	RecognizerName    string
	OutputFilePath    string
	BufferSize        int
	ReconnectInterval time.Duration
}

func Run(ctx context.Context, arg Arg) error {
	outputFilePath := arg.OutputFilePath
	if outputFilePath == "" {
		outputFilePath = fmt.Sprintf("output/%d.txt", time.Now().Unix())
	}

	// 1KB未満は許容しない
	bufferSize := arg.BufferSize
	if bufferSize < 1024 {
		bufferSize = 1024
	}

	// 1分未満は許容しない
	reconnectInterval := arg.ReconnectInterval
	if reconnectInterval < time.Minute {
		reconnectInterval = streamTimeout - streamTimeoutOffset
	}

	audioReader := os.Stdin
	resultWriter := file.NewFileWriter(
		outputFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	interimWriter := os.Stdout

	client, err := speech.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create speech client: %w", err)
	}

	recognizer, err := newRecognizer(
		arg.ProjectID,
		arg.RecognizerName,
		reconnectInterval,
		bufferSize,
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
