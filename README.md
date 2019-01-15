Cgo wrapper for zlib-ng (https://github.com/zlib-ng/zlib-ng).
It provides the same interface as "compress/gzip" and "compress/flate".

It currently supports only {linux,darwin}/amd64.

Benchmark results:

CPU: Intel(R) Xeon(R) CPU E3-1505M v6 @ 3.00GHz

```
goos: linux
goarch: amd64
pkg: github.com/yasushi-saito/zlibng
BenchmarkInflateStandardGzip-8    	     100	  16481110 ns/op
BenchmarkInflateKlauspostGzip-8   	     100	  19946956 ns/op
BenchmarkInflateZlibNG-8          	     200	   5908574 ns/op
BenchmarkInflateCGZip-8           	     200	   7771587 ns/op
BenchmarkDeflateStandardGzip-8    	       1	10572484457 ns/op
BenchmarkDeflateKlauspostGzip-8   	       1	2665946548 ns/op
BenchmarkDeflateZlibNG-8          	       1	5886680851 ns/op
BenchmarkDeflateCGZip-8           	       1	5069861793 ns/op
```

# Installation

```
go get github.com/yasushi-saito/zlibng
```

# Incorporating upstream changes

```
cd /tmp/
git clone https://github.com/zlib-ng/zlib-ng.git
git clone git@github.com:yasushi-saito/zlibng.git
cd zlibng
./patch.py ../zlib-ng
git commit -a
```
