package speech

import (
	"context"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/googleapis/gax-go/v2"
)

// SpeechClient is an interface of cloud.google.com/go/speech/apiv2.Client.
// define the interface to create a mock for testing because apiv2.Client is a struct.
//
//go:generate moq -rm -out client_mock.go . Client
type Client interface {
	StreamingRecognize(ctx context.Context, opts ...gax.CallOption) (speechpb.Speech_StreamingRecognizeClient, error)
	Close() error
}

var _ Client = (*speech.Client)(nil)
