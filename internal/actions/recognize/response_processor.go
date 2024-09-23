package recognize

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"cloud.google.com/go/speech/apiv2/speechpb"
)

//go:generate moq -rm -out response_processor_mock.go . ResponseProcessorInterface
type ResponseProcessorInterface interface {
	Start(ctx context.Context) error
}

var _ ResponseProcessorInterface = (*ResponseProcessor)(nil)

type ResponseProcessor struct {
	resultWriter  io.Writer
	interimWriter io.Writer
	responseCh    <-chan *speechpb.StreamingRecognizeResponse
}

func NewResponseProcessor(
	resultWriter io.Writer,
	interimWriter io.Writer,
	responseCh <-chan *speechpb.StreamingRecognizeResponse,
) *ResponseProcessor {
	return &ResponseProcessor{
		resultWriter:  resultWriter,
		interimWriter: interimWriter,
		responseCh:    responseCh,
	}
}

func (p *ResponseProcessor) Start(ctx context.Context) error {
	slog.Debug("ResponseProcessor: start")

	var buf bytes.Buffer
	var interimResult []byte
	defer func() {
		if len(interimResult) == 0 {
			return
		}
		if _, err := p.resultWriter.Write(interimResult); err != nil {
			slog.Error(fmt.Sprintf("failed to write interim result: %v", err))
		}
		slog.Debug("ResponseProcessor: interim result written")
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case resp, ok := <-p.responseCh:
			if !ok {
				return fmt.Errorf("failed to get response from channel")
			}

			// process response
			buf.Reset()
			for _, result := range resp.Results {
				s := result.Alternatives[0].Transcript

				if !result.IsFinal {
					buf.WriteString(s)
					continue
				}

				slog.Debug("ResponseProcessor: final result received")

				if _, err := p.resultWriter.Write([]byte(s)); err != nil {
					return fmt.Errorf("failed to write result: %w", err)
				}
				interimResult = nil
				buf.Reset()
			}

			if buf.Len() == 0 {
				continue
			}

			interimResult = buf.Bytes()
			if _, err := p.interimWriter.Write(interimResult); err != nil {
				return fmt.Errorf("failed to write interim result: %w", err)
			}
		}
	}
}
