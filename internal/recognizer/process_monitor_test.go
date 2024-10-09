package recognizer

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestNewProcessMonitor(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		processCh := make(chan struct{})
		timeoutDuration := 1 * time.Second
		want := &ProcessMonitor{
			processCh:       processCh,
			timeoutDuration: timeoutDuration,
		}

		if got := NewProcessMonitor(processCh, timeoutDuration); !reflect.DeepEqual(got, want) {
			t.Errorf("NewProcessMonitor() = %v, want %v", got, want)
		}
	})
}

func TestProcessMonitor_Start(t *testing.T) {
	t.Run("timeout", func(t *testing.T) {
		processCh := make(chan struct{})
		timeoutDuration := time.Microsecond
		m := NewProcessMonitor(processCh, timeoutDuration)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wantMsg := "inactive for a long time"
		if got := m.Start(ctx, cancel); got != nil && got.Error() != wantMsg {
			t.Errorf("ProcessMonitor.Start() = %v, want %v", got, wantMsg)
		}

		select {
		case <-ctx.Done():
		default:
			t.Error("context is not canceled")
		}
	})

	t.Run("canceled by others", func(t *testing.T) {
		processCh := make(chan struct{})
		timeoutDuration := time.Hour
		m := NewProcessMonitor(processCh, timeoutDuration)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var wg sync.WaitGroup
		wg.Add(1)
		var got error
		go func() {
			defer wg.Done()
			got = m.Start(ctx, cancel)
		}()

		cancel()
		wg.Wait()

		if !errors.Is(got, context.Canceled) {
			t.Errorf("ProcessMonitor.Start() = %v, want %v", got, context.Canceled)
		}
	})

	t.Run("extend timeout", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		processCh := make(chan struct{})
		m := &ProcessMonitor{
			processCh:       processCh,
			timeoutDuration: 100 * time.Millisecond,
		}

		var wg sync.WaitGroup
		wg.Add(1)
		var got error
		go func() {
			defer wg.Done()
			got = m.Start(ctx, cancel)
		}()

		// extend timeout every 50ms
		processCtx, processCancel := context.WithCancel(context.Background())
		defer processCancel()
		ticker := time.NewTicker(50 * time.Millisecond)
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-processCtx.Done():
				return
			case <-ticker.C:
				processCh <- struct{}{}
			}
		}()

		// If the timeout has not occurred after waiting for 150ms, the timeout has been extended.
		time.Sleep(150 * time.Millisecond)
		select {
		case <-ctx.Done():
			t.Fatalf("unexpected timeout")
		default:
		}

		// Stop extending timeout
		processCancel()

		// Wait for the timeout to occur
		<-ctx.Done()

		wg.Wait()

		wantMsg := "inactive for a long time"
		if got != nil && got.Error() != wantMsg {
			t.Errorf("ProcessMonitor.Start() = %v, want %v", got, wantMsg)
		}
	})
}
