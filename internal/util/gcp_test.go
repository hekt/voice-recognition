package util

import (
	"testing"
)

func TestResourceParent(t *testing.T) {
	type args struct {
		projectID string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				projectID: "test-project",
			},
			want: "projects/test-project/locations/global",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResourceParent(tt.args.projectID); got != tt.want {
				t.Errorf("RecognizerParent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecognizerFullname(t *testing.T) {
	type args struct {
		projectID      string
		recognizerName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				projectID:      "test-project",
				recognizerName: "test-recognizer",
			},
			want: "projects/test-project/locations/global/recognizers/test-recognizer",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RecognizerFullname(tt.args.projectID, tt.args.recognizerName); got != tt.want {
				t.Errorf("RecognizerFullname() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPhraseSetFullname(t *testing.T) {
	type args struct {
		projectID     string
		phraseSetName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				projectID:     "test-project",
				phraseSetName: "test-phrase-set",
			},
			want: "projects/test-project/locations/global/phraseSets/test-phrase-set",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PhraseSetFullname(tt.args.projectID, tt.args.phraseSetName); got != tt.want {
				t.Errorf("PhraseSetFullname() = %v, want %v", got, tt.want)
			}
		})
	}
}
