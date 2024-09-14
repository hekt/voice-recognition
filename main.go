package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

func actualLines(s string) int {
	c, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		panic(err)
	}
	w := runewidth.StringWidth(s)
	return w/c + min(1, w%c)
}

func clearLines(n int) {
	for i := 0; i < n; i++ {
		fmt.Print("\033[F\033[K")
	}
}

func printRim(s string) {
	fmt.Print(s)
}

func printFinal(s string) {
	fmt.Print("\033[K")
	fmt.Print("\033[32m")
	fmt.Print(s)
	fmt.Print("\033[0m")
	fmt.Print("\n")
}

func main() {
	ctx := context.Background()

	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	stream, err := client.StreamingRecognize(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// Send the initial configuration message.
	if err := stream.Send(&speechpb.StreamingRecognizeRequest{
		StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
			StreamingConfig: &speechpb.StreamingRecognitionConfig{
				Config: &speechpb.RecognitionConfig{
					Encoding:        speechpb.RecognitionConfig_LINEAR16,
					SampleRateHertz: 16000,
					LanguageCode:    "ja-JP",
					Model:           "latest_long",
				},
				InterimResults: true,
			},
		},
	}); err != nil {
		log.Fatal(err)
	}

	go func() {
		// Pipe stdin to the API.
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if n > 0 {
				if err := stream.Send(&speechpb.StreamingRecognizeRequest{
					StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
						AudioContent: buf[:n],
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

	previousLines := ""
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Cannot stream results: %v", err)
		}
		if err := resp.Error; err != nil {
			// Workaround while the API doesn't give a more informative error.
			if err.Code == 3 || err.Code == 11 {
				log.Print("WARNING: Speech recognition request exceeded limit of 60 seconds.")
			}
			log.Fatalf("Could not recognize: %v", err)
		}

		if len(resp.Results) == 0 {
			continue
		}

		clearLines(actualLines(previousLines) - 1)

		previousLines = ""
		fmt.Print("\r")
		for _, result := range resp.Results {
			s := result.Alternatives[0].Transcript
			if result.IsFinal {
				printFinal(s)
			} else {
				if len(s) > 0 {
					previousLines += s
				}
				printRim(s)
			}
		}
	}
}
