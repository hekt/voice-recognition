package recognize

import (
	"context"
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

const (
	// streamTimeout はストリームのタイムアウト時間を表す。
	// この時間を超えるとサーバー側からストリームが切断される。
	streamTimeout = 5 * time.Minute

	// streamTimeoutOffset はストリームのタイムアウト時間のオフセットを表す。
	// この時間だけ短く設定することで、ストリームが切断される前に再接続を試みる。
	streamTimeoutOffset = 10 * time.Second
)

type Args struct {
	ProjectID         string
	RecognizerName    string
	OutputFilePath    string
	ReconnectInterval time.Duration
}

func Run(ctx context.Context, args Args) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	client, err := speech.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create speech client: %w", err)
	}
	defer client.Close()

	stream, err := initializeStream(ctx, client, args.ProjectID, args.RecognizerName)
	if err != nil {
		return fmt.Errorf("failed to initialize stream: %w", err)
	}

	// レスポンスデータを受信する channel
	// stream から Recv する goroutine と、その結果を表示する goroutine で使う。
	responsesCh := make(chan *speechpb.StreamingRecognizeResponse)

	// 再接続をトリガーする channel
	// メインループでトリガーされ stream に送信する goroutine で再接続する。
	reconnectCh := make(chan struct{})

	// stream を受け渡しするための channel。
	// audioStream の goroutine と receiveStream の grooutine それぞれで扱うため、2つ用意する。
	// 最初にそれぞれの goroutine から取り出したあとは基本的には最大でも1つになるはず。
	newStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 2)
	newStreamCh <- stream
	newStreamCh <- stream

	reconnectInterval := args.ReconnectInterval
	if reconnectInterval == 0 {
		reconnectInterval = streamTimeout - streamTimeoutOffset
	}

	eg, egCtx := errgroup.WithContext(ctx)

	// 標準入力から受け取った音声データを gRPC Stream に送信する goroutine
	eg.Go(func() error {
		return startAudioSender(
			egCtx,
			args.ProjectID,
			args.RecognizerName,
			client,
			reconnectCh,
			newStreamCh,
		)
	})

	// gRPC Stream から結果を受信して channel に送信する goroutine
	eg.Go(func() error {
		return startResponseReceiver(egCtx, responsesCh, newStreamCh)
	})

	// メインループ。
	eg.Go(func() error {
		return startMainLoop(egCtx, args.OutputFilePath, reconnectInterval, reconnectCh, responsesCh)
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func startMainLoop(
	ctx context.Context,
	outputFilePath string,
	reconnectInterval time.Duration,
	reconnectCh chan struct{},
	responsesCh chan *speechpb.StreamingRecognizeResponse,
) error {
	timer := time.NewTimer(reconnectInterval)
	defer timer.Stop()

	var sb strings.Builder
	var interimResult string

	for {
		select {
		case <-ctx.Done():
			// SIGINT が送信されたとき、結果をファイルに書き込んで終了する。
			if interimResult != "" {
				if err := writeResultToFile(interimResult, outputFilePath); err != nil {
					return fmt.Errorf("failed to write interim result to file: %w", err)
				}
			}
			return nil
		case <-timer.C:
			// 一定時間ごとに再接続する。
			timer.Reset(reconnectInterval)
			reconnectCh <- struct{}{}
		case resp := <-responsesCh:
			// レスポンス処理
			sb.Reset()
			for _, result := range resp.Results {
				s := result.Alternatives[0].Transcript
				if result.IsFinal {
					if err := writeResultToFile(s, outputFilePath); err != nil {
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

func startAudioSender(
	ctx context.Context,
	projectID string,
	recognizerName string,
	client *speech.Client,
	reconnectCh chan struct{},
	newStreamCh chan speechpb.Speech_StreamingRecognizeClient,
) error {
	stream := <-newStreamCh

	buf := make([]byte, 8192)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-reconnectCh:
			if err := stream.CloseSend(); err != nil {
				return fmt.Errorf("failed to close send direction of stream: %w", err)
			}
			newStream, err := initializeStream(ctx, client, projectID, recognizerName)
			if err != nil {
				return fmt.Errorf("failed to initialize stream: %w", err)
			}

			// CloseSend を送ったあともサーバー側からレスポンスは送信されるため、この時点では受信側では stream を切り替えない。
			// 受信側で EOF を受信したときに newStreamCh から取り出して置き換える。
			stream = newStream
			newStreamCh <- newStream
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

func startResponseReceiver(
	ctx context.Context,
	responsesCh chan *speechpb.StreamingRecognizeResponse,
	newStreamCh chan speechpb.Speech_StreamingRecognizeClient,
) error {
	defer close(responsesCh)

	stream := <-newStreamCh

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			resp, err := stream.Recv()
			if err == io.EOF {
				// 送信側でストリームが閉じられると、受信側は最後のレスポンスのあと EOF を受信する。
				// そのタイミングで新しいストリームに切り替える。
				stream = <-newStreamCh
				continue
			}
			if err != nil {
				return fmt.Errorf("failed to receive response: %w", err)
			}
			responsesCh <- resp
		}
	}
}

func writeResultToFile(s string, path string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(s + "\n"); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func initializeStream(
	ctx context.Context,
	client *speech.Client,
	projectID string,
	recognizerName string,
) (speechpb.Speech_StreamingRecognizeClient, error) {
	stream, err := client.StreamingRecognize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	if err := stream.Send(&speechpb.StreamingRecognizeRequest{
		Recognizer: fmt.Sprintf(
			"projects/%s/locations/global/recognizers/%s",
			projectID,
			recognizerName,
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
