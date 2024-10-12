package recognizer

import (
	"context"
	"fmt"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/hekt/voice-recognition/internal/recognizer/model"
)

//go:generate moq -rm -out response_processor_mock.go . ResponseProcessorInterface
type ResponseProcessorInterface interface {
	Start(ctx context.Context) error
}

var _ ResponseProcessorInterface = (*ResponseProcessor)(nil)

type ResponseProcessor struct {
	responseCh <-chan *speechpb.StreamingRecognizeResponse
	resultCh   chan<- []*model.Result
	processCh  chan<- struct{}
}

func NewResponseProcessor(
	responseCh <-chan *speechpb.StreamingRecognizeResponse,
	resultCh chan<- []*model.Result,
	processCh chan<- struct{},
) *ResponseProcessor {
	return &ResponseProcessor{
		responseCh: responseCh,
		resultCh:   resultCh,
		processCh:  processCh,
	}
}

func (p *ResponseProcessor) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case resp, ok := <-p.responseCh:
			if !ok {
				return fmt.Errorf("response channel is closed")
			}

			// process response
			results := make([]*model.Result, 0, len(resp.Results))
			for _, result := range resp.Results {
				if len(result.Alternatives) == 0 {
					continue
				}

				results = append(results, &model.Result{
					Transcript: result.Alternatives[0].Transcript,
					IsFinal:    result.IsFinal,
				})
			}

			if len(results) == 0 {
				continue
			}

			p.resultCh <- results
			p.processCh <- struct{}{}
		}
	}
}
