package recognize

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hekt/voice-recognition/util"
)

type recognizer struct {
	projectID         string
	recognizerName    string
	reconnectInterval time.Duration
	bufferSize        int

	// Speech-to-Text API のクライアント
	client *speech.Client
	// 音声データの入力元。標準入力を想定。
	audioReader io.Reader
	// 確定した結果の出力先。ファイルを想定。
	resultWriter io.Writer
	// 中間結果の出力先。ANSI エスケープシーケンスを使っているため実質的には標準出力のみ。
	interimWriter io.Writer

	// レスポンスデータを受信する channel
	// stream から Recv する goroutine と、その結果を表示する goroutine で使う。
	responseCh chan *speechpb.StreamingRecognizeResponse
	// 再接続をトリガーする channel
	// メインループでトリガーされ stream に送信する goroutine で再接続する。
	reconnectCh chan struct{}
	// stream を受け渡しするための channel。
	// audioStream の goroutine と receiveStream の grooutine それぞれで扱うため、2つ用意する。
	// 最初にそれぞれの goroutine から取り出したあとは基本的には最大でも1つになるはず。
	newStreamCh chan speechpb.Speech_StreamingRecognizeClient
}

func newRecognizer(
	ctx context.Context,
	projectID string,
	recognizerName string,
	reconnectInterval time.Duration,
	bufferSize int,
	audioReader io.Reader,
	resultWriter io.Writer,
	interimWriter io.Writer,
) (*recognizer, error) {
	if projectID == "" {
		return nil, errors.New("project ID must be specified")
	}
	if recognizerName == "" {
		return nil, errors.New("recognizer name must be specified")
	}
	if bufferSize < 1024 {
		return nil, errors.New("buffer size must be greater than or equal to 1024")
	}
	if reconnectInterval < time.Minute {
		return nil, errors.New("reconnect interval must be greater than or equal to 1 minute")
	}
	if audioReader == nil {
		return nil, errors.New("audio reader must be specified")
	}
	if resultWriter == nil {
		return nil, errors.New("result writer must be specified")
	}
	if interimWriter == nil {
		return nil, errors.New("interim writer must be specified")
	}

	client, err := speech.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create speech client: %w", err)
	}

	return &recognizer{
		projectID:         projectID,
		recognizerName:    recognizerName,
		reconnectInterval: reconnectInterval,
		bufferSize:        bufferSize,

		audioReader:   audioReader,
		resultWriter:  resultWriter,
		interimWriter: interimWriter,

		client: client,

		responseCh:  make(chan *speechpb.StreamingRecognizeResponse),
		reconnectCh: make(chan struct{}),
		newStreamCh: make(chan speechpb.Speech_StreamingRecognizeClient, 2),
	}, nil
}

func (r *recognizer) Start(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	defer r.client.Close()

	stream, err := r.initializeStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize stream: %w", err)
	}
	r.newStreamCh <- stream
	r.newStreamCh <- stream

	eg, ctx := errgroup.WithContext(ctx)

	// 標準入力から受け取った音声データを gRPC Stream に送信する。
	// reconnectCh からのトリガーで stream の再接続も行い、新しい stream を newStreamCh に送信する。
	eg.Go(func() error {
		defer close(r.newStreamCh)
		return r.startAudioSender(ctx)
	})

	// gRPC Stream から結果を受信して responseCh に送信する。
	eg.Go(func() error {
		defer close(r.responseCh)
		return r.startResponseReceiver(ctx)
	})

	// response channel から結果を受信して標準出力やファイルに出力する。
	eg.Go(func() error {
		return r.startResponseProcessor(ctx)
	})

	// 一定時間ごとに再接続をトリガーするタイマー。
	eg.Go(func() error {
		defer close(r.reconnectCh)
		return r.startTimer(ctx)
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (r *recognizer) startAudioSender(ctx context.Context) error {
	stream, ok := <-r.newStreamCh
	if !ok {
		return fmt.Errorf("failed to get stream from channel")
	}

	buf := make([]byte, r.bufferSize)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-r.reconnectCh:
			if err := stream.CloseSend(); err != nil {
				return fmt.Errorf("failed to close send direction of stream: %w", err)
			}
			newStream, err := r.initializeStream(ctx)
			if err != nil {
				return fmt.Errorf("failed to initialize stream: %w", err)
			}

			// CloseSend を送ったあともサーバー側からレスポンスは送信されるため、
			// この時点では受信側 (startResponseReceiver) では stream を切り替えない。
			// 受信側で EOF を受信したときに newStreamCh から取り出して切り替える
			stream = newStream
			r.newStreamCh <- newStream
		default:
			n, err := r.audioReader.Read(buf)
			if n > 0 {
				err := stream.Send(&speechpb.StreamingRecognizeRequest{
					StreamingRequest: &speechpb.StreamingRecognizeRequest_Audio{
						Audio: buf[:n],
					},
				})
				if err != nil {
					return fmt.Errorf("failed to send audio data: %w", err)
				}
			}
			if err == io.EOF {
				if e := stream.CloseSend(); err != nil {
					return e
				}
				return nil
			}
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
		}
	}
}

func (r *recognizer) startResponseReceiver(ctx context.Context) error {
	stream, ok := <-r.newStreamCh
	if !ok {
		return fmt.Errorf("failed to get stream from channel")
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
				stream, ok = <-r.newStreamCh
				if !ok {
					return fmt.Errorf("failed to get stream from channel")
				}
				continue
			}
			if err != nil {
				// context canceled はコマンドを SIGINT で終了した際に発生するため、無視する。
				if status.Code(err) == codes.Canceled {
					return nil
				}
				return fmt.Errorf("failed to receive response: %w", err)
			}
			r.responseCh <- resp
		}
	}
}

func (r *recognizer) startResponseProcessor(ctx context.Context) error {
	var buf bytes.Buffer
	var interimResult []byte
	for {
		select {
		case <-ctx.Done():
			// 終了する前に確定していない中間結果を書き込む。
			if len(interimResult) > 0 {
				if err := r.writeResult(interimResult); err != nil {
					return fmt.Errorf("failed to write interim result to file: %w", err)
				}
			}
			return nil
		case resp, ok := <-r.responseCh:
			if !ok {
				return nil
			}

			// レスポンス処理
			buf.Reset()
			for _, result := range resp.Results {
				s := result.Alternatives[0].Transcript
				if result.IsFinal {
					if err := r.writeResult([]byte(s)); err != nil {
						return fmt.Errorf("failed to write result to file: %w", err)
					}
					interimResult = []byte{}
					break
				}
				buf.WriteString(s)
			}

			if buf.Len() == 0 {
				continue
			}

			interimResult = buf.Bytes()
			if err := r.writeInterim(interimResult); err != nil {
				return fmt.Errorf("failed to write interim result: %w", err)
			}
		}
	}
}

func (r *recognizer) writeResult(b []byte) error {
	bln := append(b, []byte("\n")...)
	if _, err := r.resultWriter.Write(bln); err != nil {
		return fmt.Errorf("failed to write result: %w", err)
	}

	return nil
}

var (
	clearScreen = []byte("\033[H\033[2J")
	greenColor  = []byte("\033[32m")
	resetColor  = []byte("\033[0m")
)

func (r *recognizer) writeInterim(b []byte) error {
	buf := bytes.Buffer{}
	buf.Write(clearScreen)
	buf.Write(greenColor)
	buf.Write(b)
	buf.Write(resetColor)

	if _, err := r.interimWriter.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write interim: %w", err)
	}

	return nil
}

func (r *recognizer) startTimer(ctx context.Context) error {
	timer := time.NewTimer(r.reconnectInterval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			// 一定時間ごとに再接続する。
			timer.Reset(r.reconnectInterval)
			r.reconnectCh <- struct{}{}
		}
	}
}

func (r *recognizer) initializeStream(
	ctx context.Context,
) (speechpb.Speech_StreamingRecognizeClient, error) {
	stream, err := r.client.StreamingRecognize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	if err := stream.Send(&speechpb.StreamingRecognizeRequest{
		Recognizer: util.RecognizerFullname(r.projectID, r.recognizerName),
		StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
			StreamingConfig: &speechpb.StreamingRecognitionConfig{
				StreamingFeatures: &speechpb.StreamingRecognitionFeatures{
					InterimResults: true,
				},
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to send initial request: %w", err)
	}

	return stream, nil
}
