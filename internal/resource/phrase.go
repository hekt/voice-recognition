package resource

import (
	"cloud.google.com/go/speech/apiv2/speechpb"
)

type Phrase struct {
	Value string
	Boost float32
}

func NewPhraseFromProto(pb *speechpb.PhraseSet_Phrase) *Phrase {
	return &Phrase{
		Value: pb.Value,
		Boost: pb.Boost,
	}
}
