package resource

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	myspeech "github.com/hekt/voice-recognition/internal/interfaces/speech"
	myspeechpb "github.com/hekt/voice-recognition/internal/interfaces/speechpb"
	"github.com/hekt/voice-recognition/internal/testutil"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func TestNewRecognizerManager(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &myspeech.ClientMock{}
		want := &recognizerManager{
			client: client,
		}
		got := NewRecognizerManager(client)
		if diff := cmp.Diff(
			got,
			want,
			cmp.AllowUnexported(recognizerManager{}),
			cmpopts.IgnoreUnexported(myspeech.ClientMock{}),
		); diff != "" {
			t.Errorf("NewRecognizerManager() mismatch (-want +got):\n%s", diff)
		}
	})
}

func Test_recognizerManager_Create(t *testing.T) {
	type args struct {
		args CreateRecognizerArgs
	}
	tests := []struct {
		name    string
		server  speechpb.SpeechServer
		args    args
		wantErr bool
	}{
		{
			name: "success",
			server: &myspeechpb.SpeechServerMock{
				CreateRecognizerFunc: func(
					_ context.Context,
					_ *speechpb.CreateRecognizerRequest,
				) (*longrunningpb.Operation, error) {
					return &longrunningpb.Operation{
						Done: true,
						Result: &longrunningpb.Operation_Response{
							Response: testutil.AnyResponse(t, &speechpb.Recognizer{}),
						},
					}, nil
				},
			},
			args: args{
				args: CreateRecognizerArgs{
					ProjectID:      "project-id",
					RecognizerName: "recognizer-name",
					Model:          "model",
					LanguageCode:   "language-code",
					PhraseSet:      "phrase-set",
				},
			},
			wantErr: false,
		},
		{
			name: "error on calling rpc",
			server: &myspeechpb.SpeechServerMock{
				CreateRecognizerFunc: func(
					_ context.Context,
					_ *speechpb.CreateRecognizerRequest,
				) (*longrunningpb.Operation, error) {
					return nil, errors.New("rpc error")
				},
			},
			args: args{
				args: CreateRecognizerArgs{
					ProjectID:      "project-id",
					RecognizerName: "recognizer-name",
					Model:          "model",
					LanguageCode:   "language-code",
					PhraseSet:      "phrase-set",
				},
			},
			wantErr: true,
		},
		{
			name: "error on waiting for operation",
			server: &myspeechpb.SpeechServerMock{
				CreateRecognizerFunc: func(
					_ context.Context,
					_ *speechpb.CreateRecognizerRequest,
				) (*longrunningpb.Operation, error) {
					return &longrunningpb.Operation{
						Done: true,
						Result: &longrunningpb.Operation_Error{
							Error: &status.Status{Code: int32(code.Code_UNKNOWN)},
						},
					}, nil
				},
			},
			args: args{
				args: CreateRecognizerArgs{
					ProjectID:      "project-id",
					RecognizerName: "recognizer-name",
					Model:          "model",
					LanguageCode:   "language-code",
					PhraseSet:      "phrase-set",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			m := &recognizerManager{
				client: testutil.MockSpeechClient(t, ctx, tt.server),
			}
			if err := m.Create(ctx, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("recognizerManager.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_recognizerManager_Delete(t *testing.T) {
	type args struct {
		args DeleteRecognizerArgs
	}
	tests := []struct {
		name    string
		server  speechpb.SpeechServer
		args    args
		wantErr bool
	}{
		{
			name: "success",
			server: &myspeechpb.SpeechServerMock{
				DeleteRecognizerFunc: func(
					_ context.Context,
					_ *speechpb.DeleteRecognizerRequest,
				) (*longrunningpb.Operation, error) {
					return &longrunningpb.Operation{
						Done: true,
						Result: &longrunningpb.Operation_Response{
							Response: testutil.AnyResponse(t, &speechpb.Recognizer{}),
						},
					}, nil
				},
			},
			args: args{
				args: DeleteRecognizerArgs{
					ProjectID:      "project-id",
					RecognizerName: "recognizer-name",
				},
			},
			wantErr: false,
		},
		{
			name: "error on calling rpc",
			server: &myspeechpb.SpeechServerMock{
				DeleteRecognizerFunc: func(
					_ context.Context,
					_ *speechpb.DeleteRecognizerRequest,
				) (*longrunningpb.Operation, error) {
					return nil, errors.New("rpc error")
				},
			},
			args: args{
				args: DeleteRecognizerArgs{
					ProjectID:      "project-id",
					RecognizerName: "recognizer-name",
				},
			},
			wantErr: true,
		},
		{
			name: "error on waiting for operation",
			server: &myspeechpb.SpeechServerMock{
				DeleteRecognizerFunc: func(
					_ context.Context,
					_ *speechpb.DeleteRecognizerRequest,
				) (*longrunningpb.Operation, error) {
					return &longrunningpb.Operation{
						Done: true,
						Result: &longrunningpb.Operation_Error{
							Error: &status.Status{Code: int32(code.Code_UNKNOWN)},
						},
					}, nil
				},
			},
			args: args{
				args: DeleteRecognizerArgs{
					ProjectID:      "project-id",
					RecognizerName: "recognizer-name",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			m := &recognizerManager{
				client: testutil.MockSpeechClient(t, ctx, tt.server),
			}
			if err := m.Delete(ctx, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("recognizerManager.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_recognizerManager_List(t *testing.T) {
	type args struct {
		args ListRecognizerArgs
	}
	tests := []struct {
		name      string
		server    speechpb.SpeechServer
		args      args
		wantCount int
		wantErr   bool
	}{
		{
			name: "success",
			server: &myspeechpb.SpeechServerMock{
				ListRecognizersFunc: func(
					_ context.Context,
					_ *speechpb.ListRecognizersRequest,
				) (*speechpb.ListRecognizersResponse, error) {
					return &speechpb.ListRecognizersResponse{
						Recognizers: []*speechpb.Recognizer{
							{Name: "recognizer-1"},
							{Name: "recognizer-2"},
							{Name: "recognizer-3"},
						},
					}, nil
				},
			},
			args: args{
				args: ListRecognizerArgs{
					ProjectID: "project-id",
				},
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "error on calling rpc",
			server: &myspeechpb.SpeechServerMock{
				ListRecognizersFunc: func(
					_ context.Context,
					_ *speechpb.ListRecognizersRequest,
				) (*speechpb.ListRecognizersResponse, error) {
					return nil, errors.New("rpc error")
				},
			},
			args: args{
				args: ListRecognizerArgs{
					ProjectID: "project-id",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			m := &recognizerManager{
				client: testutil.MockSpeechClient(t, ctx, tt.server),
			}
			got, err := m.List(ctx, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("recognizerManager.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("recognizerManager.List() count = %v, want %v", len(got), tt.wantCount)
			}
		})
	}
}
