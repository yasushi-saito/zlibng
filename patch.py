#!/usr/bin/env python3.6

"""This script modifies the zlib-ng source directory for cgo.

It supports only amd64 currently.

"""

import shutil
import argparse
import glob
import os
import logging

def patch_file(src_path: str, dst_path: str):
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

def copy_file(src_path: str, dst_path=""):
    if dst_path == "":
        dst_path = os.path.basename(src_path)
    shutil.copyfile(src_path, dst_path)

def main() -> None:
    logging.basicConfig(level=logging.DEBUG)

    p = argparse.ArgumentParser()
    p.add_argument("src_dir", default=os.environ["HOME"] + "/github/zlib-ng")
    args = p.parse_args()

    shutil.copyfile(args.src_dir + "/LICENSE.md", "LICENSE.md")
    shutil.copyfile(args.src_dir + "/README.md", "README-zlibng.md")

    for src_path in glob.glob(f'{args.src_dir}/arch/x86/*.c') + glob.glob(f'{args.src_dir}/arch/x86/*.h'):
        dst_path = "arch-x86-" + os.path.basename(src_path)
        patch_file(src_path, dst_path)

    for src_path in glob.glob(f'{args.src_dir}/*.h') + glob.glob(f'{args.src_dir}/*.c'):
        skip = False
        for ignore in ['gzclose', 'gzlib', 'gzread', 'gzwrite', 'gzguts']:
            if ignore in src_path:
                skip = True
                break
        if skip:
            continue
        patch_file(src_path, os.path.basename(src_path))

main()
