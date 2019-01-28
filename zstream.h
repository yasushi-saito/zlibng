#ifndef ZSTREAM_H
#define ZSTREAM_H

// format is one of Gzip or Flate.
extern int zs_inflate_init(char* stream, int format);
extern int zs_inflate_reset(char* stream);
extern int zs_inflate_end(char* stream);
extern int zs_inflate(char* stream, void* in, int in_bytes, void* out,
                      int* out_bytes, int* consumed_input);

// format is one of Gzip or Flate.
extern int zs_deflate_init(char* stream, int level, int window_bits,
                           int mem_level, int strategy);
struct zng_gz_header_s;
extern int zs_deflate_set_header(char* stream, struct zng_gz_header_s* h);
extern int zs_inflate_get_header(char* stream, struct zng_gz_header_s* h);
extern int zs_deflate(char* stream, void* in, int in_bytes, void* out,
                      int* out_bytes);
extern int zs_deflate_end(char* stream, void* out, int* out_bytes);

extern int zs_get_errno();

#endif /* ZSTREAM_H */
