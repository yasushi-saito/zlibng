package zlibng_test

// Pre 2019-09-20
//
// BenchmarkInflateCGZip-56            	   10000	    482956 ns/op
// BenchmarkDeflateCGZip-56            	       1	5514222083 ns/op
// BenchmarkInflateStandardGzip-56     	   10000	    604818 ns/op
// BenchmarkInflateKlauspostGzip-56    	   10000	    620658 ns/op
// BenchmarkInflateZlibNG-56           	   10000	    430711 ns/op
// BenchmarkDeflateStandardGzip-56     	       1	7770433615 ns/op
// BenchmarkDeflateKlauspostGzip-56    	       1	3908032173 ns/op
// BenchmarkDeflateZlibNG-56           	       1	4603557318 ns/op

// 2019-09-20
//
// BenchmarkInflateCGZip-56            	   10000	    481241 ns/op
// BenchmarkDeflateCGZip-56            	       1	5506438605 ns/op
// BenchmarkInflateStandardGzip-56     	   10000	    604598 ns/op
// BenchmarkInflateKlauspostGzip-56    	   10000	    552915 ns/op
// BenchmarkInflateZlibNG-56           	   10000	    406962 ns/op
// BenchmarkDeflateStandardGzip-56     	       1	7570097805 ns/op
// BenchmarkDeflateKlauspostGzip-56    	       1	3830474557 ns/op
// BenchmarkDeflateZlibNG-56           	       1	4725882232 ns/op

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/grailbio/testutil/assert"
	kgzip "github.com/klauspost/compress/gzip"
	"github.com/yasushi-saito/zlibng"
)

func testInflate(t *testing.T, r *rand.Rand, windowBits int, src []byte, want []byte) {
	zin, err := zlibng.NewReader(bytes.NewReader(src), zlibng.Opts{WindowBits: windowBits})
	assert.NoError(t, err)

	var (
		got []byte
		buf = make([]byte, 8192)
	)

	noProgress := 0
	iter := 0
	for {
		iter++
		n := rand.Intn(8192)
		n2, err := zin.Read(buf[:n])
		if n2 > 0 {
			got = append(got, buf[:n2]...)
			noProgress = 0
		} else if err == io.EOF {
			break
		} else {
			noProgress++
			assert.LT(t, noProgress, 2, "iter=%d", iter)
		}
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
	}
	assert.NoError(t, zin.Close())
	if !bytes.Equal(got, want) {
		t.Fatal("fail")
	}
}

func TestInflateGzipEmpty(t *testing.T) {
	compressed := bytes.Buffer{}
	w := gzip.NewWriter(&compressed)
	assert.NoError(t, w.Close())
	r := rand.New(rand.NewSource(0))
	testInflate(t, r, zlibng.Gzip, compressed.Bytes(), nil)
}

func TestInflateFlateEmpty(t *testing.T) {
	compressed := bytes.Buffer{}
	w, err := flate.NewWriter(&compressed, flate.DefaultCompression)
	assert.NoError(t, err)
	assert.NoError(t, w.Close())
	r := rand.New(rand.NewSource(0))
	testInflate(t, r, zlibng.Flate, compressed.Bytes(), nil)
}

func TestInflateGzipSmall(t *testing.T) {
	data := []byte("Blah")
	compressed := bytes.Buffer{}
	gz := gzip.NewWriter(&compressed)
	_, err := gz.Write(data)
	assert.NoError(t, err)
	assert.NoError(t, gz.Close())
	r := rand.New(rand.NewSource(0))
	testInflate(t, r, zlibng.Gzip, compressed.Bytes(), data)
}

func TestInflateFlateSmall(t *testing.T) {
	data := []byte("Blah")
	compressed := bytes.Buffer{}
	w, err := flate.NewWriter(&compressed, flate.DefaultCompression)
	assert.NoError(t, err)
	_, err = w.Write(data)
	assert.NoError(t, err)
	assert.NoError(t, w.Close())
	r := rand.New(rand.NewSource(0))
	testInflate(t, r, zlibng.Flate, compressed.Bytes(), data)
}

func TestDeflateFlateEmpty(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	testDeflate(t, r, zlibng.Flate, nil)
}

func TestDeflateFlateSmall(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	testDeflate(t, r, zlibng.Flate, []byte("Blah"))
}

func TestInflateRandom(t *testing.T) {
	for iter := 0; iter < 20; iter++ {
		i := iter
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			r := rand.New(rand.NewSource(int64(i)))
			n := r.Intn(16<<20) + 1
			uncompressed := make([]byte, n)
			_, err := r.Read(uncompressed)
			assert.NoError(t, err)

			compressed := bytes.Buffer{}
			gz := gzip.NewWriter(&compressed)
			_, err = gz.Write(uncompressed)
			assert.NoError(t, err)
			assert.NoError(t, gz.Close())
			testInflate(t, r, zlibng.Gzip, compressed.Bytes(), uncompressed)
		})
	}
}

// Test packed gzip
func TestInflateRandomPacked(t *testing.T) {
	for iter := 0; iter < 20; iter++ {
		i := iter
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			r := rand.New(rand.NewSource(int64(i)))
			compressed := bytes.Buffer{}
			uncompressed := bytes.Buffer{}

			for j := 0; j < 10; j++ {
				n := r.Intn(2<<20) + 1
				buf := make([]byte, n)
				_, err := r.Read(buf)
				assert.NoError(t, err)
				uncompressed.Write(buf)

				gz := gzip.NewWriter(&compressed)
				_, err = gz.Write(buf)
				assert.NoError(t, err)
				assert.NoError(t, gz.Close())
			}
			testInflate(t, r, zlibng.Gzip, compressed.Bytes(), uncompressed.Bytes())
		})
	}
}

func testDeflate(t *testing.T, r *rand.Rand, windowBits int, src []byte) {
	orgSrc := src
	out := bytes.Buffer{}
	zout, err := zlibng.NewWriter(&out, zlibng.Opts{WindowBits: windowBits, Level: -1})
	assert.NoError(t, err)

	for len(src) > 0 {
		n := r.Intn(8192)
		if n > len(src) {
			n = len(src)
		}
		n2, err := zout.Write(src[:n])
		assert.NoError(t, err)
		assert.EQ(t, n, n2)
		src = src[n:]
	}
	assert.NoError(t, zout.Close())

	got := bytes.Buffer{}
	var zin io.Reader
	if windowBits == zlibng.Gzip {
		zin, err = gzip.NewReader(bytes.NewReader(out.Bytes()))
		assert.NoError(t, err)
	} else {
		zin = flate.NewReader(bytes.NewReader(out.Bytes()))
		assert.NoError(t, err)
	}
	n, err := io.Copy(&got, zin)
	assert.NoError(t, err)
	assert.EQ(t, int(n), len(orgSrc))
	if !bytes.Equal(got.Bytes(), orgSrc) {
		t.Fatal("fail")
	}
}

func TestDeflateRandom(t *testing.T) {
	for iter := 0; iter < 20; iter++ {
		i := iter
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			r := rand.New(rand.NewSource(int64(i)))
			n := r.Intn(16 << 20)
			data := make([]byte, n)
			_, err := r.Read(data)
			assert.NoError(t, err)
			testDeflate(t, r, zlibng.Gzip, data)
		})
	}
}

var (
	testSmallPathFlag = flag.String("small-path",
		"/scratch-nvme/cache_tmp/0.intervals.tsv", "Plain-text file used for small tests")
	testPathFlag = flag.String("path",
		"/scratch-nvme/cache_tmp/cpg_windows1000_chr_only.txt", "Plain-text file used for in tests and benchmarks")
	testGZPathFlag = flag.String("gz-path",
		"/scratch-nvme/cache_tmp/cpg_windows1000_chr_only.txt",
		"Gzipped file used in tests and benchmarks")
	runManualTestsFlag = flag.Bool("run-manual-tests",
		false, "Run large tests using files outside the repo")
)

func TestDeflateLarge(t *testing.T) {
	if !*runManualTestsFlag {
		t.Skip("--run-manual-tests not set")
	}
	testDeflateLarge(t, *testGZPathFlag)
}

func testDeflateLarge(t *testing.T, gzPath string) {
	type reader struct {
		in             *os.File
		r              io.Reader
		buf, remaining []byte
	}
	const bufSize = 1 << 20
	var (
		err    error
		r0, r1 reader
		r      = rand.New(rand.NewSource(0))
	)
	open := func(r *reader) {
		r.in, err = os.Open(*testGZPathFlag)
		assert.NoError(t, err)
		r.buf = make([]byte, bufSize)
	}
	read := func(r *reader, want int) ([]byte, bool) {
		buf := make([]byte, want)
		remaining := buf
		for {
			n := len(remaining)
			if n > len(r.remaining) {
				n = len(r.remaining)
			}
			copy(remaining, r.remaining)
			remaining = r.remaining[n:]
			r.remaining = r.remaining[n:]
			if len(remaining) == 0 {
				break
			}
			got, err := r.r.Read(r.buf)
			if got == 0 {
				assert.EQ(t, err, io.EOF)
				break
			}
			if err != nil {
				assert.EQ(t, err, io.EOF)
			}
			r.buf = r.buf[got:]
			r.remaining = r.buf
		}
		if len(remaining) == want {
			return nil, false
		}
		return buf[0 : len(buf)-len(remaining)], true
	}

	open(&r0)
	r0.r, err = gzip.NewReader(r0.in)
	assert.NoError(t, err)
	open(&r1)
	r1.r, err = zlibng.NewReader(r1.in)
	assert.NoError(t, err)

	total := 0
	last := 0
	for {
		nMax := r.Intn(bufSize)
		buf0, ok0 := read(&r0, nMax)
		buf1, ok1 := read(&r1, nMax)
		if !bytes.Equal(buf0, buf1) {
			t.Fatalf("want %d gotn0 %d gotn1 %d", nMax, len(buf0), len(buf1))
		}
		assert.EQ(t, ok0, ok1)
		if !ok0 {
			break
		}
		total += len(buf0)
		if total-last > 1<<30 {
			log.Printf("read %d bytes", total)
			last = total
		}
	}
	assert.NoError(t, r0.in.Close())
	assert.NoError(t, r1.in.Close())
}

func benchmarkInflate(
	b *testing.B,
	path string,
	inflateFactory func(in io.Reader) (io.Reader, io.Closer, error)) {
	b.StopTimer()
	tmp, err := ioutil.TempDir("", "")
	assert.NoError(b, err)
	defer os.RemoveAll(tmp)

	in, err := os.Open(path)
	assert.NoError(b, err)
	dstPath := filepath.Join(tmp, "tmp.gz")
	out, err := os.Create(dstPath)
	assert.NoError(b, err)
	outgz := gzip.NewWriter(out)
	wantByte, err := io.Copy(outgz, in)
	assert.NoError(b, err)
	assert.NoError(b, outgz.Close())
	assert.NoError(b, in.Close())
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in, err = os.Open(dstPath)
		assert.NoError(b, err)
		inflator, closer, err := inflateFactory(in)
		assert.NoError(b, err)
		n, err := io.Copy(ioutil.Discard, inflator)
		assert.NoError(b, err)
		assert.EQ(b, n, wantByte)
		if closer != nil {
			assert.NoError(b, closer.Close())
		}
	}
}

func BenchmarkInflateStandardGzip(b *testing.B) {
	benchmarkInflate(b, *testSmallPathFlag,
		func(in io.Reader) (io.Reader, io.Closer, error) {
			r, err := gzip.NewReader(in)
			return r, nil, err
		})
}

func BenchmarkInflateKlauspostGzip(b *testing.B) {
	benchmarkInflate(b, *testSmallPathFlag,
		func(in io.Reader) (io.Reader, io.Closer, error) {
			r, err := kgzip.NewReader(in)
			return r, nil, err
		})
}

func BenchmarkInflateZlibNG(b *testing.B) {
	benchmarkInflate(b, *testSmallPathFlag,
		func(in io.Reader) (io.Reader, io.Closer, error) {
			r, err := zlibng.NewReader(in, zlibng.Opts{Buffer: 512 << 10})
			return r, r, err
		})
}

type discardingWriter struct {
	n int64
}

func (w *discardingWriter) Write(data []byte) (int, error) {
	w.n += int64(len(data))
	return len(data), nil
}

func benchmarkDeflate(
	b *testing.B,
	path string,
	deflateFactory func(out io.Writer) io.WriteCloser) {
	var w discardingWriter
	for i := 0; i < b.N; i++ {
		w = discardingWriter{}
		deflator := deflateFactory(&w)
		in, err := os.Open(path)
		assert.NoError(b, err)
		_, err = io.Copy(deflator, bufio.NewReaderSize(in, 1<<20))
		assert.NoError(b, err)
		assert.NoError(b, deflator.Close())
		assert.NoError(b, in.Close())
	}
}

func BenchmarkDeflateStandardGzip(b *testing.B) {
	benchmarkDeflate(b, *testPathFlag,
		func(out io.Writer) io.WriteCloser {
			w, err := gzip.NewWriterLevel(out, 5)
			assert.NoError(b, err)
			return w
		})
}

func BenchmarkDeflateKlauspostGzip(b *testing.B) {
	benchmarkDeflate(b, *testPathFlag,
		func(out io.Writer) io.WriteCloser {
			w, err := kgzip.NewWriterLevel(out, 5)
			assert.NoError(b, err)
			return w
		})
}

func BenchmarkDeflateZlibNG(b *testing.B) {
	benchmarkDeflate(b, *testPathFlag,
		func(out io.Writer) io.WriteCloser {
			w, err := zlibng.NewWriter(out, zlibng.Opts{Level: 5, Buffer: 512 << 10})
			assert.NoError(b, err)
			return w
		})
}
