#!/usr/bin/env python3

"""This script modifies the zlib-ng source directory for cgo.

It supports only amd64 currently.

"""

import glob
import os
import logging
from typing import List

def replace_in_place(path: str, from_str:str, to_str: str) -> None:
    lines = [] # type: List[str]
    found = False
    with open(path) as fd:
        for line in fd.readlines():
            if from_str in line:
                line = line.replace(from_str, to_str)
                found = True
            lines.append(line)
        if not found:
            return
    logging.info('rewrite %s', path)
    with open(path, 'w') as fd:
        for line in lines:
            fd.write(line)

def main() -> None:
    logging.basicConfig(level=logging.DEBUG)

    for path in glob.glob('arch/x86/*'):
        new_path = os.path.basename(path)
        logging.info('%s -> %s', path, new_path)
        os.rename(path, new_path)

    for path in glob.glob('*.h') + glob.glob('*.c'):
        replace_in_place(path, 'arch/x86/', './')

main()
