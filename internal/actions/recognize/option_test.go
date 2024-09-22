package recognize

import (
	"testing"
	"time"
)

func TestWithOutputFilePath(t *testing.T) {
	type args struct {
		outputFilePath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				outputFilePath: "test-output-file-path",
			},
			want: "test-output-file-path",
		},
		{
			name: "empty output file path",
			args: args{
				outputFilePath: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &options{}
			got := WithOutputFilePath(tt.args.outputFilePath)
			if err := got(opts); (err != nil) != tt.wantErr {
				t.Errorf("WithOutputFilePath()() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := opts.outputFilePath; got != tt.want {
				t.Errorf("WithOutputFilePath()() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithBufferSize(t *testing.T) {
	type args struct {
		bufferSize int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				bufferSize: 1024,
			},
			want: 1024,
		},
		{
			name: "buffer size less than 1024",
			args: args{
				bufferSize: 1023,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &options{}
			got := WithBufferSize(tt.args.bufferSize)
			if err := got(opts); (err != nil) != tt.wantErr {
				t.Errorf("WithBufferSize()() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := opts.bufferSize; got != tt.want {
				t.Errorf("WithBufferSize()() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithReconnectInterval(t *testing.T) {
	type args struct {
		reconnectInterval time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    time.Duration
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				reconnectInterval: time.Minute,
			},
			want: time.Minute,
		},
		{
			name: "reconnect interval less than 1 minute",
			args: args{
				reconnectInterval: time.Second,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &options{}
			got := WithReconnectInterval(tt.args.reconnectInterval)
			if err := got(opts); (err != nil) != tt.wantErr {
				t.Errorf("WithReconnectInterval()() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := opts.reconnectInterval; got != tt.want {
				t.Errorf("WithReconnectInterval()() = %v, want %v", got, tt.want)
			}
		})
	}
}
