package recognize

import (
	"context"
	"fmt"
	"time"
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
		outputFilePath = fmt.Sprintf("output-%d.txt", time.Now().Unix())
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

	recognizer, err := newRecognizer(
		ctx,
		arg.ProjectID,
		arg.RecognizerName,
		outputFilePath,
		bufferSize,
		reconnectInterval,
	)
	if err != nil {
		return err
	}

	return recognizer.Start(ctx)
}
