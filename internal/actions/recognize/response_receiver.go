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
				slog.Debug("ResponseReceiver: EOF received")

				// 送信側で stream が閉じられると、受信側は最後のレスポンスのあと EOF を受信する。
				// そのタイミングで新しい stream に切り替える。
				newStream, ok := <-r.receiveStreamCh
				if !ok {
					return fmt.Errorf("failed to get new receive stream from channel")
				}

				stream = newStream
				slog.Debug("ResponseReceiver: stream switched")
				continue
			}
			if status.Code(err) == codes.Canceled {
				// コマンドを SIGINT で終了した際に canceled が発生するため、無視して終了する。
				// なお status.Code(err) は err が nil の場合は codes.OK を返す。
				return nil
			}
			if err != nil {
				return fmt.Errorf("failed to receive response: %w", err)
			}

			r.responseCh <- resp
		}
	}
}
