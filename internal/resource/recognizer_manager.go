package resource

import (
	"context"
	"fmt"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/hekt/voice-recognition/internal/interfaces/speech"
	"google.golang.org/api/iterator"
)

type RecognizerManager interface {
	Create(ctx context.Context, args CreateRecognizerArgs) error
	Delete(ctx context.Context, args DeleteRecognizerArgs) error
	List(ctx context.Context, args ListRecognizerArgs) ([]*Recognizer, error)
}

type CreateRecognizerArgs struct {
	ProjectID      string
	RecognizerName string
	Model          string
	LanguageCode   string
	PhraseSet      string
}

type DeleteRecognizerArgs struct {
	ProjectID      string
	RecognizerName string
}

type ListRecognizerArgs struct {
	ProjectID string
}

type recognizerManager struct {
	client speech.Client
}

var _ RecognizerManager = (*recognizerManager)(nil)

func NewRecognizerManager(client speech.Client) *recognizerManager {
	return &recognizerManager{
		client: client,
	}
}

func (m *recognizerManager) Create(ctx context.Context, args CreateRecognizerArgs) error {
	phraseSets := []*speechpb.SpeechAdaptation_AdaptationPhraseSet{}
	if args.PhraseSet != "" {
		phraseSets = append(phraseSets, &speechpb.SpeechAdaptation_AdaptationPhraseSet{
			Value: &speechpb.SpeechAdaptation_AdaptationPhraseSet_PhraseSet{
				PhraseSet: PhraseSetFullname(args.ProjectID, args.PhraseSet),
			},
		})
	}

	op, err := m.client.CreateRecognizer(ctx, &speechpb.CreateRecognizerRequest{
		Parent:       ParentName(args.ProjectID),
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

	return nil
}

func (m *recognizerManager) Delete(ctx context.Context, args DeleteRecognizerArgs) error {
	op, err := m.client.DeleteRecognizer(ctx, &speechpb.DeleteRecognizerRequest{
		Name: RecognizerFullname(args.ProjectID, args.RecognizerName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete recognizer: %w", err)
	}

	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed to wait for delete operation: %w", err)
	}

	return nil
}

func (m *recognizerManager) List(ctx context.Context, args ListRecognizerArgs) ([]*Recognizer, error) {
	iterResp := m.client.ListRecognizers(ctx, &speechpb.ListRecognizersRequest{
		Parent:      ParentName(args.ProjectID),
		ShowDeleted: true,
	})

	recognizers := make([]*Recognizer, 0)
	for {
		resp, err := iterResp.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, fmt.Errorf("failed to get next response: %w", err)
		}
		recognizers = append(recognizers, RestoreRecognizerFromProto(resp))
	}

	return recognizers, nil
}
