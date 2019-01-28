// +build cgo,amd64

package zlibng

/*

#cgo linux CFLAGS: -march=ivybridge -std=c99 -Wall -D_LARGEFILE64_SOURCE=1 -DHAVE_HIDDEN -DHAVE_INTERNAL -DHAVE_BUILTIN_CTZL -DMEDIUM_STRATEGY -DX86_64 -DX86_NOCHECK_SSE2 -DUNALIGNED_OK -DUNROLL_LESS -DX86_CPUID -DX86_SSE2_FILL_WINDOW -DX86_SSE4_2_CRC_HASH -DX86_SSE4_2_CRC_INTRIN -DX86_PCLMULQDQ_CRC -DX86_QUICK_STRATEGY -I.

#cgo darwin CFLAGS: -march=ivybridge -std=c99 -Wall -DHAVE_HIDDEN -DHAVE_INTERNAL -DHAVE_BUILTIN_CTZL -DMEDIUM_STRATEGY -DX86_64 -DX86_NOCHECK_SSE2 -DUNALIGNED_OK -DUNROLL_LESS -DX86_CPUID -DX86_SSE2_FILL_WINDOW -DX86_SSE4_2_CRC_HASH -DX86_SSE4_2_CRC_INTRIN -DX86_PCLMULQDQ_CRC -DX86_QUICK_STRATEGY -I.

#include <errno.h>
#include <stdlib.h>
#include "./zlib-ng.h"
#include "./zstream.h"

*/
import "C"

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

type zstream [unsafe.Sizeof(C.zng_stream{})]C.char

// Reader is a gzip/zlib/flate reader. It implements io.ReadCloser.  Calling
// Close is optional, though strongly recommended.  NewReader() also installs a
// GC finalizer that closes the Reader, in case the application forgets to call
// Close.
type Reader struct {
	in          io.Reader
	inConsumed  bool    // true if zstream has finished consuming the current input buffer.
	inEOF       bool    // true if in reaches io.EOF
	hasGzHeader bool    // true if gzHeader was successfully set.
	zs          zstream // underlying zlib implementation.
	gzHeader    C.zng_gz_header
	inBuf       []byte
	err         error
}

func freeReader(z *Reader) {
	_ = C.zs_inflate_end(&z.zs[0])
	freeGzHeaderFields(&z.gzHeader)
}

// NewReader creates a gzip/flate reader. There can be at most one options arg.
func NewReader(in io.Reader, opts ...Opts) (*Reader, error) {
	opt, err := getOpts(opts...)
	if err != nil {
		return nil, err
	}
	if opt.WindowBits == 0 {
		opt.WindowBits = 32 + 15 // autodetect gzip/zlib
	}
	z := &Reader{
		in:         in,
		inBuf:      make([]byte, opt.Buffer),
		inConsumed: true, // force in.Read
	}
	ec := C.zs_inflate_init(&z.zs[0], C.int(opt.WindowBits))
	if ec != 0 {
		return nil, zlibReturnCodeToError(ec)
	}
	const maxStringLen = 256
	z.gzHeader.comment = (*C.uchar)(C.malloc(maxStringLen))
	z.gzHeader.comm_max = maxStringLen
	z.gzHeader.name = (*C.uchar)(C.malloc(maxStringLen))
	z.gzHeader.name_max = maxStringLen
	z.gzHeader.extra = (*C.uchar)(C.malloc(maxStringLen))
	z.gzHeader.extra_max = maxStringLen
	ec = C.zs_inflate_get_header(&z.zs[0], &z.gzHeader)
	if ec == 0 {
		z.hasGzHeader = true
	}
	runtime.SetFinalizer(z, freeReader)
	return z, nil
}

// Header reads the gzip header contents. If the file is a multi-gzip
// concatenation, this function returns the contents of the current archive.
//
// REQUIRES: Opts.GetGzipHeader=true when the reader was created.
func (z *Reader) Header() (GzipHeader, error) {
	if !z.hasGzHeader {
		return GzipHeader{}, errors.New("zlibng.header: Header not supported")
	}
	h := GzipHeader{}
	if z.gzHeader.comment != nil {
		h.Comment = C.GoString((*C.char)(unsafe.Pointer(z.gzHeader.comment)))
	}
	if z.gzHeader.extra != nil {
		h.Extra = C.GoBytes(unsafe.Pointer(z.gzHeader.extra), C.int(z.gzHeader.extra_len))
	}
	if z.gzHeader.name != nil {
		h.Name = C.GoString((*C.char)(unsafe.Pointer(z.gzHeader.name)))
	}
	if z.gzHeader.time > 0 {
		h.ModTime = time.Unix(int64(z.gzHeader.time), 0)
	}
	h.OS = byte(z.gzHeader.os)
	return h, nil
}

// Close implements io.Closer.
func (z *Reader) Close() error {
	runtime.SetFinalizer(z, nil)
	ec := C.zs_inflate_end(&z.zs[0])
	freeGzHeaderFields(&z.gzHeader)
	if z.err == io.EOF {
		return zlibReturnCodeToError(ec)
	}
	return z.err
}

// Read implements io.Reader.
func (z *Reader) Read(out []byte) (int, error) {
	var orgOut = out
	for z.err == nil && len(out) > 0 {
		var (
			outLen     = C.int(len(out))
			ret        C.int
			inConsumed C.int
		)
		if !z.inConsumed {
			ret = C.zs_inflate(&z.zs[0], nil, 0, unsafe.Pointer(&out[0]), &outLen, &inConsumed)
		} else {
			if z.inEOF {
				z.err = io.EOF
				break
			}
			n, err := z.in.Read(z.inBuf)
			if err != nil {
				if err != io.EOF {
					z.err = err
					break
				}
				z.inEOF = true
				// fall through
			}
			if n == 0 {
				if !z.inEOF {
					panic(z)
				}
				z.err = io.EOF
				break
			}
			ret = C.zs_inflate(&z.zs[0], unsafe.Pointer(&z.inBuf[0]), C.int(n), unsafe.Pointer(&out[0]), &outLen, &inConsumed)
		}
		z.inConsumed = (inConsumed != 0)
		if ret != C.Z_STREAM_END && ret != C.Z_OK {
			z.err = zlibReturnCodeToError(ret)
			break
		}
		nOut := len(out) - int(outLen)
		out = out[nOut:]
		if ret == C.Z_STREAM_END {
			ret = C.zs_inflate_reset(&z.zs[0])
			if ret != C.Z_OK {
				z.err = zlibReturnCodeToError(ret)
			}
			break
		}
	}
	return len(orgOut) - len(out), z.err
}

// Writer is the gzip/flate writer. It implements io.WriterCloser.
type Writer struct {
	out      io.Writer
	zs       zstream // underlying zlib implementation.
	gzHeader C.zng_gz_header
	outBuf   []byte
}

// NewWriter creates a gzip/flate writer. There can be at most one options arg.
// If opts is empty, NewWriter will use Opts{Format:Gzip,Level:-1}.
func NewWriter(w io.Writer, opts ...Opts) (*Writer, error) {
	opt, err := getOpts(opts...)
	if err != nil {
		return nil, err
	}
	z := &Writer{
		out:    w,
		outBuf: make([]byte, opt.Buffer),
	}
	if opt.WindowBits == 0 {
		opt.WindowBits = Gzip
	}
	if opt.MemLevel == 0 {
		opt.MemLevel = 8
	}
	if opt.Strategy == 0 {
		opt.Strategy = DefaultStrategy
	}
	ec := C.zs_deflate_init(&z.zs[0], C.int(opt.Level),
		C.int(opt.WindowBits), C.int(opt.MemLevel), C.int(opt.Strategy))
	if ec != 0 {
		return nil, zlibReturnCodeToError(ec)
	}
	return z, nil
}

// SetHeader sets the Gzip header contents.
//
// REQUIRES: No Write nor Close has been called yet.
// REQUIRES: The archive format is Gzip.
func (z *Writer) SetHeader(h GzipHeader) error {
	// comment, extra, and name should be null unless the value is
	// actually set to something.
	if len(h.Comment) > 0 {
		z.gzHeader.comment = (*C.uchar)(unsafe.Pointer(C.CString(h.Comment)))
	}
	if len(h.Extra) > 0 {
		z.gzHeader.extra = (*C.uchar)(C.CBytes(h.Extra))
		z.gzHeader.extra_len = C.uint(len(h.Extra))
	}
	if len(h.Name) > 0 {
		z.gzHeader.name = (*C.uchar)(unsafe.Pointer(C.CString(h.Name)))
	}
	if h.ModTime.After(time.Unix(0, 0)) {
		z.gzHeader.time = C.ulong(h.ModTime.Unix())
	}
	if h.OS != 0 {
		z.gzHeader.os = C.int(h.OS)
	}
	ec := C.zs_deflate_set_header(&z.zs[0], &z.gzHeader)
	return zlibReturnCodeToError(ec)
}

// Flush writes the data to the output.
func (z *Writer) flush(data []byte) error {
	n, err := z.out.Write(data)
	if err != nil {
		return err
	}
	if n < len(data) { // shouldn't happen in practice
		return fmt.Errorf("zlib: n=%d, outLen=%d", n, len(data))
	}
	return nil
}

func freeGzHeaderFields(h *C.zng_gz_header) {
	if h.comment != nil {
		C.free(unsafe.Pointer(h.comment))
	}
	if h.extra != nil {
		C.free(unsafe.Pointer(h.extra))
	}
	if h.name != nil {
		C.free(unsafe.Pointer(h.name))
	}
}

// Close implements io.Closer
func (z *Writer) Close() error {
	defer freeGzHeaderFields(&z.gzHeader)
	for {
		outLen := C.int(len(z.outBuf))
		ret := C.zs_deflate_end(&z.zs[0], unsafe.Pointer(&z.outBuf[0]), &outLen)
		if ret != 0 && ret != C.Z_STREAM_END {
			return zlibReturnCodeToError(ret)
		}
		nOut := len(z.outBuf) - int(outLen)
		if err := z.flush(z.outBuf[:nOut]); err != nil {
			return err
		}
		if ret == C.Z_STREAM_END {
			return nil
		}
	}
}

// Write implements io.Writer.
func (z *Writer) Write(in []byte) (int, error) {
	if len(in) == 0 {
		return 0, nil
	}
	var (
		outLen     = C.int(len(z.outBuf))
		inConsumed C.int
	)
	ret := C.zs_deflate(&z.zs[0], unsafe.Pointer(&in[0]), C.int(len(in)),
		unsafe.Pointer(&z.outBuf[0]), &outLen, &inConsumed)
	if ret != 0 {
		return 0, zlibReturnCodeToError(ret)
	}
	nOut := len(z.outBuf) - int(outLen)
	if err := z.flush(z.outBuf[:nOut]); err != nil {
		return 0, err
	}
	if inConsumed != 0 {
		return len(in), nil
	}
	for {
		outLen = C.int(len(z.outBuf))
		ret = C.zs_deflate(&z.zs[0], nil, 0, unsafe.Pointer(&z.outBuf[0]), &outLen, &inConsumed)
		if ret != 0 {
			return 0, zlibReturnCodeToError(ret)
		}
		nOut := len(z.outBuf) - int(outLen)
		if err := z.flush(z.outBuf[:nOut]); err != nil {
			return 0, err
		}
		if inConsumed != 0 { // outbuf didn't fillup, i.e., the input was fully consumed.
			break
		}
	}
	return len(in), nil
}

var zlibErrors = map[C.int]error{
	C.Z_OK:            nil,
	C.Z_STREAM_END:    io.EOF,
	C.Z_ERRNO:         nil, // handled separately
	C.Z_STREAM_ERROR:  errors.New("Zlib: stream error"),
	C.Z_DATA_ERROR:    errors.New("Zlib: data error"),
	C.Z_MEM_ERROR:     errors.New("Zlib: mem error"),
	C.Z_BUF_ERROR:     errors.New("Zlib: buf error"),
	C.Z_VERSION_ERROR: errors.New("Zlib: version error"),
}

func zlibReturnCodeToError(r C.int) error {
	if r == 0 {
		return nil
	}
	if r == C.Z_ERRNO {
		return unix.Errno(C.zs_get_errno())
	}
	if err, ok := zlibErrors[r]; ok {
		return err
	}
	return fmt.Errorf("Zlib: unknown error %d", r)
}
