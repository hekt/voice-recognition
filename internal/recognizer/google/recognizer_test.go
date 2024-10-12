package google

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/google/go-cmp/cmp"
	myspeech "github.com/hekt/voice-recognition/internal/interfaces/speech"
	myspeechpb "github.com/hekt/voice-recognition/internal/interfaces/speechpb"
	"github.com/hekt/voice-recognition/internal/recognizer/model"
)

func TestNewRecognizer(t *testing.T) {
	type args struct {
		ctx               context.Context
		audioCh           <-chan []byte
		resultCh          chan<- []*model.Result
		projectID         string
		recognizerName    string
		reconnectInterval time.Duration
	}
	baseArgs := args{
		ctx:               context.Background(),
		audioCh:           make(chan []byte),
		resultCh:          make(chan []*model.Result),
		projectID:         "test-project-id",
		recognizerName:    "test-recognizer-name",
		reconnectInterval: time.Minute,
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "valid",
			args:    baseArgs,
			wantErr: false,
		},
		{
			name: "empty project ID",
			args: func() args {
				a := baseArgs
				a.projectID = ""
				return a
			}(),
			wantErr: true,
		},
		{
			name: "empty recognizer name",
			args: func() args {
				a := baseArgs
				a.recognizerName = ""
				return a
			}(),
			wantErr: true,
		},
		{
			name: "reconnect interval less than 1 minute",
			args: func() args {
				a := baseArgs
				a.reconnectInterval = time.Second
				return a
			}(),
			wantErr: true,
		},
		{
			name: "nil audio channel",
			args: func() args {
				a := baseArgs
				a.audioCh = nil
				return a
			}(),
			wantErr: true,
		},
		{
			name: "nil result channel",
			args: func() args {
				a := baseArgs
				a.resultCh = nil
				return a
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRecognizer(
				tt.args.ctx,
				tt.args.audioCh,
				tt.args.resultCh,
				tt.args.projectID,
				tt.args.recognizerName,
				tt.args.reconnectInterval,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRecognizer() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if got.streamSupplier == nil {
				t.Error("streamSupplier is nil")
			}
			if got.audioSender == nil {
				t.Error("audioSender is nil")
			}
			if got.responseReceiver == nil {
				t.Error("responseReceiver is nil")
			}
			if got.responseProcessor == nil {
				t.Error("responseProcessor is nil")
			}
			if got.client == nil {
				t.Error("client is nil")
			}
			if got.responseCh == nil {
				t.Error("responseCh is nil")
			}
			if got.sendStreamCh == nil {
				t.Error("sendStreamCh is nil")
			}
			if got.receiveStreamCh == nil {
				t.Error("receiveStreamCh is nil")
			}
		})
	}
}

func TestRecognizer_Start(t *testing.T) {
	t.Run("complex test", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		audioCh := make(chan []byte)
		resultCh := make(chan []*model.Result, 3)

		// mock channels
		mockStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 2)
		mockStreamDataCh := make(chan []byte, 10)

		streamClientMock := &myspeechpb.Speech_StreamingRecognizeClientMock{
			SendFunc: func(req *speechpb.StreamingRecognizeRequest) error {
				audioRequest, ok := req.StreamingRequest.(*speechpb.StreamingRecognizeRequest_Audio)
				if !ok {
					return errors.New("unexpected request type")
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case mockStreamDataCh <- audioRequest.Audio:
					return nil
				}
			},
			RecvFunc: func() (*speechpb.StreamingRecognizeResponse, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case data := <-mockStreamDataCh:
					return &speechpb.StreamingRecognizeResponse{
						Results: []*speechpb.StreamingRecognitionResult{
							{
								Alternatives: []*speechpb.SpeechRecognitionAlternative{
									{Transcript: string(data)},
								},
								IsFinal: false,
							},
						},
					}, nil
				}
			},
			CloseSendFunc: func() error {
				return nil
			},
		}

		// internal channels
		sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)
		receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)
		responseCh := make(chan *speechpb.StreamingRecognizeResponse)

		// workers
		streamSupplier := &StreamSupplierInterfaceMock{
			SupplyFunc: func(ctx context.Context) error {
				stream := <-mockStreamCh
				sendStreamCh <- stream
				receiveStreamCh <- stream
				return nil
			},
			StartFunc: func(ctx context.Context) error {
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case stream := <-mockStreamCh:
						select {
						case <-ctx.Done():
							return ctx.Err()
						case sendStreamCh <- stream:
						}

						select {
						case <-ctx.Done():
							return ctx.Err()
						case receiveStreamCh <- stream:
						}
					}
				}
			},
		}
		audioSender := NewAudioSender(audioCh, sendStreamCh)
		responseReceiver := NewResponseReceiver(responseCh, receiveStreamCh)
		responseProcessor := NewResponseProcessor(responseCh, resultCh)

		r := &Recognizer{
			streamSupplier:    streamSupplier,
			audioSender:       audioSender,
			responseReceiver:  responseReceiver,
			responseProcessor: responseProcessor,
			client: &myspeech.ClientMock{
				CloseFunc: func() error {
					return nil
				},
			},
			responseCh:      responseCh,
			sendStreamCh:    sendStreamCh,
			receiveStreamCh: receiveStreamCh,
			audioCh:         audioCh,
			resultCh:        resultCh,
		}

		var wg sync.WaitGroup
		wg.Add(1)
		var got error
		go func() {
			defer wg.Done()
			got = r.Start(ctx)
		}()

		mockStreamCh <- streamClientMock
		audioCh <- []byte("test1")
		audioCh <- []byte("test2")
		mockStreamCh <- streamClientMock
		audioCh <- []byte("test3")

		for {
			time.Sleep(10 * time.Millisecond)
			if len(resultCh) == 3 {
				break
			}
		}

		cancel()
		wg.Wait()

		close(audioCh)
		close(resultCh)

		if !errors.Is(got, context.Canceled) {
			t.Errorf("Recognizer.Start() error = %v, wantErr %v", got, context.Canceled)
		}

		wantResults := [][]*model.Result{
			{{Transcript: "test1", IsFinal: false}},
			{{Transcript: "test2", IsFinal: false}},
			{{Transcript: "test3", IsFinal: false}},
		}
		if g, w := len(resultCh), len(wantResults); g != w {
			t.Errorf("len(resultCh) = %d, want %d", g, w)
		}
		for _, want := range wantResults {
			got := <-resultCh
			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("result (-got +want):\n%s", diff)
			}
		}
	})
}
