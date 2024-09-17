package recognize

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

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

	// is_final が来たタイミングで stream をリセット（切断・再接続）するが、
	// そのときに送信側で stream を利用しているとエラーが発生するので、ロックをかける。
	var mu sync.Mutex

	go func() {
		// Pipe stdin to the API.
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if n > 0 {
				mu.Lock()
				err := stream.Send(&speechpb.StreamingRecognizeRequest{
					StreamingRequest: &speechpb.StreamingRecognizeRequest_Audio{
						Audio: buf[:n],
					},
				})
				mu.Unlock()
				if err != nil {
					log.Printf("Could not send audio: %v", err)
				}
			}
			if err == io.EOF {
				// Nothing else to pipe, close the stream.
				if err := stream.CloseSend(); err != nil {
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

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
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
				mu.Lock()
				if err := stream.CloseSend(); err != nil {
					log.Fatalf("Could not close stream: %v", err)
				}
				stream, err = createStream(ctx, client, args.ProjectID, args.RecognizerName)
				if err != nil {
					log.Fatal(err)
				}
				mu.Unlock()
				break
			}
		}

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
