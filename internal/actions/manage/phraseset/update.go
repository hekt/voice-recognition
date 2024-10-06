package phraseset

import (
	"context"
	"fmt"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/hekt/voice-recognition/internal/resource"
)

type UpdateArgs struct {
	ProjectID     string
	PhraseSetName string
	Phrases       []string
	Boost         float32
}

func Update(ctx context.Context, args UpdateArgs) error {
	client, err := speech.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	phrases := make([]*speechpb.PhraseSet_Phrase, 0, len(args.Phrases))
	for _, phrase := range args.Phrases {
		phrases = append(phrases, &speechpb.PhraseSet_Phrase{
			Value: phrase,
			Boost: 0,
		})
	}

	op, err := client.UpdatePhraseSet(ctx, &speechpb.UpdatePhraseSetRequest{
		PhraseSet: &speechpb.PhraseSet{
			Name:    resource.PhraseSetFullname(args.ProjectID, args.PhraseSetName),
			Phrases: phrases,
			Boost:   args.Boost,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to update phrase set: %w", err)
	}

	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed to wait for update operation: %w", err)
	}

	fmt.Println("phrase set updated")

	return nil
}
