package speech

import (
	"context"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/googleapis/gax-go/v2"
)

// SpeechClient は cloud.google.com/go/speech/apiv2.Client のインターフェース。
// Client は具体型なので、テスト用にモックを作成するためにインターフェースを定義する。
//
//go:generate moq -rm -out client_mock.go . Client
type Client interface {
	StreamingRecognize(ctx context.Context, opts ...gax.CallOption) (speechpb.Speech_StreamingRecognizeClient, error)
	Close() error
}

var _ Client = (*speech.Client)(nil)
