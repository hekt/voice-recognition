package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
)

const (
	outputFilePath = "output.txt"
	recognizerName = "myrecognizer"
)

func createRecognizer() {
	ctx := context.Background()
	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	parent := fmt.Sprintf("projects/%s/locations/global", project)

	_, err = client.CreateRecognizer(ctx, &speechpb.CreateRecognizerRequest{
		Parent:       parent,
		RecognizerId: recognizerName,
		Recognizer: &speechpb.Recognizer{
			DisplayName: "default-recognizer",
			DefaultRecognitionConfig: &speechpb.RecognitionConfig{
				Model:         "long",
				LanguageCodes: []string{"ja-jp"},
				DecodingConfig: &speechpb.RecognitionConfig_ExplicitDecodingConfig{
					ExplicitDecodingConfig: &speechpb.ExplicitDecodingConfig{
						Encoding:          speechpb.ExplicitDecodingConfig_LINEAR16,
						SampleRateHertz:   16000,
						AudioChannelCount: 1,
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
		fmt.Printf("Failed to create recognizer: %v", err)
	}
}

func writeFinalResult(s string) {
	file, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(s + "\n"); err != nil {
		log.Fatal(err)
	}
}

func createStream(ctx context.Context, client *speech.Client) (speechpb.Speech_StreamingRecognizeClient, error) {
	stream, err := client.StreamingRecognize(ctx)
	if err != nil {
		return nil, err
	}

	if err := stream.Send(&speechpb.StreamingRecognizeRequest{
		Recognizer: fmt.Sprintf("projects/%s/locations/global/recognizers/%s", os.Getenv("GOOGLE_CLOUD_PROJECT"), recognizerName),
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

func main() {
	ctx := context.Background()

	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	stream, err := createStream(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	// todo おそらく stream が閉じた直後に呼ばれるとエラーが発生して死ぬので、エラーハンドリングが必要
	go func() {
		// Pipe stdin to the API.
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if n > 0 {
				if err := stream.Send(&speechpb.StreamingRecognizeRequest{
					StreamingRequest: &speechpb.StreamingRecognizeRequest_Audio{
						Audio: buf[:n],
					},
				}); err != nil {
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
		// if err := resp.Error; err != nil {
		// 	// Workaround while the API doesn't give a more informative error.
		// 	if err.Code == 3 || err.Code == 11 {
		// 		log.Print("WARNING: Speech recognition request exceeded limit of 60 seconds.")
		// 	}
		// 	log.Fatalf("Could not recognize: %v", err)
		// }

		// clear screen
		fmt.Print("\033[H\033[2J")
		t := ""
		for _, result := range resp.Results {
			s := result.Alternatives[0].Transcript
			t += s
			if result.IsFinal {
				writeFinalResult(s)
				if err := stream.CloseSend(); err != nil {
					log.Fatalf("Could not close stream: %v", err)
				}
				stream, err = createStream(ctx, client)
				if err != nil {
					log.Fatal(err)
				}
				break
			}
		}

		fmt.Println(t)
	}
}
