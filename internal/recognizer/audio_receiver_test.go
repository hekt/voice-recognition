package recognizer

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"sync"
	"testing"
)

func TestNewAudioReceiver(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		audioCh := make(chan []byte)
		audioReader := &bytes.Buffer{}
		s := NewAudioReceiver(audioReader, audioCh, 1024)

		want := &AudioReceiver{
			audioReader: audioReader,
			audioCh:     audioCh,
			bufferSize:  1024,
		}
		if !reflect.DeepEqual(s, want) {
			t.Errorf("NewAudioReceiver() = %v, want %v", s, want)
		}
	})
}

type audioReceiverTestReader struct {
	bufCh chan []byte
	eofCh <-chan struct{}
}

func (r *audioReceiverTestReader) Read(p []byte) (int, error) {
	select {
	case <-r.eofCh:
		return 0, io.EOF
	case buf := <-r.bufCh:
		n := copy(p, buf)
		if n < len(buf) {
			r.bufCh <- buf[n:]
		}
		return n, nil
	}
}

var _ io.Reader = &audioReceiverTestReader{}

func Test_AudioReceiver_Start(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		chunkSize := 16
		bufCh := make(chan []byte)
		eofCh := make(chan struct{})
		audioReader := &audioReceiverTestReader{
			bufCh: bufCh,
			eofCh: eofCh,
		}
		audioCh := make(chan []byte, 3)

		r := &AudioReceiver{
			audioReader: audioReader,
			audioCh:     audioCh,
			bufferSize:  chunkSize,
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
