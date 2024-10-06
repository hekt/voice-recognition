package resource

import (
	"testing"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestRestorePhraseSetFromProto(t *testing.T) {
	type args struct {
		pb *speechpb.PhraseSet
	}
	tests := []struct {
		name string
		args args
		want *PhraseSet
	}{
		{
			name: "success",
			args: args{
				pb: &speechpb.PhraseSet{
					Name: "test",
					Phrases: []*speechpb.PhraseSet_Phrase{
						{
							Value: "test-phrase-1",
							Boost: 1.2,
						},
						{
							Value: "test-phrase-2",
							Boost: 1.3,
						},
					},
					Boost: 1.4,
				},
			},
			want: &PhraseSet{
				Name: "test",
				Phrases: []*Phrase{
					{
						Value: "test-phrase-1",
						Boost: 1.2,
					},
					{
						Value: "test-phrase-2",
						Boost: 1.3,
					},
				},
				Boost: 1.4,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RestorePhraseSetFromProto(tt.args.pb)
			if diff := cmp.Diff(got, tt.want, cmpopts.IgnoreFields(PhraseSet{}, "Value")); diff != "" {
				t.Errorf("RestorePhraseSetFromProto() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
