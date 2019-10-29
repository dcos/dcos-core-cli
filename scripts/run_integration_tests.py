#!/usr/bin/env python3

import os
import sys

import pytest

from dcoscli.test.common import dcos_tempdir, exec_command

os.environ["CLI_TEST_SSH_USER"] = "centos"
os.environ["CLI_TEST_MASTER_PROXY"] = "1"
os.environ["CLI_TEST_SSH_KEY_PATH"] = os.environ.get('DCOS_TEST_SSH_KEY_PATH')

code, _, _ = exec_command(['./launch_aws_cluster.sh'])
assert code == 0

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
