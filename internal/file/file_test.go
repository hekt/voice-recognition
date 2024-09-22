package file

import (
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewOpenCloseFileWriter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		path := "test.txt"
		flag := os.O_CREATE
		perm := os.FileMode(0o644)

		want := &OpenCloseFileWriter{path, flag, perm}
		got := NewOpenCloseFileWriter(path, flag, perm)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("NewFileWriter() = %v, want %v", got, want)
		}
	})
}

func TestOpenCloseFileWriter_Write(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()
		path := dir + "/test.txt"
		flag := os.O_APPEND | os.O_CREATE | os.O_WRONLY
		perm := os.FileMode(0o644)

		w := &OpenCloseFileWriter{path, flag, perm}
		p := []byte("test")

		got, err := w.Write(p)
		if err != nil {
			t.Errorf("FileWriter.Write() error = %v", err)
		}
		if got != len(p) {
			t.Errorf("FileWriter.Write() = %v, want %v", got, len(p))
		}

		file, err := os.Open(path)
		if err != nil {
			t.Errorf("failed to open file: %v", err)
		}
		defer file.Close()

		b, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read file: %v", err)
		}
		if diff := cmp.Diff(p, b); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("unexisting file with no O_CREATE flag", func(t *testing.T) {
		dir := t.TempDir()
		path := dir + "/test.txt"
		flag := os.O_APPEND
		perm := os.FileMode(0o644)

		w := &OpenCloseFileWriter{path, flag, perm}
		p := []byte("test")

		_, err := w.Write(p)
		if err == nil {
			t.Error("FileWriter.Write() error = nil, want an error")
		}
	})

	t.Run("with O_RONLY flag", func(t *testing.T) {
		dir := t.TempDir()
		path := dir + "/test.txt"
		flag := os.O_CREATE | os.O_RDONLY
		perm := os.FileMode(0o644)

		w := &OpenCloseFileWriter{path, flag, perm}
		p := []byte("test")

		_, err := w.Write(p)
		if err == nil {
			t.Error("FileWriter.Write() error = nil, want an error")
		}
	})
}
