
Cgo wrapper for zlib-ng (https://github.com/zlib-ng/zlib-ng).
It provides the same interface as "compress/gzip".

Benchmark results:

CPU: Intel(R) Xeon(R) CPU E3-1505M v6 @ 3.00GHz

```
goos: linux
goarch: amd64
pkg: github.com/yasushi-saito/zlibng
BenchmarkInflateStandardGzip-8    	     100	  16347473 ns/op
BenchmarkInflateKlauspostGzip-8   	     100	  19608899 ns/op
BenchmarkInflateZlibNG-8          	     300	   5964480 ns/op
BenchmarkDeflateStandardGzip-8    	       1	10833457488 ns/op
BenchmarkDeflateKlauspostGzip-8   	       1	2715833640 ns/op
BenchmarkDeflateZlibNG-8          	       1	6687730628 ns/op
```
