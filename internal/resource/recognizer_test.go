package resource

import (
	"testing"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestRestoreRecognizerFromProto(t *testing.T) {
	type args struct {
		pb *speechpb.Recognizer
	}
	tests := []struct {
		name string
		args args
		want *Recognizer
	}{
		{
			name: "success",
			args: args{
				pb: &speechpb.Recognizer{
					Name: "test",
					DefaultRecognitionConfig: &speechpb.RecognitionConfig{
						Model:         "test-model",
						LanguageCodes: []string{"ja-JP", "en-US"},
						Adaptation: &speechpb.SpeechAdaptation{
							PhraseSets: []*speechpb.SpeechAdaptation_AdaptationPhraseSet{
								{
									Value: &speechpb.SpeechAdaptation_AdaptationPhraseSet_PhraseSet{
										PhraseSet: "test-phrase-set-1",
									},
								},
								{
									Value: &speechpb.SpeechAdaptation_AdaptationPhraseSet_InlinePhraseSet{
										InlinePhraseSet: &speechpb.PhraseSet{
											Name: "test-phrase-set-2",
											Phrases: []*speechpb.PhraseSet_Phrase{
												{
													Value: "test-phrase-2-1",
													Boost: 1.2,
												},
												{
													Value: "test-phrase-2-2",
													Boost: 1.3,
												},
											},
											Boost: 1.4,
										},
									},
								},
							},
						},
					},
				},
			},
			want: &Recognizer{
				Name:          "test",
				Model:         "test-model",
				LanguageCodes: []string{"ja-JP", "en-US"},
				PhraseSets: []*PhraseSet{
					{
						Name: "test-phrase-set-1",
					},
					{
						Name: "test-phrase-set-2",
						Phrases: []*Phrase{
							{
								Value: "test-phrase-2-1",
								Boost: 1.2,
							},
							{
								Value: "test-phrase-2-2",
								Boost: 1.3,
							},
						},
						Boost: 1.4,
					},
				},
			},
		},
		{
			name: "no default config",
			args: args{
				pb: &speechpb.Recognizer{
					Name: "test",
				},
			},
			want: &Recognizer{
				Name: "test",
			},
		},
		{
			name: "no adaptation",
			args: args{
				pb: &speechpb.Recognizer{
					Name: "test",
					DefaultRecognitionConfig: &speechpb.RecognitionConfig{
						Model:         "test-model",
						LanguageCodes: []string{"ja-JP", "en-US"},
					},
				},
			},
			want: &Recognizer{
				Name:          "test",
				Model:         "test-model",
				LanguageCodes: []string{"ja-JP", "en-US"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RestoreRecognizerFromProto(tt.args.pb)
			if diff := cmp.Diff(
				got,
				tt.want,
				cmpopts.IgnoreFields(Recognizer{}, "Value"),
				cmpopts.IgnoreFields(PhraseSet{}, "Value"),
			); diff != "" {
				t.Errorf("RestoreRecognizerFromProto() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
