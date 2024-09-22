package logger

import (
	"fmt"
	"log/slog"
	"os"
)

func NewFileLogger(path string, logLevel slog.Level) (*slog.Logger, error) {
	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		os.FileMode(0644),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	levelVar := &slog.LevelVar{}
	levelVar.Set(logLevel)

	return slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{
		Level: levelVar,
	})), nil
}
