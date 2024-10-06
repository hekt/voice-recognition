package phraseset

import (
	"context"
	"fmt"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/hekt/voice-recognition/internal/resource"
)

type CreateArgs struct {
	ProjectID     string
	PhraseSetName string
	Phrases       []string
	Boost         float32
}

func Create(ctx context.Context, args CreateArgs) error {
	client, err := speech.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	phrases := make([]*speechpb.PhraseSet_Phrase, 0, len(args.Phrases))
	for _, phrase := range args.Phrases {
		phrases = append(phrases, &speechpb.PhraseSet_Phrase{
			Value: phrase,
			Boost: 0,
		})
	}

	op, err := client.CreatePhraseSet(ctx, &speechpb.CreatePhraseSetRequest{
		PhraseSet: &speechpb.PhraseSet{
			DisplayName: args.PhraseSetName,
			Phrases:     phrases,
			Boost:       args.Boost,
		},
		PhraseSetId: args.PhraseSetName,
		Parent:      resource.ParentName(args.ProjectID),
	})

	if err != nil {
		return fmt.Errorf("failed to create phrase set: %w", err)
	}

	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed to wait for create operation: %w", err)
	}

	fmt.Println("phrase set created")

	return nil
}
