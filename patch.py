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
