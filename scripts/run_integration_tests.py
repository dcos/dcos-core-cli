#!/usr/bin/env python3

import atexit
import os
import sys

import pytest

from dcoscli.test.common import dcos_tempdir, exec_command


@atexit.register
def clenup():
    print("Destroying cluster")
    exec_command(['./terraform', 'destroy', '-auto-approve', '-no-color'])


os.environ["CLI_TEST_SSH_USER"] = "centos"
os.environ["CLI_TEST_MASTER_PROXY"] = "1"
os.environ["CLI_TEST_SSH_KEY_PATH"] = os.environ.get('DCOS_TEST_SSH_KEY_PATH')

master_ip = os.environ['MASTER_PUBLIC_IP']

with dcos_tempdir():
    code, _, _ = exec_command(['dcos', 'cluster', 'setup', '--no-check', master_ip])
    assert code == 0

    code, _, _ = exec_command(['dcos', 'plugin', 'add', '-u', '../build/' + sys.platform + '/dcos-core-cli.zip'])
    assert code == 0

    os.chdir("../python/lib/dcoscli")

    retcode = pytest.main([
        '-vv',
        '-x',
        '--durations=10',
        '-p', 'no:cacheprovider',
        'tests/integrations'
    ])

sys.exit(retcode)
