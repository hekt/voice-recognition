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
	CreateRecognizer(
		ctx context.Context,
		req *speechpb.CreateRecognizerRequest,
		opts ...gax.CallOption,
	) (*speech.CreateRecognizerOperation, error)
	ListRecognizers(
		ctx context.Context,
		req *speechpb.ListRecognizersRequest,
		opts ...gax.CallOption,
	) *speech.RecognizerIterator
	UpdateRecognizer(
		ctx context.Context,
		req *speechpb.UpdateRecognizerRequest,
		opts ...gax.CallOption,
	) (*speech.UpdateRecognizerOperation, error)
	DeleteRecognizer(
		ctx context.Context,
		req *speechpb.DeleteRecognizerRequest,
		opts ...gax.CallOption,
	) (*speech.DeleteRecognizerOperation, error)

	CreatePhraseSet(
		ctx context.Context,
		req *speechpb.CreatePhraseSetRequest,
		opts ...gax.CallOption,
	) (*speech.CreatePhraseSetOperation, error)
	ListPhraseSets(
		ctx context.Context,
		req *speechpb.ListPhraseSetsRequest,
		opts ...gax.CallOption,
	) *speech.PhraseSetIterator
	UpdatePhraseSet(
		ctx context.Context,
		req *speechpb.UpdatePhraseSetRequest,
		opts ...gax.CallOption,
	) (*speech.UpdatePhraseSetOperation, error)
	DeletePhraseSet(
		ctx context.Context,
		req *speechpb.DeletePhraseSetRequest,
		opts ...gax.CallOption,
	) (*speech.DeletePhraseSetOperation, error)

	StreamingRecognize(
		ctx context.Context,
		opts ...gax.CallOption,
	) (speechpb.Speech_StreamingRecognizeClient, error)

	Close() error
}

var _ Client = (*speech.Client)(nil)
