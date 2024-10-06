package recognizer

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/speech/apiv2/speechpb"
)

func TestNewResponseProcessor(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		resultBuf := &bytes.Buffer{}
		interimBuf := &bytes.Buffer{}
		responseCh := make(chan *speechpb.StreamingRecognizeResponse, 1)

		got := NewResponseProcessor(resultBuf, interimBuf, responseCh)
		want := &ResponseProcessor{
			resultWriter:  resultBuf,
			interimWriter: interimBuf,
			responseCh:    responseCh,
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("NewResponseProcessor() = %v, want %v", got, want)
		}
	})
}

func Test_responseProcessor_Start(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		resultBuf := &bytes.Buffer{}
		interimBuf := &bytes.Buffer{}
		responseCh := make(chan *speechpb.StreamingRecognizeResponse)

		p := &ResponseProcessor{
			resultWriter:  resultBuf,
			interimWriter: interimBuf,
			responseCh:    responseCh,
		}

		var wg sync.WaitGroup
		wg.Add(1)
		var got error
		go func() {
			defer wg.Done()
			got = p.Start(ctx)
		}()

		// (1) 中間応答レスポンス
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
		// (2) 中間応答レスポンス
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
				{
					Alternatives: []*speechpb.SpeechRecognitionAlternative{
						{Transcript: "c"},
					},
					IsFinal: false,
				},
			},
		}
		// no alternatives must be skipped
		responseCh <- &speechpb.StreamingRecognizeResponse{
			Results: []*speechpb.StreamingRecognitionResult{
				{
					Alternatives: []*speechpb.SpeechRecognitionAlternative{},
					IsFinal:      false,
				},
			},
		}
		// (3) 確定応答レスポンス
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
		// (4) 中間応答レスポンス
		responseCh <- &speechpb.StreamingRecognizeResponse{
			Results: []*speechpb.StreamingRecognitionResult{
				{
					Alternatives: []*speechpb.SpeechRecognitionAlternative{
						{Transcript: "x"},
					},
				},
				{
					Alternatives: []*speechpb.SpeechRecognitionAlternative{
						{Transcript: "y"},
					},
				},
			},
		}

		// 処理が行われるまで待つ
		time.Sleep(100 * time.Millisecond)

		// 中断して完了まで待つ
		cancel()
		wg.Wait()

		if !errors.Is(got, context.Canceled) {
			t.Errorf("Start() error = %v, want %v", got, context.Canceled)
		}

		// (3) の確定分 + 途中終了した (4)
		wantResult := "abcd" + "xy"
		if got := resultBuf.String(); got != wantResult {
			t.Errorf("Start() wrote results: %v, want %v", got, wantResult)
		}

		// (1) + (2) + (3) の未確定分 + (4)
		wantInterim := "ab" + "abc" + "x" + "xy"
		if got := interimBuf.String(); got != wantInterim {
			t.Errorf("Start() wrote interims: %v, want %v", got, wantInterim)
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
