package testutil

import "io"

var _ io.Reader = (*IOReaderMock)(nil)

type IOReaderMock struct {
	ReadFunc func(p []byte) (n int, err error)
}

func (m *IOReaderMock) Read(p []byte) (n int, err error) {
	return m.ReadFunc(p)
}

var _ io.Writer = (*IOWriterMock)(nil)

type IOWriterMock struct {
	WriteFunc func(p []byte) (n int, err error)
}

func (m *IOWriterMock) Write(p []byte) (n int, err error) {
	return m.WriteFunc(p)
}
