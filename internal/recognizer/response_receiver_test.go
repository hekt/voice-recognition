package recognizer

import (
	"context"
	"io"
	"reflect"
	"testing"

	"cloud.google.com/go/speech/apiv2/speechpb"
	ispeechpb "github.com/hekt/voice-recognition/internal/interfaces/speechpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewResponseReceiver(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		responseCh := make(chan *speechpb.StreamingRecognizeResponse, 1)
		receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)

		got := NewResponseReceiver(responseCh, receiveStreamCh)
		want := &ResponseReceiver{
			responseCh:      responseCh,
			receiveStreamCh: receiveStreamCh,
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("NewResponseReceiver() = %v, want %v", got, want)
		}
	})
}

func Test_responseReceiver_Start(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		type response struct {
			resp *speechpb.StreamingRecognizeResponse
			err  error
		}
		response1 := &speechpb.StreamingRecognizeResponse{}
		stream1ResponseCh := make(chan *response, 2)
		stream1ResponseCh <- &response{response1, nil}
		stream1ResponseCh <- &response{nil, io.EOF}
		close(stream1ResponseCh)
		stream1 := &ispeechpb.Speech_StreamingRecognizeClientMock{
			RecvFunc: func() (*speechpb.StreamingRecognizeResponse, error) {
				res := <-stream1ResponseCh
				return res.resp, res.err
			},
		}

		response2 := &speechpb.StreamingRecognizeResponse{}
		stream2ResponseCh := make(chan *response, 2)
		stream2ResponseCh <- &response{response2, nil}
		stream2ResponseCh <- &response{nil, status.Error(codes.Canceled, "canceled")}
		close(stream2ResponseCh)
		stream2 := &ispeechpb.Speech_StreamingRecognizeClientMock{
			RecvFunc: func() (*speechpb.StreamingRecognizeResponse, error) {
				res := <-stream2ResponseCh
				return res.resp, res.err
			},
		}

		responseCh := make(chan *speechpb.StreamingRecognizeResponse, 10)
		receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 2)
		receiveStreamCh <- stream1
		receiveStreamCh <- stream2

		r := &ResponseReceiver{
			responseCh:      responseCh,
			receiveStreamCh: receiveStreamCh,
		}

		got := r.Start(context.Background())

		if got != nil {
			t.Errorf("unexpected error: %v", got)
		}
		if len(responseCh) != 2 {
			t.Errorf("unexpected number of responses: got %d, want %d", len(responseCh), 2)
			return
		}
		if got, want := <-responseCh, response1; got != want {
			t.Errorf("unexpected response: got %v, want %v", got, want)
		}
		if got, want := <-responseCh, response2; got != want {
			t.Errorf("unexpected response: got %v, want %v", got, want)
		}
	})

	t.Run("closed stream", func(t *testing.T) {
		receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 2)
		close(receiveStreamCh)

		r := &ResponseReceiver{
			receiveStreamCh: receiveStreamCh,
		}

		if got := r.Start(context.Background()); got == nil {
			t.Errorf("Start() error = %v, want non-nil", got)
		}
	})
}
