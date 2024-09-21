package speechpb

import (
	"cloud.google.com/go/speech/apiv2/speechpb"
)

//go:generate moq -rm -out speech_streaming_recognize_client_mock.go . Speech_StreamingRecognizeClient
type Speech_StreamingRecognizeClient = speechpb.Speech_StreamingRecognizeClient
