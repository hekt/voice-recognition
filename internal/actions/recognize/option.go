package recognize

import (
	"errors"
	"time"
)

type options struct {
	outputFilePath    string
	bufferSize        int
	reconnectInterval time.Duration
}

type Option func(*options) error

func WithOutputFilePath(outputFilePath string) Option {
	return func(o *options) error {
		if outputFilePath == "" {
			return errors.New("output file path must be 1 or more characters")
		}
		o.outputFilePath = outputFilePath
		return nil
	}
}

func WithBufferSize(bufferSize int) Option {
	return func(o *options) error {
		if bufferSize < 1024 {
			return errors.New("buffer size must be greater than or equal to 1024")
		}
		o.bufferSize = bufferSize
		return nil
	}
}

func WithReconnectInterval(reconnectInterval time.Duration) Option {
	return func(o *options) error {
		if reconnectInterval < time.Minute {
			return errors.New("reconnect interval must be greater than or equal to 1 minute")
		}
		o.reconnectInterval = reconnectInterval
		return nil
	}
}
