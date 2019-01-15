// +build !cgo !amd64

package zlibng

import (
	"io"

	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/gzip"
)

// NewReader creates a gzip/flate writer. There can be at most one options arg.
func NewReader(in io.Reader, opts ...Opts) (io.ReadCloser, error) {
	opt, err := getOpts(opts...)
	if err != nil {
		return nil, err
	}
	if opt.Format == Gzip {
		return gzip.NewReader(in)
	}
	return flate.NewReader(in), nil
}

// NewWriter creates a gzip/flate writer. There can be at most one options arg.
// If opts is empty, NewWriter will use Opts{Format:Gzip,Level:-1}.
func NewWriter(w io.Writer, opts ...Opts) (io.WriteCloser, error) {
	opt, err := getOpts(opts...)
	if err != nil {
		return nil, err
	}
	if opt.Format == Gzip {
		return gzip.NewWriterLevel(w, opt.Level)
	}
	return flate.NewWriter(w, opt.Level)
}
