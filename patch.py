#!/usr/bin/env python3.6

"""This script modifies the zlib-ng source directory for cgo.


- Move files under arch/x86 to the toplevel directory, since cgo doesn't support subdirectories.

- Add "zng_" prefix to non-static, hidden functions and variables. This is
  needed to allow mixing libz and zlibng in a single binary. Otherwise if you
  import zlibng, and also use libz elsewhere, you'll get symbol conflicts and
  worse random SEGVs.

  Ideally zlibng should be doing this already, but this problem is especially
  serious in this package since zlibng files are statically linked.

Currently, this script supports only Linux+AMD64.

"""

import shutil
import argparse
import glob
import os
import logging

def patch_file(src_path: str, dst_path: str):
    """Copy a file from src_path to dst_path while rewriting its contents."""
    patterns = [
        ('arch/x86/', 'arch-x86-'),
        ('crc_folding.h', 'arch-x86-crc_folding.h'),
        ('"./x86.h"', '"./arch-x86-x86.h"'),
        ('"x86.h"', '"./arch-x86-x86.h"'),
        ('crc_fold_', 'zng_crc_fold_'),
        ('x86_check_features', 'zng_x86_check_features'),
        ('fill_window_c', 'zng_fill_window_c'),
        ('_tr_init', '_zng_tr_init'),
        ('_tr_tally', '_zng_tr_tally'),
        ('_tr_flush', '_zng_tr_flush'),
        ('_tr_align', '_zng_tr_align'),
        ('_tr_stored_block', '_zng_tr_stored_block'),
        ('bi_windup', 'zng_bi_windup'),
        ('_length_code', '_zng_length_code'),
        ('_dist_code', '_zng_dist_code'),
        ('functable;', 'zng_functable;'),
        ('functable.', 'zng_functable.'),
        ('functable = ', 'zng_functable = '),
        ('inflate_fast;', 'zng_inflate_fast'),
        ('inflate_table;', 'zng_inflate_table'),
        ('ct_data_staic_ltree', 'zng_ct_data_staic_ltree'),
        ('z_verbose', 'zng_z_verbose'),
        ('z_error', 'zng_z_error'),
        ('z_errmsg', 'zng_z_errmsg'),
        ('deflate_copyright', 'zng_deflate_copyright'),
        ('inflate_copyright', 'zng_inflate_copyright'),
        ('zcalloc', 'zng_zcalloc'),
        ('zcfree', 'zng_zcfree'),
        ('zng_functable.h\"', 'functable.h"'),  # undo the include name change
    ]
    logging.info('%s -> %s', src_path, dst_path)
    with open(src_path) as in_fd, open(dst_path, 'w') as out_fd:
        for line in in_fd.readlines():
            for from_str, to_str in patterns:
                line = line.replace(from_str, to_str)
            out_fd.write(line)

def main() -> None:
    """Main entry point."""
    logging.basicConfig(level=logging.DEBUG)

    parser = argparse.ArgumentParser()
    parser.add_argument("src_dir", type=str, help='Directory that stores zlib-ng source files')
    args = parser.parse_args()

    shutil.copyfile(args.src_dir + "/LICENSE.md", "LICENSE.md")
    shutil.copyfile(args.src_dir + "/README.md", "README-zlibng.md")

    for src_path in (glob.glob(f'{args.src_dir}/arch/x86/*.c') +
                     glob.glob(f'{args.src_dir}/arch/x86/*.h')):
        dst_path = "arch-x86-" + os.path.basename(src_path)
        patch_file(src_path, dst_path)

    for src_path in glob.glob(f'{args.src_dir}/*.h') + glob.glob(f'{args.src_dir}/*.c'):
        if os.path.basename(src_path) in ['gzclose.c', 'gzlib.c',
                                          'gzread.c', 'gzwrite.c', 'gzguts.h']:
            continue
        patch_file(src_path, os.path.basename(src_path))

main()
