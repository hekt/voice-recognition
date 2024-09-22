package recognize

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/hekt/voice-recognition/internal/interfaces/speech"
)

func Test_newRecognizer(t *testing.T) {
	type args struct {
		projectID         string
		recognizerName    string
		reconnectInterval time.Duration
		bufferSize        int
		client            speech.Client
		audioReader       io.Reader
		resultWriter      io.Writer
		interimWriter     io.Writer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				projectID:         "test-project-id",
				recognizerName:    "test-recognizer-name",
				reconnectInterval: time.Minute,
				bufferSize:        1024,
				client:            &speech.ClientMock{},
				audioReader:       &bytes.Buffer{},
				resultWriter:      &bytes.Buffer{},
				interimWriter:     &bytes.Buffer{},
			},
			wantErr: false,
		},
		{
			name: "invalid project ID",
			args: args{
				projectID:         "",
				recognizerName:    "test-recognizer-name",
				reconnectInterval: time.Minute,
				bufferSize:        1024,
				client:            &speech.ClientMock{},
				audioReader:       &bytes.Buffer{},
				resultWriter:      &bytes.Buffer{},
				interimWriter:     &bytes.Buffer{},
			},
			wantErr: true,
		},
		{
			name: "invalid recognizer name",
			args: args{
				projectID:         "test-project-id",
				recognizerName:    "",
				reconnectInterval: time.Minute,
				bufferSize:        1024,
				client:            &speech.ClientMock{},
				audioReader:       &bytes.Buffer{},
				resultWriter:      &bytes.Buffer{},
				interimWriter:     &bytes.Buffer{},
			},
			wantErr: true,
		},
		{
			name: "invalid buffer size",
			args: args{
				projectID:         "test-project-id",
				recognizerName:    "test-recognizer-name",
				reconnectInterval: time.Minute,
				bufferSize:        0,
				client:            &speech.ClientMock{},
				audioReader:       &bytes.Buffer{},
				resultWriter:      &bytes.Buffer{},
				interimWriter:     &bytes.Buffer{},
			},
			wantErr: true,
		},
		{
			name: "invalid reconnect interval",
			args: args{
				projectID:         "test-project-id",
				recognizerName:    "test-recognizer-name",
				reconnectInterval: 0,
				bufferSize:        1024,
				client:            &speech.ClientMock{},
				audioReader:       &bytes.Buffer{},
				resultWriter:      &bytes.Buffer{},
				interimWriter:     &bytes.Buffer{},
			},
			wantErr: true,
		},
		{
			name: "invalid client",
			args: args{
				projectID:         "test-project-id",
				recognizerName:    "test-recognizer-name",
				reconnectInterval: time.Minute,
				bufferSize:        1024,
				client:            nil,
				audioReader:       &bytes.Buffer{},
				resultWriter:      &bytes.Buffer{},
				interimWriter:     &bytes.Buffer{},
			},
			wantErr: true,
		},
		{
			name: "invalid audio reader",
			args: args{
				projectID:         "test-project-id",
				recognizerName:    "test-recognizer-name",
				reconnectInterval: time.Minute,
				bufferSize:        1024,
				client:            &speech.ClientMock{},
				audioReader:       nil,
				resultWriter:      &bytes.Buffer{},
				interimWriter:     &bytes.Buffer{},
			},
			wantErr: true,
		},
		{
			name: "invalid result writer",
			args: args{
				projectID:         "test-project-id",
				recognizerName:    "test-recognizer-name",
				reconnectInterval: time.Minute,
				bufferSize:        1024,
				client:            &speech.ClientMock{},
				audioReader:       &bytes.Buffer{},
				resultWriter:      nil,
				interimWriter:     &bytes.Buffer{},
			},
			wantErr: true,
		},
		{
			name: "invalid interim writer",
			args: args{
				projectID:         "test-project-id",
				recognizerName:    "test-recognizer-name",
				reconnectInterval: time.Minute,
				bufferSize:        1024,
				client:            &speech.ClientMock{},
				audioReader:       &bytes.Buffer{},
				resultWriter:      &bytes.Buffer{},
				interimWriter:     nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newRecognizer(
				tt.args.projectID,
				tt.args.recognizerName,
				tt.args.reconnectInterval,
				tt.args.bufferSize,
				tt.args.client,
				tt.args.audioReader,
				tt.args.resultWriter,
				tt.args.interimWriter,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("newRecognizer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("newRecognizer() = %v, want non-nil", got)
			}
		})
	}
}

func Test_recognizer_Start(t *testing.T) {
	type fields struct {
		streamSupplier    StreamSupplierInterface
		audioSender       AudioSenderInterface
		reseponseReceiver ResponseReceiverInterface
		responseProcessor ResponseProcessorInterface
		client            speech.Client
		responseCh        chan *speechpb.StreamingRecognizeResponse
		sendStreamCh      chan speechpb.Speech_StreamingRecognizeClient
		receiveStreamCh   chan speechpb.Speech_StreamingRecognizeClient
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				streamSupplier: &StreamSupplierInterfaceMock{
					SupplyFunc: func(ctx context.Context) error {
						return nil
					},
					StartFunc: func(ctx context.Context) error {
						return nil
					},
				},
				audioSender: &AudioSenderInterfaceMock{
					StartFunc: func(ctx context.Context) error {
						return nil
					},
				},
				reseponseReceiver: &ResponseReceiverInterfaceMock{
					StartFunc: func(ctx context.Context) error {
						return nil
					},
				},
				responseProcessor: &ResponseProcessorInterfaceMock{
					StartFunc: func(ctx context.Context) error {
						return nil
					},
				},
				client: &speech.ClientMock{
					CloseFunc: func() error {
						return nil
					},
				},
				responseCh:      make(chan *speechpb.StreamingRecognizeResponse),
				sendStreamCh:    make(chan speechpb.Speech_StreamingRecognizeClient),
				receiveStreamCh: make(chan speechpb.Speech_StreamingRecognizeClient),
			},
			wantErr: false,
		},
		{
			name: "failed to start stream supplier",
			fields: fields{
				streamSupplier: &StreamSupplierInterfaceMock{
					SupplyFunc: func(ctx context.Context) error {
						return nil
					},
					StartFunc: func(ctx context.Context) error {
						return errors.New("test error")
					},
				},
				audioSender: &AudioSenderInterfaceMock{
					StartFunc: func(ctx context.Context) error {
						return nil
					},
				},
				reseponseReceiver: &ResponseReceiverInterfaceMock{
					StartFunc: func(ctx context.Context) error {
						return nil
					},
				},
				responseProcessor: &ResponseProcessorInterfaceMock{
					StartFunc: func(ctx context.Context) error {
						return nil
					},
				},
				client: &speech.ClientMock{
					CloseFunc: func() error {
						return nil
					},
				},
				responseCh:      make(chan *speechpb.StreamingRecognizeResponse),
				sendStreamCh:    make(chan speechpb.Speech_StreamingRecognizeClient),
				receiveStreamCh: make(chan speechpb.Speech_StreamingRecognizeClient),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &recognizer{
				streamSupplier:    tt.fields.streamSupplier,
				audioSender:       tt.fields.audioSender,
				reseponseReceiver: tt.fields.reseponseReceiver,
				responseProcessor: tt.fields.responseProcessor,
				client:            tt.fields.client,
				responseCh:        tt.fields.responseCh,
				sendStreamCh:      tt.fields.sendStreamCh,
				receiveStreamCh:   tt.fields.receiveStreamCh,
			}
			if err := r.Start(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("recognizer.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
