package recognizer

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/speech/apiv2/speechpb"
	myspeech "github.com/hekt/voice-recognition/internal/interfaces/speech"
	myspeechpb "github.com/hekt/voice-recognition/internal/interfaces/speechpb"
	"github.com/hekt/voice-recognition/internal/recognizer/model"
	"github.com/hekt/voice-recognition/internal/testutil"
)

func Test_New(t *testing.T) {
	type args struct {
		projectID         string
		recognizerName    string
		reconnectInterval time.Duration
		bufferSize        int
		inactiveTimeout   time.Duration
		client            myspeech.Client
		audioReader       io.Reader
		resultWriter      io.Writer
		interimWriter     io.Writer
	}
	validArgs := args{
		projectID:         "test-project-id",
		recognizerName:    "test-recognizer-name",
		reconnectInterval: time.Minute,
		bufferSize:        1024,
		inactiveTimeout:   time.Minute,
		client:            &myspeech.ClientMock{},
		audioReader:       &bytes.Buffer{},
		resultWriter:      &bytes.Buffer{},
		interimWriter:     &bytes.Buffer{},
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "success",
			args:    validArgs,
			wantErr: false,
		},
		{
			name: "invalid project ID",
			args: func() args {
				a := validArgs
				a.projectID = ""
				return a
			}(),
			wantErr: true,
		},
		{
			name: "invalid recognizer name",
			args: func() args {
				a := validArgs
				a.recognizerName = ""
				return a
			}(),
			wantErr: true,
		},
		{
			name: "invalid buffer size",
			args: func() args {
				a := validArgs
				a.bufferSize = 0
				return a
			}(),
			wantErr: true,
		},
		{
			name: "invalid reconnect interval",
			args: func() args {
				a := validArgs
				a.reconnectInterval = 0
				return a
			}(),
			wantErr: true,
		},
		{
			name: "invalid inactive timeout",
			args: func() args {
				a := validArgs
				a.inactiveTimeout = 0
				return a
			}(),
			wantErr: true,
		},
		{
			name: "invalid client",
			args: func() args {
				a := validArgs
				a.client = nil
				return a
			}(),
			wantErr: true,
		},
		{
			name: "invalid audio reader",
			args: func() args {
				a := validArgs
				a.audioReader = nil
				return a
			}(),
			wantErr: true,
		},
		{
			name: "invalid result writer",
			args: func() args {
				a := validArgs
				a.resultWriter = nil
				return a
			}(),
			wantErr: true,
		},
		{
			name: "invalid interim writer",
			args: func() args {
				a := validArgs
				a.interimWriter = nil
				return a
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(
				tt.args.projectID,
				tt.args.recognizerName,
				tt.args.reconnectInterval,
				tt.args.bufferSize,
				tt.args.inactiveTimeout,
				tt.args.client,
				tt.args.audioReader,
				tt.args.resultWriter,
				tt.args.interimWriter,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("New() = %v, want non-nil", got)
			}
		})
	}
}

func Test_Recognizer_Start(t *testing.T) {
	type fields struct {
		streamSupplier    StreamSupplierInterface
		audioReceiver     AudioReceiverInterface
		audioSender       AudioSenderInterface
		reseponseReceiver ResponseReceiverInterface
		responseProcessor ResponseProcessorInterface
		resultWriter      ResultWriterInterface
		processMonitor    ProcessMonitorInterface
		client            myspeech.Client
		audioCh           chan []byte
		responseCh        chan *speechpb.StreamingRecognizeResponse
		sendStreamCh      chan speechpb.Speech_StreamingRecognizeClient
		receiveStreamCh   chan speechpb.Speech_StreamingRecognizeClient
		processCh         chan struct{}
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
					SupplyFunc: func(context.Context) error {
						return nil
					},
					StartFunc: func(context.Context) error {
						return nil
					},
				},
				audioReceiver: &AudioReceiverInterfaceMock{
					StartFunc: func(context.Context) error {
						return nil
					},
				},
				audioSender: &AudioSenderInterfaceMock{
					StartFunc: func(context.Context) error {
						return nil
					},
				},
				reseponseReceiver: &ResponseReceiverInterfaceMock{
					StartFunc: func(context.Context) error {
						return nil
					},
				},
				responseProcessor: &ResponseProcessorInterfaceMock{
					StartFunc: func(context.Context) error {
						return nil
					},
				},
				resultWriter: &ResultWriterInterfaceMock{
					StartFunc: func(context.Context) error {
						return nil
					},
				},
				processMonitor: &ProcessMonitorInterfaceMock{
					StartFunc: func(context.Context) error {
						return nil
					},
				},
				client: &myspeech.ClientMock{
					CloseFunc: func() error {
						return nil
					},
				},
				audioCh:         make(chan []byte),
				responseCh:      make(chan *speechpb.StreamingRecognizeResponse),
				sendStreamCh:    make(chan speechpb.Speech_StreamingRecognizeClient),
				receiveStreamCh: make(chan speechpb.Speech_StreamingRecognizeClient),
				processCh:       make(chan struct{}),
			},
			wantErr: false,
		},
		{
			name: "failed to start stream supplier",
			fields: fields{
				streamSupplier: &StreamSupplierInterfaceMock{
					SupplyFunc: func(context.Context) error {
						return nil
					},
					StartFunc: func(context.Context) error {
						return errors.New("test error")
					},
				},
				audioReceiver: &AudioReceiverInterfaceMock{
					StartFunc: func(context.Context) error {
						return nil
					},
				},
				audioSender: &AudioSenderInterfaceMock{
					StartFunc: func(context.Context) error {
						return nil
					},
				},
				reseponseReceiver: &ResponseReceiverInterfaceMock{
					StartFunc: func(context.Context) error {
						return nil
					},
				},
				responseProcessor: &ResponseProcessorInterfaceMock{
					StartFunc: func(context.Context) error {
						return nil
					},
				},
				resultWriter: &ResultWriterInterfaceMock{
					StartFunc: func(context.Context) error {
						return nil
					},
				},
				processMonitor: &ProcessMonitorInterfaceMock{
					StartFunc: func(context.Context) error {
						return nil
					},
				},
				client: &myspeech.ClientMock{
					CloseFunc: func() error {
						return nil
					},
				},
				audioCh:         make(chan []byte),
				responseCh:      make(chan *speechpb.StreamingRecognizeResponse),
				sendStreamCh:    make(chan speechpb.Speech_StreamingRecognizeClient),
				receiveStreamCh: make(chan speechpb.Speech_StreamingRecognizeClient),
				processCh:       make(chan struct{}),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Recognizer{
				streamSupplier:    tt.fields.streamSupplier,
				audioReceiver:     tt.fields.audioReceiver,
				audioSender:       tt.fields.audioSender,
				reseponseReceiver: tt.fields.reseponseReceiver,
				responseProcessor: tt.fields.responseProcessor,
				resultWriter:      tt.fields.resultWriter,
				processMonitor:    tt.fields.processMonitor,
				client:            tt.fields.client,
				audioCh:           tt.fields.audioCh,
				responseCh:        tt.fields.responseCh,
				sendStreamCh:      tt.fields.sendStreamCh,
				receiveStreamCh:   tt.fields.receiveStreamCh,
				processCh:         tt.fields.processCh,
			}
			if err := r.Start(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("recognizer.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_Recognizer_Start_Dataflow(t *testing.T) {
	t.Run("read,send", func(t *testing.T) {
		// io reader
		audioReader := testutil.NewChannelReader()

		// channels
		audioCh := make(chan []byte)
		sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient)
		receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient)
		responseCh := make(chan *speechpb.StreamingRecognizeResponse)
		processCh := make(chan struct{})

		// clients
		sendCalls := make(chan struct{})
		closeSendCalls := make(chan struct{})
		streamMock := &myspeechpb.Speech_StreamingRecognizeClientMock{
			SendFunc: func(*speechpb.StreamingRecognizeRequest) error {
				defer func() { sendCalls <- struct{}{} }()
				return nil
			},
			CloseSendFunc: func() error {
				defer func() { closeSendCalls <- struct{}{} }()
				return nil
			},
		}
		clientCloseCalls := make(chan struct{})
		clientMock := &myspeech.ClientMock{
			CloseFunc: func() error {
				defer func() { clientCloseCalls <- struct{}{} }()
				return nil
			},
		}

		// workers
		audioReceiver := NewAudioReceiver(audioReader, audioCh, 4)
		audioSender := NewAudioSender(audioCh, sendStreamCh)
		streamSupplier := &StreamSupplierInterfaceMock{
			SupplyFunc: func(context.Context) error {
				return nil
			},
			StartFunc: func(context.Context) error {
				return nil
			},
		}
		responseReceiver := &ResponseReceiverInterfaceMock{
			StartFunc: func(context.Context) error {
				return nil
			},
		}
		responseProcessor := &ResponseProcessorInterfaceMock{
			StartFunc: func(context.Context) error {
				return nil
			},
		}
		resultWriter := &ResultWriterInterfaceMock{
			StartFunc: func(context.Context) error {
				return nil
			},
		}
		processMonitor := &ProcessMonitorInterfaceMock{
			StartFunc: func(context.Context) error {
				return nil
			},
		}

		r := &Recognizer{
			streamSupplier:    streamSupplier,
			audioReceiver:     audioReceiver,
			audioSender:       audioSender,
			reseponseReceiver: responseReceiver,
			responseProcessor: responseProcessor,
			resultWriter:      resultWriter,
			processMonitor:    processMonitor,
			client:            clientMock,
			audioCh:           audioCh,
			responseCh:        responseCh,
			sendStreamCh:      sendStreamCh,
			receiveStreamCh:   receiveStreamCh,
			processCh:         processCh,
		}

		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup
		wg.Add(1)
		var got error
		go func() {
			defer wg.Done()
			got = r.Start(ctx)
		}()

		// Suplly stream to AudioSender.
		sendStreamCh <- streamMock

		// Send bytes to reader to be read by AudioReceiver.
		audioReader.BufCh <- bytes.Repeat([]byte("a"), 4)
		// AudioReceiver reads bytes and sends to audioCh.
		// AudioSender reads bytes from audioCh and sends to stream.
		<-sendCalls

		// Send EOF to reader.
		audioReader.EOFCh <- struct{}{}
		// AudioReceiver reads EOF.

		// Close context to stop recognizer.
		cancel()
		// AudioSender detects context canceled and closes stream
		<-closeSendCalls
		// recognizer closes client
		<-clientCloseCalls

		wg.Wait()

		if !errors.Is(got, context.Canceled) {
			t.Errorf("recognizer.Start() error = %v, want %v", got, context.Canceled)
		}
	})

	t.Run("receive,process,write", func(t *testing.T) {
		// io writers
		ioResultWriter := &bytes.Buffer{}
		ioInterimWriter := &bytes.Buffer{}

		// channels
		audioCh := make(chan []byte, 1)
		sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient)
		receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient)
		responseCh := make(chan *speechpb.StreamingRecognizeResponse)
		resultCh := make(chan []*model.Result)
		processCh := make(chan struct{}, 4)

		// clients
		recvResponseCh := make(chan *speechpb.StreamingRecognizeResponse)
		streamMock := &myspeechpb.Speech_StreamingRecognizeClientMock{
			RecvFunc: func() (*speechpb.StreamingRecognizeResponse, error) {
				resp, ok := <-recvResponseCh
				if !ok {
					return nil, io.EOF
				}
				return resp, nil
			},
		}
		clientCloseCalls := make(chan struct{})
		clientMock := &myspeech.ClientMock{
			CloseFunc: func() error {
				defer func() { clientCloseCalls <- struct{}{} }()
				return nil
			},
		}

		// workers
		audioReceiver := &AudioReceiverInterfaceMock{
			StartFunc: func(context.Context) error {
				return nil
			},
		}
		audioSender := &AudioSenderInterfaceMock{
			StartFunc: func(context.Context) error {
				return nil
			},
		}
		streamSupplier := &StreamSupplierInterfaceMock{
			SupplyFunc: func(context.Context) error {
				return nil
			},
			StartFunc: func(context.Context) error {
				return nil
			},
		}
		responseReceiver := NewResponseReceiver(responseCh, receiveStreamCh)
		responseProcessor := NewResponseProcessor(responseCh, resultCh, processCh)
		resultWriter := NewResultWriter(resultCh, ioResultWriter, ioInterimWriter)
		processMonitor := NewProcessMonitor(processCh, time.Minute)

		r := &Recognizer{
			streamSupplier:    streamSupplier,
			audioReceiver:     audioReceiver,
			audioSender:       audioSender,
			reseponseReceiver: responseReceiver,
			responseProcessor: responseProcessor,
			resultWriter:      resultWriter,
			processMonitor:    processMonitor,
			client:            clientMock,
			audioCh:           audioCh,
			responseCh:        responseCh,
			sendStreamCh:      sendStreamCh,
			receiveStreamCh:   receiveStreamCh,
			processCh:         processCh,
		}

		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup
		wg.Add(1)
		var got error
		go func() {
			defer wg.Done()
			got = r.Start(ctx)
		}()

		// Suplly stream to ResponseReceiver.
		receiveStreamCh <- streamMock

		// Send responses.
		responseBuilder := func(transcript string, isFinal bool) *speechpb.StreamingRecognizeResponse {
			return &speechpb.StreamingRecognizeResponse{
				Results: []*speechpb.StreamingRecognitionResult{
					{
						Alternatives: []*speechpb.SpeechRecognitionAlternative{
							{
								Transcript: transcript,
							},
						},
						IsFinal: isFinal,
					},
				},
			}
		}
		recvResponseCh <- responseBuilder("aaaa", false)
		recvResponseCh <- responseBuilder("bbbb", false)
		recvResponseCh <- responseBuilder("aaaabbbbc", true)
		close(recvResponseCh)

		// wait for process to finish
		time.Sleep(100 * time.Millisecond)

		// Close context to stop recognizer.
		cancel()

		<-clientCloseCalls

		wg.Wait()

		if !errors.Is(got, context.Canceled) {
			t.Errorf("recognizer.Start() error = %v, want %v", got, context.Canceled)
		}
		if g, w := ioResultWriter.String(), "aaaabbbbc"; g != w {
			t.Errorf("resultWriter = %q, want %q", g, w)
		}
		if g, w := ioInterimWriter.String(), "aaaabbbb"; g != w {
			t.Errorf("interimWriter = %q, want %q", g, w)
		}
	})
}
