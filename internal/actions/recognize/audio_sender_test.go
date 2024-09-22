package recognize

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"sync"
	"testing"

	"cloud.google.com/go/speech/apiv2/speechpb"
	ispeechpb "github.com/hekt/voice-recognition/internal/interfaces/speechpb"
)

func TestNewAudioSender(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)
		audioReader := &bytes.Buffer{}
		s := NewAudioSender(audioReader, sendStreamCh, 1024)

		want := &AudioSender{
			audioReader:  audioReader,
			sendStreamCh: sendStreamCh,
			bufferSize:   1024,
		}
		if !reflect.DeepEqual(s, want) {
			t.Errorf("NewAudioSender() = %v, want %v", s, want)
		}
	})
}

type audioSenderTestReader struct {
	bufCh chan []byte
	eofCh <-chan struct{}
}

func (r *audioSenderTestReader) Read(p []byte) (int, error) {
	select {
	case buf := <-r.bufCh:
		n := copy(p, buf)
		if n < len(buf) {
			r.bufCh <- buf[n:]
		}
		return n, nil
	case <-r.eofCh:
		return 0, io.EOF
	}
}

var _ io.Reader = &audioSenderTestReader{}

func Test_audioSender_Start(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sentBuf := bytes.Buffer{}
		crateStreamMock := func() *ispeechpb.Speech_StreamingRecognizeClientMock {
			return &ispeechpb.Speech_StreamingRecognizeClientMock{
				SendFunc: func(req *speechpb.StreamingRecognizeRequest) error {
					audioReq, ok := req.StreamingRequest.(*speechpb.StreamingRecognizeRequest_Audio)
					if !ok {
						t.Errorf("unexpected request type: %T", req.StreamingRequest)
					}
					sentBuf.Write(audioReq.Audio)
					return nil
				},
				CloseSendFunc: func() error {
					return nil
				},
			}
		}
		stream1 := crateStreamMock()
		stream2 := crateStreamMock()

		bufCh := make(chan []byte)
		eofCh := make(chan struct{})
		audioReader := &audioSenderTestReader{
			bufCh: bufCh,
			eofCh: eofCh,
		}
		sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 2)

		s := &AudioSender{
			audioReader:  audioReader,
			sendStreamCh: sendStreamCh,
			bufferSize:   16,
		}

		var wg sync.WaitGroup
		wg.Add(1)
		var got error
		go func() {
			defer wg.Done()
			got = s.Start(context.Background())
		}()

		sendStreamCh <- stream1
		bufCh <- []byte("aaaaaaaaaaaaaaaa")
		bufCh <- []byte("bbbbbbbbbbbbbbbb")
		sendStreamCh <- stream2
		bufCh <- []byte("c")
		eofCh <- struct{}{}

		wg.Wait()

		if got != nil {
			t.Errorf("audioSender.Start() = %v, want nil", got)
		}
		wantSent := "aaaaaaaaaaaaaaaa" + "bbbbbbbbbbbbbbbb" + "c"
		if got := sentBuf.String(); got != wantSent {
			t.Errorf("sent audio = %q, want %q", got, wantSent)
		}
		if count := len(stream1.SendCalls()); count != 2 {
			t.Errorf("stream1.Send() called %d times, want 2 times", count)
		}
		if count := len(stream1.CloseSendCalls()); count != 1 {
			t.Errorf("stream1.CloseSend() called %d times, want 1 times", count)
		}
		if count := len(stream2.SendCalls()); count != 1 {
			t.Errorf("stream2.Send() called %d times, want 1 times", count)
		}
		if count := len(stream2.CloseSendCalls()); count != 1 {
			t.Errorf("stream2.CloseSend() called %d times, want 1 times", count)
		}
	})

	t.Run("closed stream", func(t *testing.T) {
		sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)
		s := &AudioSender{
			sendStreamCh: sendStreamCh,
		}
		close(sendStreamCh)

		if got := s.Start(context.Background()); got == nil {
			t.Errorf("audioSender.Start() = nil, want error")
		}
	})
}
