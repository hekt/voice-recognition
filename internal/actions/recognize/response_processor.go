package recognize

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/speech/apiv2/speechpb"
)

type ResponseProcessor interface {
	Start(ctx context.Context) error
}

type responseProcessor struct {
	resultWriter  io.Writer
	interimWriter io.Writer
	responseCh    <-chan *speechpb.StreamingRecognizeResponse
}

func NewResponseProcessor(
	resultWriter io.Writer,
	interimWriter io.Writer,
	responseCh <-chan *speechpb.StreamingRecognizeResponse,
) ResponseProcessor {
	return &responseProcessor{
		resultWriter:  resultWriter,
		interimWriter: interimWriter,
		responseCh:    responseCh,
	}
}

func (p *responseProcessor) Start(ctx context.Context) error {
	var buf bytes.Buffer
	var interimResult []byte
	for {
		select {
		case <-ctx.Done():
			// 終了する前に確定していない中間結果を書き込む。
			if len(interimResult) > 0 {
				if _, err := p.resultWriter.Write(interimResult); err != nil {
					return fmt.Errorf("failed to write interim result: %w", err)
				}
			}
			return nil
		case resp, ok := <-p.responseCh:
			if !ok {
				return nil
			}

			// レスポンス処理
			buf.Reset()
			for _, result := range resp.Results {
				s := result.Alternatives[0].Transcript
				if result.IsFinal {
					if _, err := p.resultWriter.Write([]byte(s)); err != nil {
						return fmt.Errorf("failed to write result: %w", err)
					}
					interimResult = []byte{}
					buf.Reset()
					continue
				}
				buf.WriteString(s)
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
