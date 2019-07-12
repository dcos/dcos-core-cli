import six

from dcoscli.test.common import exec_command


def test_help():
    with open('dcoscli/data/help/quota.txt') as content:
        planned_stdout = six.b(content.read())
    returncode, stdout, stderr = exec_command(['dcos', 'quota', '--help'])
    assert returncode == 0
    assert stderr == b''
    assert stdout == planned_stdout
