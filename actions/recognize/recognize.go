package recognize

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

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
	reconnect := false

	// 標準入力から受け取った音声データを gRPC Stream に送信する goroutine
	go func() {
		buf := make([]byte, 1024)
		for {
			if reconnect {
				if err := sendStream.CloseSend(); err != nil {
					log.Fatalf("Could not close stream: %v", err)
				}
				sendStream, err = createStream(ctx, client, args.ProjectID, args.RecognizerName)
				if err != nil {
					log.Fatalf("Could not create stream: %v", err)
				}
				reconnect = false
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

	// sigint されたときの中間結果をファイルに書き出すためのバッファ
	var interrimResult string

	// sigint されたときに interrimResult をファイルに書き出してから終了する
	go func() {
		trap := make(chan os.Signal, 1)
		signal.Notify(trap, os.Interrupt)
		sig := <-trap
		fmt.Printf("Received signal: %v\n", sig)
		if interrimResult != "" {
			writeFinalResult(interrimResult, args.OutputFilePath)
		}
		os.Exit(0)
	}()

	// gRPC Stream から結果を受信して標準出力に表示する
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

		// clear screen
		fmt.Print("\033[H\033[2J")
		t := ""
		for _, result := range resp.Results {
			s := result.Alternatives[0].Transcript
			t += s
			if result.IsFinal {
				writeFinalResult(s, args.OutputFilePath)
				reconnect = true
				interrimResult = ""
				break
			}
		}

		interrimResult = t
		fmt.Println(t)
	}
}

func writeFinalResult(s string, path string) {
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
