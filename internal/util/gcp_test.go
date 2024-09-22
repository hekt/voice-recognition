package util

import "testing"

func TestRecognizerParent(t *testing.T) {
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
			if got := RecognizerParent(tt.args.projectID); got != tt.want {
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
