package testutil

import "io"

var _ io.Reader = (*ChannelReader)(nil)

type ChannelReader struct {
	BufCh chan []byte
	EOFCh chan struct{}
}

func NewChannelReader() *ChannelReader {
	bufCh := make(chan []byte)
	eofCh := make(chan struct{})

	return &ChannelReader{
		BufCh: bufCh,
		EOFCh: eofCh,
	}
}

func (r *ChannelReader) Read(p []byte) (int, error) {
	select {
	case <-r.EOFCh:
		return 0, io.EOF
	case buf := <-r.BufCh:
		n := copy(p, buf)
		if n < len(buf) {
			r.BufCh <- buf[n:]
		}
		return n, nil
	}
}

func (r *ChannelReader) Close() {
	close(r.BufCh)
	close(r.EOFCh)
}
