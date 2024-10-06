package resource

import (
	"fmt"

	"cloud.google.com/go/speech/apiv2/speechpb"
)

type PhraseSet struct {
	Value string

	Name    string
	Phrases []*Phrase
	Boost   float32
}

func RestorePhraseSetFromProto(pb *speechpb.PhraseSet) *PhraseSet {
	phrases := make([]*Phrase, 0, len(pb.Phrases))
	for _, p := range pb.Phrases {
		phrases = append(phrases, NewPhraseFromProto(p))
	}

	return &PhraseSet{
		Name:    pb.Name,
		Phrases: phrases,
		Boost:   pb.Boost,
		Value:   fmt.Sprintf("%v", pb),
	}
}
