package recognize

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
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
	ProjectID      string
	RecognizerName string
	OutputFilePath string
}

func Run(ctx context.Context, args Args) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	timer := time.NewTimer(streamTimeout - streamTimeoutOffset)
	defer timer.Stop()

	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	stream, err := initializeStream(ctx, client, args.ProjectID, args.RecognizerName)
	if err != nil {
		log.Fatal(err)
	}

	// sendStream を終了 (CloseSend) したあともレスポンスが送信されることから、それぞれ置き換えるタイミングが違う。
	// そのため sendStream と receiveStream を別々に持つ。
	sendStream := stream
	receiveStream := stream

	// レスポンスデータを受信する channel
	// stream から Recv する goroutine と、その結果を表示する goroutine で使う。
	responses := make(chan *speechpb.StreamingRecognizeResponse)

	// 再接続をトリガーする channel
	// メインループでトリガーされ stream に送信する goroutine で再接続する。
	reconnect := make(chan struct{})

	// 標準入力から受け取った音声データを gRPC Stream に送信する goroutine
	go func() {
		buf := make([]byte, 8192)
		for {
			select {
			case <-reconnect:
				// CloseSend を送ったあともレスポンスは送信されるため、この時点では receiveStream は置き換えない。
				// receiveStream 側で EOF を受信したときに置き換える。
				if err := sendStream.CloseSend(); err != nil {
					log.Fatalf("Could not close stream: %v", err)
				}
				sendStream, err = initializeStream(ctx, client, args.ProjectID, args.RecognizerName)
				if err != nil {
					log.Fatalf("Could not create stream: %v", err)
				}
			default:
				n, err := os.Stdin.Read(buf)
				if n > 0 {
					err := sendStream.Send(&speechpb.StreamingRecognizeRequest{
						StreamingRequest: &speechpb.StreamingRecognizeRequest_Audio{
							Audio: buf[:n],
						},
					})
					if err != nil {
						log.Printf("Could not send audio: %v", err)
					}
				}
				if err == io.EOF {
					if err := sendStream.CloseSend(); err != nil {
						log.Fatalf("Could not close stream: %v", err)
					}
					return
				}
				if err != nil {
					log.Printf("Could not read from stdin: %v", err)
					continue
				}
			}
		}
	}()

	// gRPC Stream から結果を受信して channel に送信する goroutine
	go func() {
		defer close(responses)
		for {
			resp, err := receiveStream.Recv()
			if err == io.EOF {
				// CloseSend を送信したあとに EOF を受信したとき、新しいストリームを受信する。
				if receiveStream == sendStream {
					fmt.Println("no new stream")
					break
				}
				receiveStream = sendStream
				continue
			}
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					log.Fatalf("Cannot stream results: %v", err)
				}
			}
			responses <- resp
		}
	}()

	var sb strings.Builder
	var interimResult string
	for {
		select {
		case <-ctx.Done():
			// SIGINT が送信されたとき、結果をファイルに書き込んで終了する。
			if interimResult != "" {
				writeResultToFile(interimResult, args.OutputFilePath)
			}
			return
		case <-timer.C:
			// 一定時間ごとに再接続する。
			timer.Reset(streamTimeout - streamTimeoutOffset)
			reconnect <- struct{}{}
		case resp := <-responses:
			// レスポンス処理
			sb.Reset()
			for _, result := range resp.Results {
				s := result.Alternatives[0].Transcript
				sb.WriteString(s)
				if result.IsFinal {
					writeResultToFile(s, args.OutputFilePath)
					interimResult = ""
					break
				}
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

func writeResultToFile(s string, path string) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(s + "\n"); err != nil {
		log.Fatal(err)
	}
}

func initializeStream(
	ctx context.Context,
	client *speech.Client,
	projectID string,
	recognizerName string,
) (speechpb.Speech_StreamingRecognizeClient, error) {
	stream, err := client.StreamingRecognize(ctx)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return stream, nil
}
