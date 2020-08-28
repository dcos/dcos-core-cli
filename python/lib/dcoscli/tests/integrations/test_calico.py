import sys

import pytest

from dcoscli.test.common import assert_command


def test_calico_version():
    with open('tests/data/calico/version.txt') as content:
        assert_command(['dcos', 'calico', 'version'],
                       stdout=content.read().encode('utf-8'))


@pytest.mark.skipif(sys.platform == 'win32',
                    reason='Different error message on windows')
def test_calico_flag_parsing():
    assert_command(['dcos', 'calico', 'apply', '-f', 'not-existing-file.yml'],
                   returncode=1,
                   stderr=b'Failed to execute command: '
                          b'open not-existing-file.yml: '
                          b'no such file or directory\n'
                          b'Error: exit status 1\n')
