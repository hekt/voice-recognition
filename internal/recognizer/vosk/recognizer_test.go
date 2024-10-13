package vosk

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	myvosk "github.com/hekt/voice-recognition/internal/interfaces/vosk"
	"github.com/hekt/voice-recognition/internal/recognizer/model"
)

func TestNewRecognizer(t *testing.T) {
	type args struct {
		recognizer *myvosk.VoskRecognizerMock
		audioCh    <-chan []byte
		resultCh   chan<- []*model.Result
	}
	baseArgs := args{
		recognizer: &myvosk.VoskRecognizerMock{},
		audioCh:    make(chan []byte),
		resultCh:   make(chan []*model.Result),
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "success",
			args:    baseArgs,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRecognizer(tt.args.recognizer, tt.args.audioCh, tt.args.resultCh)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRecognizer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got == nil {
				t.Errorf("NewRecognizer() = nil, want non-nil")
			}
		})
	}
}

func TestRecognizer_Start(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		mockWavefromCh := make(chan int)
		mockResultCh := make(chan []byte)
		mockPartialResultCh := make(chan []byte)
		recognizer := &myvosk.VoskRecognizerMock{
			AcceptWaveformFunc: func(bytes []byte) int {
				return <-mockWavefromCh
			},
			ResultFunc: func() []byte {
				return <-mockResultCh
			},
			PartialResultFunc: func() []byte {
				return <-mockPartialResultCh
			},
		}
		audioCh := make(chan []byte)
		resultCh := make(chan []*model.Result, 3)

		r := &Recognizer{
			recognizer: recognizer,
			audioCh:    audioCh,
			resultCh:   resultCh,
		}

		var wg sync.WaitGroup
		wg.Add(1)
		var err error
		go func() {
			defer wg.Done()
			err = r.Start(ctx)
		}()

		audioCh <- []byte("hello")
		mockWavefromCh <- 0
		mockPartialResultCh <- []byte(`{"partial":"hello"}`)
		audioCh <- []byte("world")
		mockWavefromCh <- 1
		mockResultCh <- []byte(`{"text":"world"}`)

		for {
			time.Sleep(10 * time.Millisecond)
			if len(resultCh) == 2 {
				break
			}
		}

		cancel()
		wg.Wait()

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Recognizer.Start() error = %v, want %v", err, context.Canceled)
		}
		wantResults := [][]*model.Result{
			{{Transcript: "hello", IsFinal: false}},
			{{Transcript: "world", IsFinal: true}},
		}
		for _, want := range wantResults {
			got := <-resultCh
			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("unexpected result (-got +want):\n%s", diff)
			}
		}
	})
}

func Test_parsePartialResult(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "success",
			args:    args{data: []byte(`{"partial":"hello"}`)},
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "empty",
			args:    args{data: []byte(`{"partial":""}`)},
			want:    "",
			wantErr: false,
		},
		{
			name:    "non partial result",
			args:    args{data: []byte(`{"text":"hello"}`)},
			want:    "",
			wantErr: false,
		},
		{
			name:    "invalid json",
			args:    args{data: []byte(`{"partial":"hello"`)},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePartialResult(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePartialResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parsePartialResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseResult(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "success",
			args:    args{data: []byte(`{"text":"hello"}`)},
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "empty",
			args:    args{data: []byte(`{"text":""}`)},
			want:    "",
			wantErr: false,
		},
		{
			name:    "partial result",
			args:    args{data: []byte(`{"partial":"hello"}`)},
			want:    "",
			wantErr: false,
		},
		{
			name:    "invalid json",
			args:    args{data: []byte(`{"text":"hello"`)},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseResult(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseResult() = %v, want %v", got, tt.want)
			}
		})
	}
}
