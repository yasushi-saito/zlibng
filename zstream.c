#include "./zstream.h"
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "./zlib-ng.h"

int zs_inflate_init(char* stream, int window_bits, struct zng_gz_header_s* h,
                    int* get_header_status) {
  zng_stream* zs = (zng_stream*)stream;
  memset(zs, 0, sizeof(*zs));
  int ec = zng_inflateInit2(zs, window_bits);
  if (ec != 0) {
    return ec;
  }
  *get_header_status = zng_inflateGetHeader(zs, h);
  return 0;
}

int zs_inflate_end(char* stream) { return zng_inflateEnd((zng_stream*)stream); }

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

int zs_deflate_init(char* stream, int level, int window_bits, int mem_level,
                    int strategy) {
  zng_stream* zs = (zng_stream*)stream;
  memset(zs, 0, sizeof(*zs));
  return zng_deflateInit2(zs, level, Z_DEFLATED, window_bits, mem_level,
                          strategy);
}

int zs_deflate(char* stream, void* in, int in_bytes, void* out, int* out_bytes,
               int* consumed_input) {
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
  *consumed_input = (zs->avail_in == 0);
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
