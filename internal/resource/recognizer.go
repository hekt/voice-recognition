package resource

import (
	"fmt"

	"cloud.google.com/go/speech/apiv2/speechpb"
)

type Recognizer struct {
	Value string

	Name          string
	Model         string
	LanguageCodes []string
	PhraseSets    []*PhraseSet
}

func RestoreRecognizerFromProto(pb *speechpb.Recognizer) *Recognizer {
	r := &Recognizer{
		Name:  pb.Name,
		Value: fmt.Sprintf("%v", pb),
	}

	config := pb.GetDefaultRecognitionConfig()
	if config == nil {
		return r
	}

	r.Model = config.Model
	r.LanguageCodes = config.LanguageCodes

	if config.Adaptation == nil {
		return r
	}

	phraseSets := make([]*PhraseSet, 0, len(config.Adaptation.PhraseSets))
	for _, p := range config.Adaptation.PhraseSets {
		switch v := p.Value.(type) {
		case *speechpb.SpeechAdaptation_AdaptationPhraseSet_PhraseSet:
			phraseSets = append(phraseSets, &PhraseSet{
				Name: v.PhraseSet,
			})
		case *speechpb.SpeechAdaptation_AdaptationPhraseSet_InlinePhraseSet:
			phraseSets = append(phraseSets, RestorePhraseSetFromProto(v.InlinePhraseSet))
		}
	}
	r.PhraseSets = phraseSets

	return r
}
