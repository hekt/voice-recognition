package recognizer

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hekt/voice-recognition/internal/recognizer/model"
)

func TestNewResultWriter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		resultCh := make(chan []*model.Result)
		resultWriter := &bytes.Buffer{}
		interimWriter := &bytes.Buffer{}
		want := &ResultWriter{
			resultCh:      resultCh,
			resultWriter:  resultWriter,
			interimWriter: interimWriter,
		}
		got := NewResultWriter(resultCh, resultWriter, interimWriter)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("NewResultWriter() = %v, want %v", got, want)
		}
	})
}

func TestResultWriter_Start(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		resultCh := make(chan []*model.Result)
		resultWriter := &bytes.Buffer{}
		interimWriter := &bytes.Buffer{}
		w := &ResultWriter{
			resultCh:      resultCh,
			resultWriter:  resultWriter,
			interimWriter: interimWriter,
		}

		ctx, cancel := context.WithCancel(context.Background())

		var wg sync.WaitGroup
		wg.Add(1)
		var got error
		go func() {
			defer wg.Done()
			got = w.Start(ctx)
		}()

		resultCh <- []*model.Result{
			{Transcript: "a", IsFinal: false},
			{Transcript: "b", IsFinal: false},
		}
		resultCh <- []*model.Result{
			{Transcript: "abc", IsFinal: true},
		}
		resultCh <- []*model.Result{
			{Transcript: "d", IsFinal: false},
		}

		cancel()
		wg.Wait()

		if !errors.Is(got, context.Canceled) {
			t.Errorf("unexpected error: %v", got)
		}
		if diff := cmp.Diff(resultWriter.String(), "abcd"); diff != "" {
			t.Errorf("unexpected result: (-got +want)\n%s", diff)
		}
		if diff := cmp.Diff(interimWriter.String(), "abd"); diff != "" {
			t.Errorf("unexpected interim: (-got +want)\n%s", diff)
		}
	})

	t.Run("result channel is closed", func(t *testing.T) {
		resultCh := make(chan []*model.Result)
		resultWriter := &bytes.Buffer{}
		interimWriter := &bytes.Buffer{}
		w := &ResultWriter{
			resultCh:      resultCh,
			resultWriter:  resultWriter,
			interimWriter: interimWriter,
		}

		close(resultCh)
		got := w.Start(context.Background())

		if got == nil {
			t.Error("got = nil, want an error")
		}
	})
}
