package vosk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	myvosk "github.com/hekt/voice-recognition/internal/interfaces/vosk"
	"github.com/hekt/voice-recognition/internal/recognizer/model"
)

var _ model.RecognizerCoreInterface = (*Recognizer)(nil)

type Recognizer struct {
	recognizer myvosk.VoskRecognizer
	audioCh    <-chan []byte
	resultCh   chan<- []*model.Result
}

func NewRecognizer(
	recognizer myvosk.VoskRecognizer,
	audioCh <-chan []byte,
	resultCh chan<- []*model.Result,
) (*Recognizer, error) {
	if recognizer == nil {
		return nil, errors.New("recognizer must be specified")
	}
	if audioCh == nil {
		return nil, errors.New("audio channel must be specified")
	}
	if resultCh == nil {
		return nil, errors.New("result channel must be specified")
	}

	return &Recognizer{
		recognizer: recognizer,
		audioCh:    audioCh,
		resultCh:   resultCh,
	}, nil
}

func (r *Recognizer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case audio, ok := <-r.audioCh:
			if !ok {
				return errors.New("audio channel closed")
			}

			n := r.recognizer.AcceptWaveform(audio)

			var results []*model.Result
			if n == 0 {
				t, err := parsePartialResult(r.recognizer.PartialResult())
				if err != nil {
					return fmt.Errorf("failed to parse partial result: %w", err)
				}
				if t == "" {
					continue
				}
				results = []*model.Result{
					{Transcript: t, IsFinal: false},
				}
			} else {
				t, err := parseResult(r.recognizer.Result())
				if err != nil {
					return fmt.Errorf("failed to parse result: %w", err)
				}
				if t == "" {
					continue
				}
				results = []*model.Result{
					{Transcript: t, IsFinal: true},
				}
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case r.resultCh <- results:
			}
		}
	}
}

func parsePartialResult(data []byte) (string, error) {
	var kv map[string]string
	if err := json.Unmarshal(data, &kv); err != nil {
		return "", err
	}
	return kv["partial"], nil
}

func parseResult(data []byte) (string, error) {
	var kv map[string]string
	if err := json.Unmarshal(data, &kv); err != nil {
		return "", err
	}
	return kv["text"], nil
}
