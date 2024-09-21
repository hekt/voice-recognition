package recognize

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"

	"cloud.google.com/go/speech/apiv2/speechpb"
	pkgspeechpb "github.com/hekt/voice-recognition/pkg/speechpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewResponseReceiver(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		responseCh := make(chan *speechpb.StreamingRecognizeResponse, 1)
		receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)

		got := NewResponseReceiver(responseCh, receiveStreamCh)
		want := &responseReceiver{
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
		stream1RecvCalls := 0
		stream1Responses := []response{
			{&speechpb.StreamingRecognizeResponse{}, nil},
			{nil, io.EOF},
			{nil, errors.New("unexpected error")}, // ハンドリングができていればこのレスポンスは取得されない
		}
		stream1 := &pkgspeechpb.Speech_StreamingRecognizeClientMock{
			RecvFunc: func() (*speechpb.StreamingRecognizeResponse, error) {
				if stream1RecvCalls > len(stream1Responses) {
					t.Fatalf("unexpected call to Recv on stream1: %d", stream1RecvCalls)
				}
				res := stream1Responses[stream1RecvCalls]
				stream1RecvCalls++
				return res.resp, res.err
			},
		}

		stream2RecvCalls := 0
		stream2Responses := []response{
			{&speechpb.StreamingRecognizeResponse{}, nil},
			{nil, status.Error(codes.Canceled, "canceled")},
		}
		stream2 := &pkgspeechpb.Speech_StreamingRecognizeClientMock{
			RecvFunc: func() (*speechpb.StreamingRecognizeResponse, error) {
				if stream2RecvCalls > len(stream2Responses) {
					t.Fatalf("unexpected call to Recv on stream2: %d", stream2RecvCalls)
				}
				res := stream2Responses[stream2RecvCalls]
				stream2RecvCalls++
				return res.resp, res.err
			},
		}

		responseCh := make(chan *speechpb.StreamingRecognizeResponse, 10)
		receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 2)
		receiveStreamCh <- stream1
		receiveStreamCh <- stream2

		r := &responseReceiver{
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
		if got, want := <-responseCh, stream1Responses[0].resp; got != want {
			t.Errorf("unexpected response: got %v, want %v", got, want)
		}
		if got, want := <-responseCh, stream2Responses[0].resp; got != want {
			t.Errorf("unexpected response: got %v, want %v", got, want)
		}
	})

	t.Run("closed stream", func(t *testing.T) {
		receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 2)
		close(receiveStreamCh)

		r := &responseReceiver{
			receiveStreamCh: receiveStreamCh,
		}

		if got := r.Start(context.Background()); got == nil {
			t.Errorf("Start() error = %v, want non-nil", got)
		}
	})
}
