package zlibng

import (
	"errors"
	"fmt"
)

// defaultBufferSize is the default value of Opts.Buffer
const defaultBufferSize = 512 * 1024

// Format defines the compression format.
type Format int

const (
	// Gzip is the format defined in RFC1952
	Gzip Format = iota
	// Flate is the format defined in RFC1951
	Flate
)

// Opts define the options passed to NewReader and NewWriter.
type Opts struct {
	// Format specifies the compression format. Gzip is RFC1952, Flate is RFC 1951.
	Format Format
	// Buffer specifies the internal buffer size used during compression and decompression.
	// The default value is 512KiB.
	Buffer int
	// Level specifies the compression level, used only by the writer.
	// The default value of 0 means no compression, which is probably not what you want.
	// -1 is the default compression level. If you don't pass any Opts to NewWriter,
	// it will use -1 as the value.
	Level int
}

func getOpts(opts ...Opts) (Opts, error) {
	opt := Opts{Level: -1}
	switch len(opts) {
	case 0:
	case 1:
		opt = opts[0]
	default:
		return opt, errors.New("zlibng: at most one option can be specified")
	}
	if opt.Buffer <= 0 {
		opt.Buffer = defaultBufferSize
	}
	if opt.Format != Gzip && opt.Format != Flate {
		return opt, fmt.Errorf("zlibng: invalid format %v", opt.Format)
	}
	return opt, nil
}
