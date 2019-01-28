// +build cgo

package zlibng_test

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/grailbio/testutil/assert"
	"github.com/klauspost/compress/gzip"
	"github.com/vitessio/vitess/go/cgzip"
	"github.com/yasushi-saito/zlibng"
)

func TestDeflateHeader(t *testing.T) {
	out := bytes.Buffer{}
	zout, err := zlibng.NewWriter(&out)
	assert.NoError(t, err)

	now := time.Unix(time.Now().Unix(), 0)
	wantHeader := zlibng.GzipHeader{Comment: "hello", Name: "blah", Extra: []byte{3, 2, 1}, ModTime: now, OS: 11}
	assert.NoError(t, zout.SetHeader(wantHeader))
	data := []byte("testdata")
	n, err := zout.Write(data)
	assert.NoError(t, err)
	assert.EQ(t, n, len(data))
	assert.NoError(t, zout.Close())
	{
		zin, err := zlibng.NewReader(bytes.NewReader(out.Bytes()))
		assert.NoError(t, err)
		got := bytes.Buffer{}
		_, err = io.Copy(&got, zin)
		assert.NoError(t, err)
		assert.EQ(t, string(got.Bytes()), string(data))
		gotHeader, err := zin.Header()
		assert.NoError(t, err)
		assert.EQ(t, gotHeader, wantHeader)
	}
	{
		zin, err := gzip.NewReader(bytes.NewReader(out.Bytes()))
		assert.NoError(t, err)
		got := bytes.Buffer{}
		_, err = io.Copy(&got, zin)
		assert.NoError(t, err)
		assert.EQ(t, string(got.Bytes()), string(data))
	}
}

func BenchmarkInflateCGZip(b *testing.B) {
	benchmarkInflate(b, *testSmallPathFlag,
		func(in io.Reader) (io.Reader, io.Closer, error) {
			r, err := cgzip.NewReaderBuffer(in, 512<<10)
			return r, r, err
		})
}

func BenchmarkDeflateCGZip(b *testing.B) {
	benchmarkDeflate(b, *testPathFlag,
		func(out io.Writer) io.WriteCloser {
			w, err := cgzip.NewWriterLevel(out, 5)
			assert.NoError(b, err)
			return w
		})
}
