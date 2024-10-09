package recognizer

import (
	"context"
	"errors"
	"time"
)

//go:generate moq -rm -out process_monitor_mock.go . ProcessMonitorInterface
type ProcessMonitorInterface interface {
	Start(context.Context) error
}

var _ ProcessMonitorInterface = &ProcessMonitor{}

type ProcessMonitor struct {
	processCh       <-chan struct{}
	timeoutDuration time.Duration
}

func NewProcessMonitor(
	processCh <-chan struct{},
	timeoutDuration time.Duration,
) *ProcessMonitor {
	return &ProcessMonitor{
		processCh:       processCh,
		timeoutDuration: timeoutDuration,
	}
}

func (m *ProcessMonitor) Start(ctx context.Context) error {
	timer := time.NewTimer(m.timeoutDuration)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-m.processCh:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(m.timeoutDuration)
		case <-timer.C:
			return errors.New("inactive for a long time")
		}
	}
}
