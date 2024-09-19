package speech

import (
	"context"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/googleapis/gax-go/v2"
)

// SpeechClient は cloud.google.com/go/speech/apiv2.Client のインターフェース。
// Client は具体型なので、テスト用にモックを作成するためにインターフェースを定義する。
//
//go:generate moq -out client_mock.go . Client
type Client interface {
	StreamingRecognize(ctx context.Context, opts ...gax.CallOption) (speechpb.Speech_StreamingRecognizeClient, error)
	Close() error
}
