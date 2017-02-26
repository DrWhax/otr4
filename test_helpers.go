package otr4

import "io"

type fixedRandReader struct {
	data []byte
	at   int
}

func fixedRand(data []byte) io.Reader {
	return &fixedRandReader{data, 0}
}

func (r *fixedRandReader) Read(p []byte) (n int, err error) {
	if r.at < len(r.data) {
		n = copy(p, r.data[r.at:])
		r.at += fieldBytes
		return
	}
	return 0, io.ErrUnexpectedEOF
}
