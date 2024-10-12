package recognizer

import (
	"bytes"
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/hekt/voice-recognition/internal/testutil"
)

func TestNewAudioReceiver(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		audioCh := make(chan []byte)
		audioReader := &bytes.Buffer{}
		s := NewAudioReceiver(audioReader, audioCh, 1024)

		want := &AudioReader{
			reader:     audioReader,
			audioCh:    audioCh,
			bufferSize: 1024,
		}
		if !reflect.DeepEqual(s, want) {
			t.Errorf("NewAudioReceiver() = %v, want %v", s, want)
		}
	})
}

func Test_AudioReceiver_Start(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		chunkSize := 16
		audioReader := testutil.NewChannelReader()
		bufCh := audioReader.BufCh
		eofCh := audioReader.EOFCh
		audioCh := make(chan []byte, 3)

		r := &AudioReader{
			reader:     audioReader,
			audioCh:    audioCh,
			bufferSize: chunkSize,
		}

		var wg sync.WaitGroup
		wg.Add(1)
		var got error
		go func() {
			defer wg.Done()
			got = r.Start(context.Background())
		}()

		firstChunk := bytes.Repeat([]byte("a"), chunkSize)
		secondChunk := bytes.Repeat([]byte("b"), chunkSize)
		thirdChunk := []byte("c")

		bufCh <- firstChunk
		bufCh <- secondChunk
		bufCh <- thirdChunk
		eofCh <- struct{}{}

		wg.Wait()

		if got != nil {
			t.Errorf("audioSender.Start() = %v, want nil", got)
		}
		if g, w := <-audioCh, firstChunk; !reflect.DeepEqual(g, w) {
			t.Errorf("audioCh = %v, want %v", g, w)
		}
		if g, w := <-audioCh, secondChunk; !reflect.DeepEqual(g, w) {
			t.Errorf("audioCh = %v, want %v", g, w)
		}
		if g, w := <-audioCh, thirdChunk; !reflect.DeepEqual(g, w) {
			t.Errorf("audioCh = %v, want %v", g, w)
		}
	})
}
