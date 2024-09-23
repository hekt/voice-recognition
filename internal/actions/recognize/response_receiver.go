package recognize

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate moq -rm -out response_receiver_mock.go . ResponseReceiverInterface
type ResponseReceiverInterface interface {
	Start(ctx context.Context) error
}

var _ ResponseReceiverInterface = (*ResponseReceiver)(nil)

type ResponseReceiver struct {
	responseCh      chan<- *speechpb.StreamingRecognizeResponse
	receiveStreamCh <-chan speechpb.Speech_StreamingRecognizeClient
}

func NewResponseReceiver(
	responseCh chan<- *speechpb.StreamingRecognizeResponse,
	receiveStreamCh <-chan speechpb.Speech_StreamingRecognizeClient,
) *ResponseReceiver {
	return &ResponseReceiver{
		responseCh:      responseCh,
		receiveStreamCh: receiveStreamCh,
	}
}

func (r *ResponseReceiver) Start(ctx context.Context) error {
	slog.Debug("ResponseReceiver: start")

	stream, ok := <-r.receiveStreamCh
	if !ok {
		return fmt.Errorf("failed to get receive stream from channel")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			resp, err := stream.Recv()
			if err == io.EOF {
				// when the stream is closed by the sender, the receiver will receive EOF after the final response.
				// at that time, switch to new stream.
				slog.Debug("ResponseReceiver: EOF received")

				select {
				case newStream, ok := <-r.receiveStreamCh:
					if !ok {
						return fmt.Errorf("failed to get new receive stream from channel")
					}
					stream = newStream
					slog.Debug("ResponseReceiver: stream switched")
					continue
				case <-ctx.Done():
					return nil
				}
			}
			if status.Code(err) == codes.Canceled {
				// when the command is terminated by SIGINT, context canceled occurs in the stream.
				// status.Code(err) returns codes.OK if err is nil.
				return nil
			}
			if err != nil {
				return fmt.Errorf("failed to receive response: %w", err)
			}

			r.responseCh <- resp
		}
	}
}
