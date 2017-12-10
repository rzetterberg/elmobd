#!/usr/bin/env python3

"""Automation of common tasks (such as running multiple tests)."""

import argparse
import glob
import logging
import os
import sys
from subprocess import check_call, CalledProcessError

#-------------------------------------------------------------------------------
# Constants

SCRIPT_NAME = 'helper'

#-------------------------------------------------------------------------------
# Logger setup

logger = logging.getLogger(SCRIPT_NAME)
logger.setLevel(logging.DEBUG)

basic_formatter = logging.Formatter(
    '[%(levelname)s] %(message)s'
)

stream_handler = logging.StreamHandler()
stream_handler.setLevel(logging.INFO)

stream_handler.setFormatter(basic_formatter)
logger.addHandler(stream_handler)

#-------------------------------------------------------------------------------
# Utilities

class cd:
    """Context manager for changing the current working directory"""
    def __init__(self, new_path):
        self.new_path = os.path.expanduser(new_path)

    def __enter__(self):
        self.saved_path = os.getcwd()

        os.chdir(self.new_path)

    def __exit__(self, etype, value, traceback):
        os.chdir(self.saved_path)

def command(*args, **kwargs):
    opts = {
        'can_fail':       False,
        'capture_output': False
    }

    for key, default in opts.items():
        opts[key] = kwargs.get(key, default)

        if key in kwargs:
            del kwargs[key]

    f = call if opts['can_fail'] else check_call
    f = check_output if opts['capture_output'] else f

    try:
        result = f(*args, **kwargs)

        if opts['capture_output']:
            return str(result, encoding = 'utf-8')
        else:
            return result
    except CalledProcessError as e:
        MAX_ARGS = 10

        if len(e.cmd) > MAX_ARGS:
            cmd = ' '.join(e.cmd[0:MAX_ARGS])
            cmd = '%s ... (%d args truncated)' % (cmd, len(e.cmd) - MAX_ARGS)
        else:
            cmd = ' '.join(e.cmd)

        logger.error(
            'Command failed (returned %d): %s' % (e.returncode, cmd)
        )

        sys.exit(e.returncode)

#-------------------------------------------------------------------------------
# Main tasks

def check(args):
    logger.info('Checking quality of project')

    command(['go', 'fmt'])
    command(['go', 'tool', 'vet', '--all', '.'])
    command(['go', 'build', '-v', './...'])

def test(args):
    logger.info('Running unit tests')
    command(['go', 'test', './...'])

    example_files = glob.glob('./examples/**/*.go', recursive=True)

    for efile in example_files:
        logger.info('Running example %s' % efile)
        command(['go', 'run', efile])

#-------------------------------------------------------------------------------

if __name__ == '__main__':
    parser     = argparse.ArgumentParser(prog=SCRIPT_NAME)
    subparsers = parser.add_subparsers()

    parser.add_argument(
        '--log_level',
        type=str,
        choices=['error', 'info', 'debug'],
        help='Controls the log level, "info" is default'
    )

    check_parser = subparsers.add_parser(
        'check',
    )

    check_parser.set_defaults(func=check)

    test_parser = subparsers.add_parser(
        'test',
    )

    test_parser.set_defaults(func=test)

    args = parser.parse_args()

    if args.log_level == 'debug':
        stream_handler.setLevel(logging.DEBUG)
    elif args.log_level == 'error':
        stream_handler.setLevel(logging.ERROR)

    if hasattr(args, 'func'):
        args.func(args)
    else:
        logger.error('No command given to run')
        sys.exit(1)
