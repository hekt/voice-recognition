package recognize

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
	"golang.org/x/sync/errgroup"
)

type recognizer struct {
	projectID         string
	recognizerName    string
	outputFilePath    string
	reconnectInterval time.Duration
	bufferSize        int

	// Speech-to-Text API のクライアント
	client *speech.Client
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
	projectID, recognizerName, outputFilePath string,
	bufferSize int,
	reconnectInterval time.Duration,
) (*recognizer, error) {
	if projectID == "" {
		return nil, errors.New("project ID must be specified")
	}
	if recognizerName == "" {
		return nil, errors.New("recognizer name must be specified")
	}
	if outputFilePath == "" {
		return nil, errors.New("output file path must be specified")
	}
	if bufferSize < 1024 {
		return nil, errors.New("buffer size must be greater than or equal to 1024")
	}
	if reconnectInterval < time.Minute {
		return nil, errors.New("reconnect interval must be greater than or equal to 1 minute")
	}

	client, err := speech.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create speech client: %w", err)
	}

	return &recognizer{
		projectID:         projectID,
		recognizerName:    recognizerName,
		outputFilePath:    outputFilePath,
		reconnectInterval: reconnectInterval,
		bufferSize:        bufferSize,

		client:      client,
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

func (r *recognizer) startResponseProcessor(ctx context.Context) error {
	var sb strings.Builder
	var interimResult string
	for {
		select {
		case <-ctx.Done():
			// 最後に中間結果をファイルに書き込む。
			if interimResult != "" {
				if err := r.writeResultToFile(interimResult); err != nil {
					return fmt.Errorf("failed to write interim result to file: %w", err)
				}
			}
			return nil
		case resp, ok := <-r.responseCh:
			if !ok {
				return nil
			}

			// レスポンス処理
			sb.Reset()
			for _, result := range resp.Results {
				s := result.Alternatives[0].Transcript
				if result.IsFinal {
					if err := r.writeResultToFile(s); err != nil {
						return fmt.Errorf("failed to write result to file: %w", err)
					}
					interimResult = ""
					break
				}
				sb.WriteString(s)
			}

			if sb.Len() == 0 {
				continue
			}

			t := sb.String()

			// clear screen
			fmt.Print("\033[H\033[2J")
			// 緑で表示
			fmt.Print("\033[32m")
			fmt.Print(t)
			fmt.Print("\033[0m")
			fmt.Print("\n")

			interimResult = t
		}
	}
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

			// CloseSend を送ったあともサーバー側からレスポンスは送信されるため、この時点では受信側では stream を切り替えない。
			// 受信側で EOF を受信したときに newStreamCh から取り出して置き換える。
			stream = newStream
			r.newStreamCh <- newStream
		default:
			n, err := os.Stdin.Read(buf)
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
				// 送信側でストリームが閉じられると、受信側は最後のレスポンスのあと EOF を受信する。
				// そのタイミングで新しいストリームに切り替える。
				stream, ok = <-r.newStreamCh
				if !ok {
					return fmt.Errorf("failed to get stream from channel")
				}
				continue
			}
			if err != nil {
				return fmt.Errorf("failed to receive response: %w", err)
			}
			r.responseCh <- resp
		}
	}
}

func (r *recognizer) writeResultToFile(s string) error {
	file, err := os.OpenFile(r.outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(s + "\n"); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func (r *recognizer) initializeStream(
	ctx context.Context,
) (speechpb.Speech_StreamingRecognizeClient, error) {
	stream, err := r.client.StreamingRecognize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	if err := stream.Send(&speechpb.StreamingRecognizeRequest{
		Recognizer: fmt.Sprintf(
			"projects/%s/locations/global/recognizers/%s",
			r.projectID,
			r.recognizerName,
		),
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
