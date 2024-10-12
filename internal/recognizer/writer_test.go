package recognizer

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDecoratedInterimWriter_Write(t *testing.T) {
	wantFormat := "\033[H\033[2J" + "\033[32m" + "%s" + "\033[0m"

	t.Run("write once", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := &DecoratedInterimWriter{Writer: buf}
		if _, err := w.Write([]byte("test")); err != nil {
			t.Errorf("Write() error = %v, wantErr %v", err, false)
		}

		want := fmt.Sprintf(wantFormat, "test")
		if got := buf.String(); got != want {
			t.Errorf("Write() writes %v, want %v", got, want)
		}
	})

	t.Run("write twice", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := &DecoratedInterimWriter{Writer: buf}
		if _, err := w.Write([]byte("test1")); err != nil {
			t.Errorf("Write() error = %v, wantErr %v", err, false)
		}
		if _, err := w.Write([]byte("test2")); err != nil {
			t.Errorf("Write() error = %v, wantErr %v", err, false)
		}

		want := fmt.Sprintf(wantFormat, "test1") + fmt.Sprintf(wantFormat, "test2")
		if got := buf.String(); got != want {
			t.Errorf("Write() writes %v, want %v", got, want)
		}
	})
}
func TestDecoratedResultWriter_Write(t *testing.T) {
	wantFormat := "\n%s"

	t.Run("write once", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := &DecoratedResultWriter{Writer: buf}
		if _, err := w.Write([]byte("test")); err != nil {
			t.Errorf("Write() error = %v, wantErr %v", err, false)
		}

		want := fmt.Sprintf(wantFormat, "test")
		if got := buf.String(); got != want {
			t.Errorf("Write() writes %v, want %v", got, want)
		}
	})

	t.Run("write twice", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := &DecoratedResultWriter{Writer: buf}
		if _, err := w.Write([]byte("test1")); err != nil {
			t.Errorf("Write() error = %v, wantErr %v", err, false)
		}
		if _, err := w.Write([]byte("test2")); err != nil {
			t.Errorf("Write() error = %v, wantErr %v", err, false)
		}

		want := fmt.Sprintf(wantFormat, "test1") + fmt.Sprintf(wantFormat, "test2")
		if got := buf.String(); got != want {
			t.Errorf("Write() writes %v, want %v", got, want)
		}
	})
}

func TestNotifyingWriter_Write(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		writer := &bytes.Buffer{}
		notifyCh := make(chan struct{}, 2)
		w := &NotifyingWriter{
			Writer:   writer,
			NotifyCh: notifyCh,
		}

		if _, err := w.Write([]byte("test1")); err != nil {
			t.Errorf("Write() error = %v, wantErr %v", err, false)
		}
		if _, err := w.Write([]byte("test2")); err != nil {
			t.Errorf("Write() error = %v, wantErr %v", err, false)
		}
		if got, want := writer.String(), "test1test2"; got != want {
			t.Errorf("Write() writes %v, want %v", got, want)
		}
		if got, want := len(notifyCh), 2; got != want {
			t.Errorf("NotifyCh length %v, want %v", got, want)
		}
	})
}
