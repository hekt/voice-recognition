package recognizer

import (
	"context"
	"fmt"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/hekt/voice-recognition/internal/resource"
)

type CreateArgs struct {
	ProjectID      string
	RecognizerName string
	Model          string
	LanguageCode   string
	PhraseSet      string
}

func Create(ctx context.Context, args CreateArgs) error {
	client, err := speech.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	phraseSets := []*speechpb.SpeechAdaptation_AdaptationPhraseSet{}
	if args.PhraseSet != "" {
		phraseSets = append(phraseSets, &speechpb.SpeechAdaptation_AdaptationPhraseSet{
			Value: &speechpb.SpeechAdaptation_AdaptationPhraseSet_PhraseSet{
				PhraseSet: resource.PhraseSetFullname(args.ProjectID, args.PhraseSet),
			},
		})
	}

	op, err := client.CreateRecognizer(ctx, &speechpb.CreateRecognizerRequest{
		Parent:       resource.ParentName(args.ProjectID),
		RecognizerId: args.RecognizerName,
		Recognizer: &speechpb.Recognizer{
			DisplayName: args.RecognizerName,
			DefaultRecognitionConfig: &speechpb.RecognitionConfig{
				Model:         args.Model,
				LanguageCodes: []string{args.LanguageCode},
				DecodingConfig: &speechpb.RecognitionConfig_ExplicitDecodingConfig{
					ExplicitDecodingConfig: &speechpb.ExplicitDecodingConfig{
						Encoding:          speechpb.ExplicitDecodingConfig_LINEAR16,
						SampleRateHertz:   16000,
						AudioChannelCount: 1,
					},
				},
				Features: &speechpb.RecognitionFeatures{
					EnableAutomaticPunctuation: true,
				},
				Adaptation: &speechpb.SpeechAdaptation{
					PhraseSets: phraseSets,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create recognizer: %w", err)
	}

	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed to wait for create operation: %w", err)
	}

	fmt.Println("Recognizer created")

	return nil
}
