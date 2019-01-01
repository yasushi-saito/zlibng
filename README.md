Cgo wrapper for zlib-ng (https://github.com/zlib-ng/zlib-ng).
It provides the same interface as "compress/gzip".

Benchmark results:

CPU: Intel(R) Xeon(R) CPU E3-1505M v6 @ 3.00GHz

```
goos: linux
goarch: amd64
pkg: github.com/yasushi-saito/zlibng
BenchmarkInflateStandardGzip-8    	     100	  17955301 ns/op
BenchmarkInflateKlauspostGzip-8   	     100	  22067228 ns/op
BenchmarkInflateZlibNG-8          	     200	   6395802 ns/op
BenchmarkDeflateStandardGzip-8    	       1	11391486024 ns/op
BenchmarkDeflateKlauspostGzip-8   	       1	2916114323 ns/op
BenchmarkDeflateZlibNG-8          	       1	6181263938 ns/op
```
