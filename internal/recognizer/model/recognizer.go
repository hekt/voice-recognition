package model

import "context"

//go:generate moq -rm -out recognizer_mock.go . RecognizerCoreInterface
type RecognizerCoreInterface interface {
	Start(ctx context.Context) error
}
