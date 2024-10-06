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

func TestNewPhraseSetManager(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &myspeech.ClientMock{}
		want := &phraseSetManager{
			client: client,
		}
		got := NewPhraseSetManager(client)
		if diff := cmp.Diff(
			got,
			want,
			cmp.AllowUnexported(phraseSetManager{}),
			cmpopts.IgnoreUnexported(myspeech.ClientMock{}),
		); diff != "" {
			t.Errorf("NewPhraseSetManager() mismatch (-want +got):\n%s", diff)
		}
	})
}

func Test_phraseSetManager_Create(t *testing.T) {
	type args struct {
		args CreatePhraseSetArgs
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
				CreatePhraseSetFunc: func(
					_ context.Context,
					_ *speechpb.CreatePhraseSetRequest,
				) (*longrunningpb.Operation, error) {
					return &longrunningpb.Operation{
						Done: true,
						Result: &longrunningpb.Operation_Response{
							Response: testutil.AnyResponse(t, &speechpb.PhraseSet{}),
						},
					}, nil
				},
			},
			args: args{
				args: CreatePhraseSetArgs{
					ProjectID:     "test-project-id",
					PhraseSetName: "test-phrase-set-name",
					Phrases:       []string{"test-phrase"},
					Boost:         0,
				},
			},
			wantErr: false,
		},
		{
			name: "error on calling rpc",
			server: &myspeechpb.SpeechServerMock{
				CreatePhraseSetFunc: func(
					_ context.Context,
					_ *speechpb.CreatePhraseSetRequest,
				) (*longrunningpb.Operation, error) {
					return nil, errors.New("rpc error")
				},
			},
			args: args{
				args: CreatePhraseSetArgs{
					ProjectID:     "test-project-id",
					PhraseSetName: "test-phrase-set-name",
					Phrases:       []string{"test-phrase"},
					Boost:         0,
				},
			},
			wantErr: true,
		},
		{
			name: "error on waiting for operation",
			server: &myspeechpb.SpeechServerMock{
				CreatePhraseSetFunc: func(
					_ context.Context,
					_ *speechpb.CreatePhraseSetRequest,
				) (*longrunningpb.Operation, error) {
					return &longrunningpb.Operation{
						Done: true,
						Result: &longrunningpb.Operation_Error{
							Error: &status.Status{
								Code: int32(code.Code_UNKNOWN),
							},
						},
					}, nil
				},
			},
			args: args{
				args: CreatePhraseSetArgs{
					ProjectID:     "test-project-id",
					PhraseSetName: "test-phrase-set-name",
					Phrases:       []string{"test-phrase"},
					Boost:         0,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			m := &phraseSetManager{
				client: testutil.MockSpeechClient(t, ctx, tt.server),
			}
			if err := m.Create(ctx, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("phraseSetManager.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_phraseSetManager_Update(t *testing.T) {
	type args struct {
		args UpdatePhraseSetArgs
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
				UpdatePhraseSetFunc: func(
					_ context.Context,
					_ *speechpb.UpdatePhraseSetRequest,
				) (*longrunningpb.Operation, error) {
					return &longrunningpb.Operation{
						Done: true,
						Result: &longrunningpb.Operation_Response{
							Response: testutil.AnyResponse(t, &speechpb.PhraseSet{}),
						},
					}, nil
				},
			},
			args: args{
				args: UpdatePhraseSetArgs{
					ProjectID:     "test-project-id",
					PhraseSetName: "test-phrase-set-name",
					Phrases:       []string{"test-phrase"},
					Boost:         0,
				},
			},
			wantErr: false,
		},
		{
			name: "error on calling rpc",
			server: &myspeechpb.SpeechServerMock{
				UpdatePhraseSetFunc: func(
					_ context.Context,
					_ *speechpb.UpdatePhraseSetRequest,
				) (*longrunningpb.Operation, error) {
					return nil, errors.New("rpc error")
				},
			},
			args: args{
				args: UpdatePhraseSetArgs{
					ProjectID:     "test-project-id",
					PhraseSetName: "test-phrase-set-name",
					Phrases:       []string{"test-phrase"},
					Boost:         0,
				},
			},
			wantErr: true,
		},
		{
			name: "error on waiting for operation",
			server: &myspeechpb.SpeechServerMock{
				UpdatePhraseSetFunc: func(
					_ context.Context,
					_ *speechpb.UpdatePhraseSetRequest,
				) (*longrunningpb.Operation, error) {
					return &longrunningpb.Operation{
						Done: true,
						Result: &longrunningpb.Operation_Error{
							Error: &status.Status{
								Code: int32(code.Code_UNKNOWN),
							},
						},
					}, nil
				},
			},
			args: args{
				args: UpdatePhraseSetArgs{
					ProjectID:     "test-project-id",
					PhraseSetName: "test-phrase-set-name",
					Phrases:       []string{"test-phrase"},
					Boost:         0,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			m := &phraseSetManager{
				client: testutil.MockSpeechClient(t, ctx, tt.server),
			}
			if err := m.Update(ctx, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("phraseSetManager.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_phraseSetManager_List(t *testing.T) {
	type args struct {
		args ListPhraseSetArgs
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
				ListPhraseSetsFunc: func(
					_ context.Context,
					_ *speechpb.ListPhraseSetsRequest,
				) (*speechpb.ListPhraseSetsResponse, error) {
					return &speechpb.ListPhraseSetsResponse{
						PhraseSets: []*speechpb.PhraseSet{
							{Name: "phrase-set-1"},
							{Name: "phrase-set-2"},
							{Name: "phrase-set-3"},
						},
					}, nil
				},
			},
			args: args{
				args: ListPhraseSetArgs{
					ProjectID: "test-project-id",
				},
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "error on calling rpc",
			server: &myspeechpb.SpeechServerMock{
				ListPhraseSetsFunc: func(
					_ context.Context,
					_ *speechpb.ListPhraseSetsRequest,
				) (*speechpb.ListPhraseSetsResponse, error) {
					return nil, errors.New("rpc error")
				},
			},
			args: args{
				args: ListPhraseSetArgs{
					ProjectID: "test-project-id",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			m := &phraseSetManager{
				client: testutil.MockSpeechClient(t, ctx, tt.server),
			}
			got, err := m.List(ctx, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("phraseSetManager.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("phraseSetManager.List() got = %v, wantCount %v", len(got), tt.wantCount)
			}
		})
	}
}
