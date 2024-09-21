package recognize

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate moq -rm -out response_receiver_mock.go . ResponseReceiver
type ResponseReceiver interface {
	Start(ctx context.Context) error
}

type responseReceiver struct {
	responseCh      chan<- *speechpb.StreamingRecognizeResponse
	receiveStreamCh <-chan speechpb.Speech_StreamingRecognizeClient
}

func NewResponseReceiver(
	responseCh chan<- *speechpb.StreamingRecognizeResponse,
	receiveStreamCh <-chan speechpb.Speech_StreamingRecognizeClient,
) ResponseReceiver {
	return &responseReceiver{
		responseCh:      responseCh,
		receiveStreamCh: receiveStreamCh,
	}
}

func (r *responseReceiver) Start(ctx context.Context) error {
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
				// 送信側で stream が閉じられると、受信側は最後のレスポンスのあと EOF を受信する。
				// そのタイミングで新しい stream に切り替える。
				newStream, ok := <-r.receiveStreamCh
				if !ok {
					return fmt.Errorf("failed to get new receive stream from channel")
				}
				stream = newStream
				continue
			}
			if err != nil {
				// コマンドを SIGINT で終了した際に context canceled error が発生するため、無視する。
				if status.Code(err) == codes.Canceled {
					return nil
				}
				return fmt.Errorf("failed to receive response: %w", err)
			}
			r.responseCh <- resp
		}
	}
}
