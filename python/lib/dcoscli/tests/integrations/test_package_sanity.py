from dcoscli.test.common import exec_command


def test_install_certified_packages_cli():
    pkgs = [
        'cassandra',
        'kubernetes',
        'confluent-kafka',
    ]

    for pkg in pkgs:
        code, stdout, stderr = exec_command(
            ['dcos', 'package', 'install', '--cli', '--yes', pkg])
        assert "New command available: dcos " + pkg in stderr.decode()
        assert stdout == b''
        assert code == 0
