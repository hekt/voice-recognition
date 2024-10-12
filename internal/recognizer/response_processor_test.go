package recognizer

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/google/go-cmp/cmp"
	"github.com/hekt/voice-recognition/internal/recognizer/model"
)

func TestNewResponseProcessor(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		responseCh := make(chan *speechpb.StreamingRecognizeResponse)
		resultCh := make(chan []*model.Result)
		processCh := make(chan struct{})

		got := NewResponseProcessor(responseCh, resultCh, processCh)
		want := &ResponseProcessor{
			responseCh: responseCh,
			resultCh:   resultCh,
			processCh:  processCh,
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("NewResponseProcessor() = %v, want %v", got, want)
		}
	})
}

func Test_ResponseProcessor_Start(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		responseCh := make(chan *speechpb.StreamingRecognizeResponse)
		resultCh := make(chan []*model.Result, 2)
		processCh := make(chan struct{})

		p := &ResponseProcessor{
			responseCh: responseCh,
			resultCh:   resultCh,
			processCh:  processCh,
		}

		var wg sync.WaitGroup
		wg.Add(1)
		var got error
		go func() {
			defer wg.Done()
			got = p.Start(ctx)
		}()

		// IsFinal: false
		responseCh <- &speechpb.StreamingRecognizeResponse{
			Results: []*speechpb.StreamingRecognitionResult{
				{
					Alternatives: []*speechpb.SpeechRecognitionAlternative{
						{Transcript: "a"},
					},
					IsFinal: false,
				},
				{
					Alternatives: []*speechpb.SpeechRecognitionAlternative{
						{Transcript: "b"},
					},
					IsFinal: false,
				},
			},
		}
		<-processCh

		// no alternatives must be skipped
		responseCh <- &speechpb.StreamingRecognizeResponse{
			Results: []*speechpb.StreamingRecognitionResult{
				{
					Alternatives: []*speechpb.SpeechRecognitionAlternative{},
					IsFinal:      false,
				},
			},
		}

		// IsFinal: true
		responseCh <- &speechpb.StreamingRecognizeResponse{
			Results: []*speechpb.StreamingRecognitionResult{
				{
					Alternatives: []*speechpb.SpeechRecognitionAlternative{
						{Transcript: "abcd"},
					},
					IsFinal: true,
				},
				{
					Alternatives: []*speechpb.SpeechRecognitionAlternative{
						{Transcript: "x"},
					},
					IsFinal: false,
				},
			},
		}
		<-processCh

		// 中断して完了まで待つ
		cancel()
		wg.Wait()

		if !errors.Is(got, context.Canceled) {
			t.Errorf("Start() error = %v, want %v", got, context.Canceled)
		}

		wantResults := [][]*model.Result{
			{
				{Transcript: "a", IsFinal: false},
				{Transcript: "b", IsFinal: false},
			},
			{
				{Transcript: "abcd", IsFinal: true},
				{Transcript: "x", IsFinal: false},
			},
		}
		close(resultCh)
		gotResults := make([][]*model.Result, 0, len(resultCh))
		for rs := range resultCh {
			gotResults = append(gotResults, rs)
		}
		if diff := cmp.Diff(gotResults, wantResults); diff != "" {
			t.Errorf("unexpected result: (-got +want)\n%s", diff)
		}
	})

	t.Run("closed stream", func(t *testing.T) {
		responseCh := make(chan *speechpb.StreamingRecognizeResponse)
		close(responseCh)

		p := &ResponseProcessor{
			responseCh: responseCh,
		}

		if got := p.Start(context.Background()); got == nil {
			t.Error("Start() error = nil, want an error")
		}
	})
}
