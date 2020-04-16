from dcoscli.test.common import exec_command

pytestmark = pytest.mark.skip("all tests still WIP")


def test_install_certified_packages_cli():
    pkgs = [
        'cassandra',
        'kubernetes',
        'confluent-kafka',
    ]

    for pkg in pkgs:
        code, stdout, stderr = exec_command(
            ['dcos', 'package', 'install', '--cli', '--yes', pkg])
        assert "New commands available: " + pkg in stderr.decode()
        assert stdout == b''
        assert code == 0
