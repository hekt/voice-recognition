package recognize

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
)

type Args struct {
	ProjectID      string
	RecognizerName string
	OutputFilePath string
}

func Run(ctx context.Context, args Args) {
	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	stream, err := createStream(ctx, client, args.ProjectID, args.RecognizerName)
	if err != nil {
		log.Fatal(err)
	}

	// sendStream を先に終了させるために、receiveStream と sendStream を分ける。
	// reconnect が true のときは sendStream を新しい stream に切り替える。
	// receiveStream は sendStream の最後のレスポンスで io.EOF を受け取るので、
	// そのタイミングで sendStream に切り替え、結果を受信する
	sendStream := stream
	receiveStream := stream

	reconnect := &atomic.Bool{}
	reconnect.Store(false)

	// sigint されたときの中間結果をファイルに書き出すために使う。
	// 更新するのはメインループ内のみなので atomic でなくてもよい。
	var interimResult string

	// 音声データを受信する channel
	// stream から Recv する goroutine と、その結果を表示する goroutine で使う。
	responses := make(chan *speechpb.StreamingRecognizeResponse)

	// sigint されたときに interrimResult をファイルに書き出してから終了する goroutine
	go func() {
		trap := make(chan os.Signal, 1)
		signal.Notify(trap, os.Interrupt)
		sig := <-trap
		fmt.Printf("Received signal: %v\n", sig)
		if interimResult != "" {
			write(interimResult, args.OutputFilePath)
		}
		os.Exit(0)
	}()

	// 標準入力から受け取った音声データを gRPC Stream に送信する goroutine
	go func() {
		buf := make([]byte, 8192)
		for {
			if reconnect.Load() {
				if err := sendStream.CloseSend(); err != nil {
					log.Fatalf("Could not close stream: %v", err)
				}
				sendStream, err = createStream(ctx, client, args.ProjectID, args.RecognizerName)
				if err != nil {
					log.Fatalf("Could not create stream: %v", err)
				}
				reconnect.Store(false)
			}
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
	}()

	// gRPC Stream から結果を受信して channel に送信する goroutine
	go func() {
		defer close(responses)
		for {
			resp, err := receiveStream.Recv()
			if err == io.EOF {
				fmt.Println("Stream closed. reconnectiong...")
				if receiveStream == sendStream {
					fmt.Println("no new stream")
					break
				}
				receiveStream = sendStream
				continue
			}
			if err != nil {
				log.Fatalf("Cannot stream results: %v", err)
			}
			responses <- resp
		}
	}()

	// メインループ。結果を受信して表示する。
	var sb strings.Builder
	for resp := range responses {
		sb.Reset()
		for _, result := range resp.Results {
			s := result.Alternatives[0].Transcript
			sb.WriteString(s)
			if result.IsFinal {
				write(s, args.OutputFilePath)
				reconnect.Store(true)
				interimResult = ""
				break
			}
		}

		if sb.Len() == 0 {
			continue
		}

		t := sb.String()
		if len(t) != len(interimResult) {
			t = sb.String()
			// clear screen
			fmt.Print("\033[H\033[2J")
			// 緑で表示
			fmt.Print("\033[32m")
			fmt.Print(t)
			fmt.Print("\033[0m")
			fmt.Print("\n")
		}

		interimResult = t
	}
}

func write(s string, path string) {
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

func splitInterimResult(current string, previous string) (string, string) {
	c := []rune(current)
	p := []rune(previous)

	minLen := min(len(c), len(p))
	var i int
	for i = 0; i < minLen; i++ {
		if c[i] != p[i] {
			break
		}
	}

	prefix := string(c[:i])
	rest := string(p[i:])
	return prefix, rest
}

func createStream(
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
