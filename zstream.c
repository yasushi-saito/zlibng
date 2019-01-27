#include "./zstream.h"
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "./zlib-ng.h"

int zs_inflate_init(char* stream, int format) {
  int flags = 32 + 15;  // autodetect gzip or zlib
  if (format != 0) {
    flags = -15;  // flate
  }
  zng_stream* zs = (zng_stream*)stream;
  memset(zs, 0, sizeof(*zs));
  // 32 makes it autodetect gzip or flate files
  return zng_inflateInit2(zs, flags);
}

void zs_inflate_end(char* stream) { zng_inflateEnd((zng_stream*)stream); }

int zs_inflate_reset(char* stream) {
  zng_stream* zs = (zng_stream*)stream;
  return zng_inflateReset(zs);
}

int zs_get_errno() { return errno; }

int zs_inflate(char* stream, void* in, int in_bytes, void* out, int* out_bytes,
               int* consumed_input) {
  zng_stream* zs = (zng_stream*)stream;
  if (in_bytes > 0) {
    if (zs->avail_in != 0) {
      abort();
    }
    zs->avail_in = in_bytes;
    zs->next_in = in;
  } else {
    if (zs->avail_in == 0) {
      abort();
    }
  }
  zs->next_out = out;
  zs->avail_out = *out_bytes;
  int ret = zng_inflate((zng_stream*)stream, Z_NO_FLUSH);
  if (ret == Z_OK || ret == Z_STREAM_END) {
    *out_bytes = zs->avail_out;
  }
  *consumed_input = (zs->avail_in == 0);
  return ret;
}

int zs_deflate_init(char* stream, int format, int level, int window_bits,
                    int mem_level, int strategy) {
  if (window_bits == 0) {
    window_bits = 16 + 15;  // gzip
    if (format != 0) {
      window_bits = -15;
    }
  }
  if (mem_level == 0) {
    mem_level = 8;
  }
  if (strategy == 0) {
    strategy = Z_DEFAULT_STRATEGY;
  }
  zng_stream* zs = (zng_stream*)stream;
  memset(zs, 0, sizeof(*zs));
  return zng_deflateInit2(zs, level, Z_DEFLATED, window_bits, mem_level,
                          strategy);
}

int zs_deflate(char* stream, void* in, int in_bytes, void* out,
               int* out_bytes) {
  zng_stream* zs = (zng_stream*)stream;
  if (in_bytes > 0) {
    if (zs->avail_in != 0) {  // has buffered input
      abort();
    }
    zs->avail_in = in_bytes;
    zs->next_in = in;
  } else if (zs->avail_in == 0) {
    abort();
  }
  zs->next_out = out;
  zs->avail_out = *out_bytes;
  int ret = zng_deflate(zs, Z_NO_FLUSH);
  *out_bytes = zs->avail_out;
  return ret;
}

int zs_deflate_end(char* stream, void* out, int* out_bytes) {
  zng_stream* zs = (zng_stream*)stream;
  if (zs->avail_in != 0) {
    abort();
  }
  zs->next_out = out;
  zs->avail_out = *out_bytes;
  int ret = zng_deflate(zs, Z_FINISH);
  *out_bytes = zs->avail_out;
  if (ret != Z_OK) {
    zng_deflateEnd(zs);
  }
  return ret;
}

int zs_deflate_set_header(char* stream, zng_gz_header* h) {
  return zng_deflateSetHeader((zng_stream*)stream, h);
}

int zs_inflate_get_header(char* stream, zng_gz_header* h) {
  return zng_inflateGetHeader((zng_stream*)stream, h);
}
