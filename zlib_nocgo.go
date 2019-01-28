// +build !cgo !amd64

package zlibng

import (
	"errors"
	"io"

	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/gzip"
)

type reader struct {
	io.ReadCloser
}

// NewReader creates a gzip/flate writer. There can be at most one options arg.
func NewReader(in io.Reader, opts ...Opts) (reader, error) {
	opt, err := getOpts(opts...)
	if err != nil {
		return reader{}, err
	}
	if opt.WindowBits == Flate {
		z := flate.NewReader(in)
		return reader{z}, nil
	}
	z, err := gzip.NewReader(in)
	return reader{z}, err
}

func (r reader) Header() (GzipHeader, error) {
	return GzipHeader{}, errors.New("zlibng.Header: Not supported")
}

type writer struct{ io.WriteCloser }

// NewWriter creates a gzip/flate writer. There can be at most one options arg.
// If opts is empty, NewWriter will use Opts{Format:Gzip,Level:-1}.
func NewWriter(w io.Writer, opts ...Opts) (writer, error) {
	opt, err := getOpts(opts...)
	if err != nil {
		return writer{}, err
	}
	if opt.WindowBits == Flate {
		z, err := flate.NewWriter(w, opt.Level)
		return writer{z}, err
	}
	z, err := gzip.NewWriterLevel(w, opt.Level)
	return writer{z}, err
}

func (w writer) SetHeader(GzipHeader) error {
	return errors.New("zlibng.SetHeader: Not supported")
}
