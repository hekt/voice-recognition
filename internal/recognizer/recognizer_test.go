package recognizer

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	myspeech "github.com/hekt/voice-recognition/internal/interfaces/speech"
	"github.com/hekt/voice-recognition/internal/recognizer/model"
	"github.com/hekt/voice-recognition/internal/testutil"
)

func Test_New(t *testing.T) {
	type args struct {
		client            myspeech.Client
		projectID         string
		recognizerName    string
		reconnectInterval time.Duration
		bufferSize        int
		inactiveTimeout   time.Duration
		audioReader       io.Reader
		resultWriter      io.Writer
		interimWriter     io.Writer
	}
	validArgs := args{
		client:            &myspeech.ClientMock{},
		projectID:         "test-project-id",
		recognizerName:    "test-recognizer-name",
		reconnectInterval: time.Minute,
		bufferSize:        1024,
		inactiveTimeout:   time.Minute,
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
			name: "invalid buffer size",
			args: func() args {
				a := validArgs
				a.bufferSize = 0
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
			ctx := context.Background()
			got, err := New(
				ctx,
				tt.args.client,
				tt.args.projectID,
				tt.args.recognizerName,
				tt.args.reconnectInterval,
				tt.args.bufferSize,
				tt.args.inactiveTimeout,
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
		recognizer     model.RecognizerCoreInterface
		audioReader    AudioReaderInterface
		resultWriter   ResultWriterInterface
		processMonitor ProcessMonitorInterface
		audioCh        chan []byte
		resultCh       chan []*model.Result
		processCh      chan struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				recognizer: &model.RecognizerCoreInterfaceMock{
					StartFunc: func(context.Context) error {
						return nil
					},
				},
				audioReader: &AudioReaderInterfaceMock{
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
				audioCh:   make(chan []byte),
				resultCh:  make(chan []*model.Result),
				processCh: make(chan struct{}),
			},
			wantErr: false,
		},
		{
			name: "failed to start recognizer",
			fields: fields{
				recognizer: &model.RecognizerCoreInterfaceMock{
					StartFunc: func(context.Context) error {
						return errors.New("test error")
					},
				},
				audioReader: &AudioReaderInterfaceMock{
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
				audioCh:   make(chan []byte),
				resultCh:  make(chan []*model.Result),
				processCh: make(chan struct{}),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Recognizer{
				recognizer:     tt.fields.recognizer,
				audioReader:    tt.fields.audioReader,
				resultWriter:   tt.fields.resultWriter,
				processMonitor: tt.fields.processMonitor,
				audioCh:        tt.fields.audioCh,
				resultCh:       tt.fields.resultCh,
				processCh:      tt.fields.processCh,
			}
			if err := r.Start(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("recognizer.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_Recognizer_Start_Dataflow(t *testing.T) {
	t.Run("read,recognize,write", func(t *testing.T) {
		chunkSize := 4

		ctx, cancel := context.WithCancel(context.Background())

		// io reader/writer
		ioAudioReader := testutil.NewChannelReader()
		ioResultWriter := &bytes.Buffer{}
		ioInterimWriter := &bytes.Buffer{}

		// channels
		audioCh := make(chan []byte)
		resultCh := make(chan []*model.Result)
		processCh := make(chan struct{}, 3)

		// workers
		recognizer := &model.RecognizerCoreInterfaceMock{
			StartFunc: func(ctx context.Context) error {
				for {
					var d []byte
					select {
					case <-ctx.Done():
						return ctx.Err()
					case d = <-audioCh:
					}

					result := &model.Result{
						Transcript: string(d),
						IsFinal:    len(d) < chunkSize,
					}

					select {
					case <-ctx.Done():
						return ctx.Err()
					case resultCh <- []*model.Result{result}:
					}
				}
			},
		}
		audioReader := NewAudioReceiver(ioAudioReader, audioCh, 4)
		resultWriter := NewResultWriter(
			resultCh,
			&NotifyingWriter{
				Writer:   ioResultWriter,
				NotifyCh: processCh,
			},
			&NotifyingWriter{
				Writer:   ioInterimWriter,
				NotifyCh: processCh,
			},
		)
		processMonitor := &ProcessMonitorInterfaceMock{
			StartFunc: func(context.Context) error {
				return nil
			},
		}

		r := &Recognizer{
			recognizer:     recognizer,
			audioReader:    audioReader,
			resultWriter:   resultWriter,
			processMonitor: processMonitor,
			audioCh:        audioCh,
			resultCh:       resultCh,
			processCh:      processCh,
		}

		var wg sync.WaitGroup
		wg.Add(1)
		var got error
		go func() {
			defer wg.Done()
			got = r.Start(ctx)
		}()

		ioAudioReader.BufCh <- bytes.Repeat([]byte("a"), chunkSize)
		ioAudioReader.BufCh <- bytes.Repeat([]byte("b"), chunkSize)
		ioAudioReader.BufCh <- bytes.Repeat([]byte("c"), chunkSize-1)
		ioAudioReader.EOFCh <- struct{}{}

		// wait for the result to be written
		for {
			time.Sleep(10 * time.Millisecond)
			if len(processCh) == 3 {
				break
			}
		}

		cancel()
		wg.Wait()

		if !errors.Is(got, context.Canceled) {
			t.Errorf("recognizer.Start() error = %v, want %v", got, context.Canceled)
		}
		if g, w := ioResultWriter.String(), "ccc"; g != w {
			t.Errorf("result writer = %q, want %q", g, w)
		}
		if g, w := ioInterimWriter.String(), "aaaabbbb"; g != w {
			t.Errorf("interim writer = %q, want %q", g, w)
		}
	})
}
