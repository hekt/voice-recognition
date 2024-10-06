package recognizer

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/gax-go/v2"
	ispeech "github.com/hekt/voice-recognition/internal/interfaces/speech"
	ispeechpb "github.com/hekt/voice-recognition/internal/interfaces/speechpb"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestNewStreamSupplier(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &ispeech.ClientMock{}
		sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient)
		receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient)
		recognizerFullName := "projects/test-project/locations/global/recognizers/test-recognizer"
		supplyInterval := 5 * time.Minute

		got := NewStreamSupplier(client, sendStreamCh, receiveStreamCh, recognizerFullName, supplyInterval)
		want := &StreamSupplier{
			client:             client,
			sendStreamCh:       sendStreamCh,
			receiveStreamCh:    receiveStreamCh,
			recognizerFullName: recognizerFullName,
			supplyInterval:     supplyInterval,
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("NewStreamSupplier() = %v, want %v", got, want)
		}
	})
}

func Test_streamSupplier_Start(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// 停止させるためのコンテキスト
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// initializeStream で使われる
		stream := &ispeechpb.Speech_StreamingRecognizeClientMock{
			SendFunc: func(_ *speechpb.StreamingRecognizeRequest) error {
				return nil
			},
		}
		client := &ispeech.ClientMock{
			StreamingRecognizeFunc: func(
				_ context.Context,
				_ ...gax.CallOption,
			) (speechpb.Speech_StreamingRecognizeClient, error) {
				return stream, nil
			},
		}

		// ブロッキングの動作を確認するためにバッファを 1 にしている。
		// 実際の動作でもバッファは 1 で使う想定。
		sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)
		receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)

		s := &StreamSupplier{
			client:          client,
			supplyInterval:  10 * time.Millisecond,
			sendStreamCh:    sendStreamCh,
			receiveStreamCh: receiveStreamCh,
		}

		var wg sync.WaitGroup
		var got error
		wg.Add(1)
		go func() {
			defer wg.Done()
			got = s.Start(ctx)
		}()

		// 2回取り出す
		for i := 0; i < 2; i++ {
			gotSendStream := <-sendStreamCh
			if gotSendStream != stream {
				t.Errorf("streamSupplier.Start() supplies %v, want %v, attempt = %d", gotSendStream, stream, i+1)
			}
			gotReceiveStream := <-receiveStreamCh
			if gotReceiveStream != stream {
				t.Errorf("streamSupplier.Start() supplies %v, want %v, attempt = %d", gotReceiveStream, stream, i+1)
			}
		}

		// stream を取り出したあとの最後の処理を確実におこなうために待機
		time.Sleep(100 * time.Millisecond)

		// キャンセルしてループを停止させ、goroutine が完了するのを待つ
		cancel()
		wg.Wait()

		if !errors.Is(got, context.Canceled) {
			t.Errorf("streamSupplier.Start() error = %v, want %v", got, context.Canceled)
		}

		// バッファが 1 なので、二回取り出したあと三回目を送信して四回目で待機していたはず
		if len(client.StreamingRecognizeCalls()) != 4 {
			t.Errorf("streamSupplier.Start() calls StreamingRecognize %d times, want 4 times", len(client.StreamingRecognizeCalls()))
		}
	})

	t.Run("initializeStream error", func(t *testing.T) {
		// initializeStream で使われる
		client := &ispeech.ClientMock{
			StreamingRecognizeFunc: func(
				_ context.Context,
				_ ...gax.CallOption,
			) (speechpb.Speech_StreamingRecognizeClient, error) {
				return nil, errors.New("test")
			},
		}

		s := &StreamSupplier{
			client:         client,
			supplyInterval: 10 * time.Millisecond,
		}
		if err := s.Start(context.Background()); err == nil {
			t.Errorf("recognizer.startStreamSupplier() error = %v, wantErr %v", err, true)
		}
	})
}

func Test_streamSupplier_Supply(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// initializeStream で使われる
		stream := &ispeechpb.Speech_StreamingRecognizeClientMock{
			SendFunc: func(_ *speechpb.StreamingRecognizeRequest) error {
				return nil
			},
		}
		client := &ispeech.ClientMock{
			StreamingRecognizeFunc: func(
				_ context.Context,
				_ ...gax.CallOption,
			) (speechpb.Speech_StreamingRecognizeClient, error) {
				return stream, nil
			},
		}

		sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)
		receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)

		s := &StreamSupplier{
			client:          client,
			sendStreamCh:    sendStreamCh,
			receiveStreamCh: receiveStreamCh,
		}

		if err := s.Supply(context.Background()); err != nil {
			t.Errorf("streamSupplier.Supply() error = %v, want nil", err)
		}
		if gotSendStream := <-sendStreamCh; gotSendStream != stream {
			t.Errorf("streamSupplier.Supply() supplies %v, want %v", gotSendStream, stream)
		}
		if gotReceiveStream := <-receiveStreamCh; gotReceiveStream != stream {
			t.Errorf("streamSupplier.Supply() supplies %v, want %v", gotReceiveStream, stream)
		}
	})

	t.Run("buffer is full", func(t *testing.T) {
		// initializeStream で使われる
		stream := &ispeechpb.Speech_StreamingRecognizeClientMock{
			SendFunc: func(_ *speechpb.StreamingRecognizeRequest) error {
				return nil
			},
		}
		client := &ispeech.ClientMock{
			StreamingRecognizeFunc: func(
				_ context.Context,
				_ ...gax.CallOption,
			) (speechpb.Speech_StreamingRecognizeClient, error) {
				return stream, nil
			},
		}

		sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)
		receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)

		s := &StreamSupplier{
			client:          client,
			sendStreamCh:    sendStreamCh,
			receiveStreamCh: receiveStreamCh,
		}

		ctx := context.Background()

		// 一度実行してバッファを埋める
		if err := s.Supply(ctx); err != nil {
			t.Errorf("streamSupplier.Supply() error = %v, want nil", err)
		}

		var wg sync.WaitGroup
		wg.Add(1)
		done := false
		go func() {
			defer wg.Done()
			if err := s.Supply(ctx); err != nil {
				t.Errorf("streamSupplier.Supply() error = %v, want nil", err)
			}
			done = true
		}()

		time.Sleep(100 * time.Millisecond)

		// 一度目の Supply でバッファが埋まっているので、二度目の Supply はブロックされる
		if done {
			t.Error("streamSupplier.Supply() is done, want block")
		}

		// バッファを空にする
		<-sendStreamCh
		<-receiveStreamCh

		wg.Wait()

		// バッファが空になったので、ブロックが解消され完了する
		if !done {
			t.Error("streamSupplier.Supply() is not done, want done")
		}
	})

	t.Run("context canceled", func(t *testing.T) {
		// initializeStream で使われる
		stream := &ispeechpb.Speech_StreamingRecognizeClientMock{
			SendFunc: func(_ *speechpb.StreamingRecognizeRequest) error {
				return nil
			},
		}
		client := &ispeech.ClientMock{
			StreamingRecognizeFunc: func(
				_ context.Context,
				_ ...gax.CallOption,
			) (speechpb.Speech_StreamingRecognizeClient, error) {
				return stream, nil
			},
		}

		sendStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)
		receiveStreamCh := make(chan speechpb.Speech_StreamingRecognizeClient, 1)

		s := &StreamSupplier{
			client:          client,
			sendStreamCh:    sendStreamCh,
			receiveStreamCh: receiveStreamCh,
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// 一度実行してバッファを埋める
		if err := s.Supply(ctx); err != nil {
			t.Errorf("streamSupplier.Supply() error = %v, want nil", err)
		}

		var wg sync.WaitGroup
		wg.Add(1)
		var got error
		go func() {
			defer wg.Done()
			got = s.Supply(ctx)
		}()

		cancel()
		wg.Wait()

		if !errors.Is(got, context.Canceled) {
			t.Errorf("streamSupplier.Supply() error = %v, want %v", got, context.Canceled)
		}
	})
}

func Test_streamSupplier_initializeStream(t *testing.T) {
	recognizer := "projects/test-project/locations/global/recognizers/test-recognizer"

	type fields struct {
		client             ispeech.Client
		recognizerFullName string
	}
	type test struct {
		name    string
		fields  fields
		want    speechpb.Speech_StreamingRecognizeClient
		wantErr bool
	}
	tests := []test{
		// success
		func() test {
			stream := &ispeechpb.Speech_StreamingRecognizeClientMock{
				SendFunc: func(
					req *speechpb.StreamingRecognizeRequest,
				) error {
					if req.Recognizer != recognizer {
						t.Errorf(
							"recognizer.initializeStream() req.Recognizer = %v, want %v",
							req.Recognizer,
							recognizer,
						)
					}
					wantStreamingRequest := &speechpb.StreamingRecognizeRequest_StreamingConfig{
						StreamingConfig: &speechpb.StreamingRecognitionConfig{
							StreamingFeatures: &speechpb.StreamingRecognitionFeatures{
								InterimResults: true,
							},
						},
					}
					if diff := cmp.Diff(req.StreamingRequest, wantStreamingRequest, protocmp.Transform()); diff != "" {
						t.Errorf("recognizer.initializeStream() req.StreamingRequest (-got +want) = %v", diff)
					}

					return nil
				},
			}
			client := &ispeech.ClientMock{
				StreamingRecognizeFunc: func(
					_ context.Context,
					_ ...gax.CallOption,
				) (speechpb.Speech_StreamingRecognizeClient, error) {
					return stream, nil
				},
			}
			return test{
				name: "success",
				fields: fields{
					recognizerFullName: recognizer,
					client:             client,
				},
				want:    stream,
				wantErr: false,
			}
		}(),
		// client initialization error
		func() test {
			client := &ispeech.ClientMock{
				StreamingRecognizeFunc: func(
					_ context.Context,
					_ ...gax.CallOption,
				) (speechpb.Speech_StreamingRecognizeClient, error) {
					return nil, errors.New("test")
				},
			}
			return test{
				name: "client initialization error",
				fields: fields{
					recognizerFullName: recognizer,
					client:             client,
				},
				want:    nil,
				wantErr: true,
			}
		}(),
		// stream send error
		func() test {
			stream := &ispeechpb.Speech_StreamingRecognizeClientMock{
				SendFunc: func(
					_ *speechpb.StreamingRecognizeRequest,
				) error {
					return errors.New("test")
				},
			}
			client := &ispeech.ClientMock{
				StreamingRecognizeFunc: func(
					_ context.Context,
					_ ...gax.CallOption,
				) (speechpb.Speech_StreamingRecognizeClient, error) {
					return stream, nil
				},
			}
			return test{
				name: "stream send error",
				fields: fields{
					recognizerFullName: recognizer,
					client:             client,
				},
				want:    nil,
				wantErr: true,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			s := &StreamSupplier{
				client:             tt.fields.client,
				recognizerFullName: tt.fields.recognizerFullName,
			}
			got, err := s.initializeStream(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("streamSupplier.initializeStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("streamSupplier.initializeStream() = %v, want %v", got, tt.want)
			}
		})
	}
}
