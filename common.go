package zlibng

import (
	"errors"
	"time"
)

// Strategy values
const (
	// FilteredStrategy stands for C.Z_FILTERED
	FilteredStrategy = 1
	// HuffmanOnlyStrategy stands for  C.Z_HUFFMAN_ONLY
	HuffmanOnlyStrategy = 2
	// RLEStrategy stands for C.Z_RLE
	RLEStrategy = 3
	// FixedStrategy stands for C.Z_FIXED
	FixedStrategy = 4
	// DefaultStrategy stands for C.Z_DEFAULT_STRATEGY
	DefaultStrategy = 0
)

// DefaultBufferSize is the default value of Opts.Buffer
const DefaultBufferSize = 512 * 1024

const (
	// Gzip is the value of Opts.WindowBits to use FLATE format as defined in RFC1952
	Gzip = 16 + 15
	// Flate should is the value of Opts.WindowBits to use FLATE format as defined in RFC1951
	Flate = -15
)

// GzipHeader alters the contents the gzip header. It is stored in
// Opts.GzipHeader to control the contents of the header.
//
// TODO(saito) Support other fields if needed.
type GzipHeader struct {
	Comment string
	Extra   []byte
	ModTime time.Time
	Name    string
	// OS field, cf. RFC1952 Section 2.3. Default: 255
	OS byte
}

// Opts define the options passed to NewReader and NewWriter.
type Opts struct {
	// WindowBits specifies the windowBits arg for deflateInit and inflateInit. It
	// specifies the compression window size as well as the header format.  If
	// unset, Gzip is used.
	WindowBits int
	// Buffer specifies the internal buffer size used during compression and
	// decompression.  The default value is 512KiB.
	Buffer int
	// Level specifies the compression level, used only by the writer.
	// The default value of 0 means no compression, which is probably not what you want.
	// -1 is the default compression level. If you don't pass any Opts to NewWriter,
	// it will use -1 as the value.
	Level int

	// The following fields are not for general use. They are only for NewWriter,
	// and they are ignored by NewReader. If they are nonzero, they are passed
	// verbatim to deflateInit2. See the zlib doc (http://zlib.net/manual.html)
	// for more details.

	// MemLevel specifies the memLevel arg for deflateInit2. If unset, value of 8
	// is used.
	MemLevel int
	// Strategy specifies the strategy arg for deflateInit. If unset,
	// Z_DEFAULT_STRATEGY is used.
	Strategy int
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
		opt.Buffer = DefaultBufferSize
	}
	return opt, nil
}
