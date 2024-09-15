package actions

import (
	"context"
	"fmt"
	"log"
	"os"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
)

func Initialize() {
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
