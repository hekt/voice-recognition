package resource

import (
	"testing"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/google/go-cmp/cmp"
)

func TestNewPhraseFromProto(t *testing.T) {
	type args struct {
		pb *speechpb.PhraseSet_Phrase
	}
	tests := []struct {
		name string
		args args
		want *Phrase
	}{
		{
			name: "success",
			args: args{
				pb: &speechpb.PhraseSet_Phrase{
					Value: "test",
					Boost: 1.2,
				},
			},
			want: &Phrase{
				Value: "test",
				Boost: 1.2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPhraseFromProto(tt.args.pb)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("NewPhraseFromProto() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
