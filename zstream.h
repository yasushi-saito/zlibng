#ifndef ZSTREAM_H
#define ZSTREAM_H

// format is one of Gzip or Flate.
extern int zs_inflate_init(char* stream, int format);
extern int zs_inflate_reset(char* stream);
extern void zs_inflate_end(char* stream);
extern int zs_inflate(char* stream, void* in, int in_bytes, void* out,
                      int* out_bytes, int* consumed_input);

// format is one of Gzip or Flate.
extern int zs_deflate_init(char* stream, int format, int level);
extern int zs_deflate(char* stream, void* in, int in_bytes, void* out,
                      int* out_bytes);
extern int zs_deflate_end(char* stream, void* out, int* out_bytes);

extern int zs_get_errno();

#endif /* ZSTREAM_H */
