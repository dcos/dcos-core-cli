from dcoscli.test.common import assert_command


def test_calico_version():
    with open('tests/data/calico/version.txt') as content:
        assert_command(['dcos', 'calico', 'version'],
                       stdout=content.read().encode('utf-8'))
