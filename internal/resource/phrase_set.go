package resource

import (
	"context"
	"fmt"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/hekt/voice-recognition/internal/interfaces/speech"
	"google.golang.org/api/iterator"
)

type PhraseSetManager interface {
	Create(ctx context.Context, args CreatePhraseSetArgs) error
	Update(ctx context.Context, args UpdatePhraseSetArgs) error
	List(ctx context.Context, args ListPhraseSetArgs) ([]*PhraseSet, error)
}

type CreatePhraseSetArgs struct {
	ProjectID     string
	PhraseSetName string
	Phrases       []string
	Boost         float32
}

type UpdatePhraseSetArgs struct {
	ProjectID     string
	PhraseSetName string
	Phrases       []string
	Boost         float32
}

type ListPhraseSetArgs struct {
	ProjectID string
}

type PhraseSet struct {
	Value string
}

type phraseSetManager struct {
	client speech.Client
}

var _ PhraseSetManager = (*phraseSetManager)(nil)

func NewPhraseSetManager(client speech.Client) *phraseSetManager {
	return &phraseSetManager{
		client: client,
	}
}

func (m *phraseSetManager) Create(ctx context.Context, args CreatePhraseSetArgs) error {

	phrases := make([]*speechpb.PhraseSet_Phrase, 0, len(args.Phrases))
	for _, phrase := range args.Phrases {
		phrases = append(phrases, &speechpb.PhraseSet_Phrase{
			Value: phrase,
			Boost: 0,
		})
	}

	op, err := m.client.CreatePhraseSet(ctx, &speechpb.CreatePhraseSetRequest{
		PhraseSet: &speechpb.PhraseSet{
			DisplayName: args.PhraseSetName,
			Phrases:     phrases,
			Boost:       args.Boost,
		},
		PhraseSetId: args.PhraseSetName,
		Parent:      ParentName(args.ProjectID),
	})

	if err != nil {
		return fmt.Errorf("failed to create phrase set: %w", err)
	}

	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed to wait for create operation: %w", err)
	}

	return nil
}

func (m *phraseSetManager) Update(ctx context.Context, args UpdatePhraseSetArgs) error {
	phrases := make([]*speechpb.PhraseSet_Phrase, 0, len(args.Phrases))
	for _, phrase := range args.Phrases {
		phrases = append(phrases, &speechpb.PhraseSet_Phrase{
			Value: phrase,
			Boost: 0,
		})
	}

	op, err := m.client.UpdatePhraseSet(ctx, &speechpb.UpdatePhraseSetRequest{
		PhraseSet: &speechpb.PhraseSet{
			Name:    PhraseSetFullname(args.ProjectID, args.PhraseSetName),
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

	return nil
}

func (m *phraseSetManager) List(ctx context.Context, args ListPhraseSetArgs) ([]*PhraseSet, error) {
	iterResp := m.client.ListPhraseSets(ctx, &speechpb.ListPhraseSetsRequest{
		Parent:      ParentName(args.ProjectID),
		ShowDeleted: true,
	})

	phraseSets := make([]*PhraseSet, 0)
	for {
		resp, err := iterResp.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, fmt.Errorf("failed to get next response: %w", err)
		}
		phraseSets = append(phraseSets, &PhraseSet{
			Value: fmt.Sprintf("%v", resp),
		})
	}

	return phraseSets, nil
}
