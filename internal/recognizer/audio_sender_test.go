package recognizer

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"cloud.google.com/go/speech/apiv2/speechpb"
	ispeechpb "github.com/hekt/voice-recognition/internal/interfaces/speechpb"
)

func TestNewAudioSender(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		audioCh := make(chan []byte)
		sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)
		s := NewAudioSender(audioCh, sendStreamCh)

		want := &AudioSender{
			audioCh:      audioCh,
			sendStreamCh: sendStreamCh,
		}
		if !reflect.DeepEqual(s, want) {
			t.Errorf("NewAudioSender() = %v, want %v", s, want)
		}
	})
}

func Test_audioSender_Start(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		expectedTotalSendCallsCount := 3
		expectedTotalCloseSendCallsCount := 2

		sentBuf := bytes.Buffer{}
		sendCalls := make(chan struct{}, expectedTotalSendCallsCount)
		closeSendCalls := make(chan struct{}, expectedTotalCloseSendCallsCount)
		crateStreamMock := func() *ispeechpb.Speech_StreamingRecognizeClientMock {
			return &ispeechpb.Speech_StreamingRecognizeClientMock{
				SendFunc: func(req *speechpb.StreamingRecognizeRequest) error {
					defer func() { sendCalls <- struct{}{} }()

					audioReq, ok := req.StreamingRequest.(*speechpb.StreamingRecognizeRequest_Audio)
					if !ok {
						t.Errorf("unexpected request type: %T", req.StreamingRequest)
					}
					sentBuf.Write(audioReq.Audio)

					return nil
				},
				CloseSendFunc: func() error {
					defer func() { closeSendCalls <- struct{}{} }()

					return nil
				},
			}
		}
		stream1 := crateStreamMock()
		stream2 := crateStreamMock()

		audioCh := make(chan []byte, 3)
		sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 2)

		s := &AudioSender{
			audioCh:      audioCh,
			sendStreamCh: sendStreamCh,
		}

		var wg sync.WaitGroup
		wg.Add(1)
		var got error
		go func() {
			defer wg.Done()
			got = s.Start(ctx)
		}()

		firstChunk := bytes.Repeat([]byte("a"), 16)
		secondChunk := bytes.Repeat([]byte("b"), 16)
		thirdChunk := []byte("c")

		sendStreamCh <- stream1
		audioCh <- firstChunk
		<-sendCalls
		audioCh <- secondChunk
		<-sendCalls
		sendStreamCh <- stream2
		<-closeSendCalls
		audioCh <- thirdChunk
		<-sendCalls

		cancel()
		wg.Wait()

		if !errors.Is(got, context.Canceled) {
			t.Errorf("audioSender.Start() error = %v, want %v", got, context.Canceled)
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
