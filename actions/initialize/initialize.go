package initialize

import (
	"context"
	"fmt"
	"log"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
)

type Args struct {
	ProjectID      string
	RecognizerName string
	Model          string
}

func Run(ctx context.Context, args Args) {
	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	parent := fmt.Sprintf("projects/%s/locations/global", args.ProjectID)

	_, err = client.CreateRecognizer(ctx, &speechpb.CreateRecognizerRequest{
		Parent:       parent,
		RecognizerId: args.RecognizerName,
		Recognizer: &speechpb.Recognizer{
			DisplayName: "default-recognizer",
			DefaultRecognitionConfig: &speechpb.RecognitionConfig{
				Model:         args.Model,
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
