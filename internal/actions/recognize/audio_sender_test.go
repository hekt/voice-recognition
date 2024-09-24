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

var _ io.Reader = &audioSenderTestReader{}

func Test_audioSender_Start(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expectedTotalSendCallsCount := 3
		expectedTotalCloseSendCallsCount := 2

		sentBuf := bytes.Buffer{}
		sendCalls := make(chan struct{}, expectedTotalSendCallsCount)
		closeSendCalls := make(chan struct{}, expectedTotalCloseSendCallsCount)
		crateStreamMock := func() *ispeechpb.Speech_StreamingRecognizeClientMock {
			return &ispeechpb.Speech_StreamingRecognizeClientMock{
				SendFunc: func(req *speechpb.StreamingRecognizeRequest) error {
					audioReq, ok := req.StreamingRequest.(*speechpb.StreamingRecognizeRequest_Audio)
					if !ok {
						t.Errorf("unexpected request type: %T", req.StreamingRequest)
					}
					sentBuf.Write(audioReq.Audio)
					sendCalls <- struct{}{}
					return nil
				},
				CloseSendFunc: func() error {
					closeSendCalls <- struct{}{}
					return nil
				},
			}
		}
		stream1 := crateStreamMock()
		stream2 := crateStreamMock()

		chunkSize := 16
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
			bufferSize:   chunkSize,
		}

		var wg sync.WaitGroup
		wg.Add(1)
		var got error
		go func() {
			defer wg.Done()
			got = s.Start(context.Background())
		}()

		firstChunk := bytes.Repeat([]byte("a"), chunkSize)
		secondChunk := bytes.Repeat([]byte("b"), chunkSize)
		thirdChunk := []byte("c")

		// Send initial stream1 to AudioSender.
		sendStreamCh <- stream1

		// Send first data chunk to AudioSender.
		bufCh <- firstChunk

		// Wait until the first Send() call is made on stream1.
		<-sendCalls

		// At this point, AudioSender is waiting for the next Read() result.

		// Send stream2 to AudioSender while it's waiting for Read().
		sendStreamCh <- stream2

		// AudioSender is still waiting for Read() result; stream not switched yet.

		// Send second data chunk to AudioSender.
		bufCh <- secondChunk

		// AudioSender reads the data, sends it to stream1, then checks select and switches to stream2.
		<-closeSendCalls

		// Now, AudioSender has switched to stream2.

		// Send final data chunk to AudioSender.
		bufCh <- thirdChunk

		// Signal EOF to AudioSender to stop reading.
		eofCh <- struct{}{}

		wg.Wait()

		if got != nil {
			t.Errorf("audioSender.Start() = %v, want nil", got)
		}
		wantSent := string(firstChunk) + string(secondChunk) + string(thirdChunk)
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
