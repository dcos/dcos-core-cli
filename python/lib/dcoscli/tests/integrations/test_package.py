import contextlib
import json
import os
import subprocess
import sys

import pytest
import six

from dcos import config, constants, errors, http, subcommand, util

from dcoscli.test.common import (assert_command, assert_lines, base64_to_dict,
                                 delete_zk_nodes, exec_command,
                                 file_json, update_config)
from dcoscli.test.marathon import watch_all_deployments
from dcoscli.test.package import (BOOTSTRAP_REGISTRY_REPO,
                                  setup_universe_server,
                                  teardown_universe_server, UNIVERSE_REPO,
                                  UNIVERSE_TEST_REPOS)
from dcoscli.test.service import get_services, service_shutdown

pytestmark = pytest.mark.skip("all tests still WIP")


@pytest.fixture
def env():
    r = os.environ.copy()
    r.update({constants.PATH_ENV: os.environ[constants.PATH_ENV]})

    return r


def setup_module(module):
    setup_universe_server()


def teardown_module(module):
    services = get_services()
    for framework in services:
        if framework['name'] == 'chronos':
            service_shutdown(framework['id'])

    teardown_universe_server()


@pytest.fixture(scope="module")
def zk_znode(request):
    request.addfinalizer(delete_zk_nodes)
    return request


def test_repo_list():
    repo_list = bytes(
        (
            "Universe: {0}\n"
            "Bootstrap Registry: {1}\n"
            "helloworld-universe: {helloworld-universe}\n"
        ).format(UNIVERSE_REPO, BOOTSTRAP_REGISTRY_REPO,
                 **UNIVERSE_TEST_REPOS),
        'utf-8'
    )

    assert_command(['dcos', 'package', 'repo', 'list'], stdout=repo_list)

    # test again, but override the dcos_url with a cosmos_url config
    dcos_url = config.get_config_val("core.dcos_url")
    with update_config('package.cosmos_url', dcos_url):
        assert_command(['dcos', 'package', 'repo', 'list'], stdout=repo_list)


def test_repo_add_and_remove():
    repo17 = "https://universe.mesosphere.com/repo-1.7"
    repo_list = bytes(
        (
            "Universe: {1}\n"
            "1.7-universe: {0}\n"
            "Bootstrap Registry: {2}\n"
            "helloworld-universe: {helloworld-universe}\n"
        ).format(repo17,  UNIVERSE_REPO, BOOTSTRAP_REGISTRY_REPO,
                 **UNIVERSE_TEST_REPOS),
        'utf-8'
    )

    args = ["1.7-universe", repo17, '--index=1']
    _repo_add(args, repo_list)

    repo_list = bytes(
        (
            "Universe: {0}\n"
            "Bootstrap Registry: {1}\n"
            "helloworld-universe: {helloworld-universe}\n"
        ).format(UNIVERSE_REPO, BOOTSTRAP_REGISTRY_REPO,
                 **UNIVERSE_TEST_REPOS),
        'utf-8'
    )
    _repo_remove(['1.7-universe'], repo_list)


def test_repo_remove_multi_and_empty():
    repos = ['Universe', 'Bootstrap Registry']
    repos.extend(UNIVERSE_TEST_REPOS.keys())

    repos_remove_cmd = ['dcos', 'package', 'repo', 'remove']
    repos_remove_cmd.extend(repos)
    assert_command(repos_remove_cmd)

    returncode, stdout, stderr = exec_command(
        ['dcos', 'package', 'repo', 'list'])
    assert returncode == 0
    assert stdout == b''
    assert stderr == b''

    # Add back test repos
    for name, url in UNIVERSE_TEST_REPOS.items():
        assert_command(['dcos', 'package', 'repo', 'add', name, url])

    assert_command(
        ['dcos', 'package', 'repo', 'add', 'Universe', UNIVERSE_REPO])

    assert_command(['dcos', 'package', 'repo', 'add', 'Bootstrap Registry',
                    BOOTSTRAP_REGISTRY_REPO])


def test_describe_nonexistent():
    stderr = b"Error: Package [xyzzy] not found\n"
    assert_command(['dcos', 'package', 'describe', 'xyzzy'],
                   stderr=stderr,
                   returncode=1)


def test_describe_nonexistent_version():
    stderr = b'Error: Version [a.b.c] of package [helloworld] not found\n'
    assert_command(['dcos', 'package', 'describe', 'helloworld',
                    '--package-version=a.b.c'],
                   stderr=stderr,
                   returncode=1)


def test_describe_options():
    stdout = file_json(
        'tests/data/package/json/test_describe_app_options.json')
    stdout = json.loads(stdout.decode('utf-8'))
    expected_labels = stdout.pop("labels", None)

    with util.temptext(b'{"name": "hallo", "port": 80}') as options:
        returncode, stdout_, stderr = exec_command(
            ['dcos', 'package', 'describe', '--app', '--options',
             options[1], 'helloworld'])

    stdout_ = json.loads(stdout_.decode('utf-8'))
    actual_labels = stdout_.pop("labels", None)

    for label, value in expected_labels.items():
        if label in ["DCOS_PACKAGE_OPTIONS"]:
            # We covert the metadata into a dictionary
            # so that failures in equality are more descriptive
            assert base64_to_dict(value) == \
                base64_to_dict(actual_labels.get(label))
        else:
            assert value == actual_labels.get(label)

    assert stdout == stdout_
    assert stderr == b''
    assert returncode == 0


@pytest.mark.parametrize("command_to_run,expected_output_file", [
    ('helloworld --app', 'test_describe_app_helloworld.json'),
    ('helloworld --config', 'test_describe_helloworld_config.json'),
    ('helloworld --app --render', 'test_describe_helloworld_app_render.json'),
    ('helloworld --app --render --app-id=helloworld',
     'test_describe_helloworld_app_id_render.json'),
    ('helloworld --app --cli', 'test_describe_app_cli.json'),
    ('helloworld --package-versions',
     'test_describe_helloworld_versions.json'),
])
def test_describe(command_to_run, expected_output_file):
    stdout = file_json('tests/data/package/json/' + expected_output_file)
    expected_stdout = json.loads(stdout.decode('utf-8'))

    returncode, stdout, stderr = exec_command(
        ['dcos', 'package', 'describe'] + command_to_run.split(' '),
    )

    assert returncode == 0
    assert stderr == b''

    actual_stdout = json.loads(stdout.decode('utf-8'))
    assert expected_stdout == actual_stdout


@pytest.mark.parametrize("command_to_run,expected_output_file", [
    ('helloworld --cli', 'test_describe_cli_helloworld.json'),
    ('helloworld --package-version=0.1.0', 'test_describe_helloworld.json'),
])
def test_describe_cli(command_to_run, expected_output_file):
    stdout = file_json('tests/data/package/json/' + expected_output_file)
    expected_stdout = json.loads(stdout.decode('utf-8'))

    returncode, stdout, stderr = exec_command(
        ['dcos', 'package', 'describe'] + command_to_run.split(' '),
        )

    assert returncode == 0
    assert stderr == b''

    actual_stdout = json.loads(stdout.decode('utf-8'))
    diff = set(expected_stdout) - set(actual_stdout)
    assert diff == set()


def test_bad_install():
    stderr = (
        b'This is a Community service. '
        b'Community services are not tested '
        b'for production environments. '
        b'There may be bugs, incomplete features, '
        b'incorrect documentation, or other discrepancies.\n'
        b'By Deploying, you agree to the Terms '
        b'and Conditions https://mesosphere.com/'
        b'catalog-terms-conditions/#community-services\n'
        b'A sample pre-installation message\n'
        b'Installing Marathon app for package [helloworld] version '
        b'[0.1.0]\n'
        b'Error: Options JSON failed validation\n'
    )
    with util.temptext(b'{"nom": "hallo"}') as options:
        args = ['--options='+options[1], '--yes']
        _install_bad_helloworld(args=args, stderr=stderr)


def test_bad_install_helloworld_msg():
    terms_conditions = (
        b'This is a Community service. '
        b'Community services are not tested '
        b'for production environments. '
        b'There may be bugs, incomplete features, '
        b'incorrect documentation, or other discrepancies.\n'
        b'By Deploying, you agree to the Terms '
        b'and Conditions https://mesosphere.com/'
        b'catalog-terms-conditions/#community-services\n'
        b'A sample pre-installation message\n'
        b'Installing Marathon app for package [helloworld] version '
        b'[0.1.0]\n'
    )

    stderr = (
        terms_conditions +
        b'Installing CLI subcommand for package [helloworld] '
        b'version [0.1.0]\n'
        b'New commands available: http\n'
        b'A sample post-installation message\n'
    )

    with util.temptext(b'{"name": "/foo"}') as foo, \
            util.temptext(b'{"name": "/foo/bar"}') as foobar:

        _install_helloworld(['--yes', '--options='+foo[1]],
                            stderr=stderr)

        stderr = terms_conditions + b'Error: Object is not valid\n'

        _install_helloworld(['--yes', '--options='+foobar[1]],
                            stderr=stderr,
                            returncode=1)
        _uninstall_helloworld()


@pytest.mark.skipif(sys.platform == 'win32', reason='DCOS_OSS-5624')
def test_uninstall_cli_only_when_no_apps_remain():
    with util.temptext(b'{"name": "/hello1"}') as opts_hello1, \
            util.temptext(b'{"name": "/hello2"}') as opts_hello2:
        stderr = (
            b'This is a Community service. '
            b'Community services are not tested '
            b'for production environments. '
            b'There may be bugs, incomplete features, '
            b'incorrect documentation, or other discrepancies.\n'
            b'By Deploying, you agree to the Terms '
            b'and Conditions https://mesosphere.com/'
            b'catalog-terms-conditions/#community-services\n'
            b'A sample pre-installation message\n'
            b'Installing Marathon app for package [helloworld] version '
            b'[0.1.0]\n'
            b'Installing CLI subcommand for package [helloworld] '
            b'version [0.1.0]\n'
            b'New commands available: http\n'
            b'A sample post-installation message\n'
        )

        uninstall_stderr = (
            b'Uninstalled package [helloworld] version [0.1.0]\n'
        )
        with _package(name='helloworld',
                      args=['--yes', '--options='+opts_hello1[1]],
                      stderr=stderr,
                      uninstall_app_id='/hello1',
                      uninstall_stderr=uninstall_stderr):

            with _package(name='helloworld',
                          args=['--yes', '--options='+opts_hello2[1]],
                          stderr=stderr,
                          uninstall_app_id='/hello2',
                          uninstall_stderr=uninstall_stderr):

                subcommand.command_executables('http')

            # helloworld subcommand should still be there as there is the
            # /hello1 app installed
            subcommand.command_executables('http')

        # helloworld subcommand should have been removed
        with pytest.raises(errors.DCOSException) as exc_info:
            subcommand.command_executables('helloworld')

        assert str(exc_info.value) == "'helloworld' is not a dcos command."


@pytest.mark.parametrize("command_to_run,expected_error", [
    ('helloworld --yes --options=asdf.json', b"Error: couldn't find options file 'asdf.json'\n"),
    ('helloworld --app --cli', b'Error: --app and --cli are mutually exclusive\n'),
    ('cassandra --package-version=a.b.c', b'Error: Version [a.b.c] of package [cassandra] not found\n'),
    ('missing-package', b'Error: Package [missing-package] not found\n'),
])
def test_install_error(command_to_run, expected_error):
    returncode, stdout, stderr = exec_command(
        ['dcos', 'package', 'install'] + command_to_run.split(' '),
    )

    assert returncode == 1
    assert stdout == b''
    assert stderr == expected_error


def test_install_specific_version():
    stderr = (
        b'This is a Community service. '
        b'Community services are not tested '
        b'for production environments. '
        b'There may be bugs, incomplete features, '
        b'incorrect documentation, or other discrepancies.\n'
        b'By Deploying, you agree to the Terms and Conditions https://'
        b'mesosphere.com/catalog-terms-conditions/#community-services\n'
        b'A sample pre-installation message\n'
        b'Installing Marathon app for package [helloworld] version [0.1.0]\n'
        b'Installing CLI subcommand '
        b'for package [helloworld] version [0.1.0]\n'
        b'New commands available: http\n'
        b'A sample post-installation message\n'
    )

    uninstall_stderr = b'Uninstalled package [helloworld] version [0.1.0]\n'

    with _package(name='helloworld',
                  args=[
                      '--yes',
                      '--package-version=0.1.0'
                  ],
                  stderr=stderr,
                  uninstall_stderr=uninstall_stderr):

        returncode, stdout, stderr = exec_command(
            ['dcos', 'package', 'list', 'helloworld', '--json'])
        assert returncode == 0
        assert stderr == b''
        assert json.loads(stdout.decode('utf-8'))[0]['version'] == "0.1.0"


def test_install_noninteractive():
    expected_stderr = (
        b'This is a Community service. '
        b'Community services are not tested '
        b'for production environments. '
        b'There may be bugs, incomplete features, '
        b'incorrect documentation, or other discrepancies.\n'
        b'By Deploying, you agree to the Terms and Conditions https://'
        b'mesosphere.com/catalog-terms-conditions/#community-services\n'
        b'This DC/OS Service is currently in preview.\n'
        b"Error: couldn't get confirmation\n"
    )
    expected_stdout = b'Continue installing? [yes/no] '
    returncode, stdout, stderr = exec_command(
        ['dcos', 'package', 'install', 'hello-world'],
        timeout=30,
        stdin=subprocess.DEVNULL)

    assert returncode == 1
    assert stderr == expected_stderr
    assert stdout == expected_stdout


def test_package_metadata():
    _install_helloworld()

    # test marathon labels
    expected_source = bytes(
        UNIVERSE_TEST_REPOS['helloworld-universe'],
        'utf-8'
    )

    expected_labels = {
        'DCOS_PACKAGE_NAME': b'helloworld',
        'DCOS_PACKAGE_VERSION': b'0.1.0',
        'DCOS_PACKAGE_SOURCE': expected_source,
    }

    app_labels = _get_app_labels('helloworld')
    for label, value in expected_labels.items():
        assert value == six.b(app_labels.get(label))

    # test local package.json
    package = file_json(
        'tests/data/package/json/test_package_metadata.json')
    package = json.loads(package.decode("UTF-8"))

    cmd = subcommand.InstalledSubcommand("dcos-http")

    # test local package.json
    assert cmd.package_json()['marathon'] == package['marathon']

    # uninstall helloworld
    _uninstall_helloworld()
    # TODO(janisz): Remove after DCOS_OSS-5619
    assert_command(['dcos', 'plugin', 'remove', 'dcos-http'])


def test_uninstall_missing():
    stderr = 'Error: Package [chronos] is not installed\n'
    _uninstall_chronos(returncode=1, stderr=stderr)

    stderr = 'Error: Package [chronos] is not installed\n'
    _uninstall_chronos(
        args=['--app-id=chronos-1'],
        returncode=1,
        stderr=stderr)


@pytest.mark.skip(reason="DCOS_OSS-5619")
def test_uninstall_subcommand():
    _install_helloworld()
    _uninstall_helloworld()
    _list(args=['--json'], stdout=b'[]\n')


@pytest.mark.skip(reason="DCOS_OSS-5619")
def test_uninstall_cli():
    _install_helloworld()
    _uninstall_cli_helloworld()

    stdout_json = {
        "apps": [
            "/helloworld"
        ],
        "command": {
            "name": "helloworld"
        },
        "description": "Example DCOS application package",
        "framework": False,
        "maintainer": "support@mesosphere.io",
        "name": "helloworld",
        "packagingVersion": "3.0",
        "postInstallNotes": "A sample post-installation message",
        "preInstallNotes": "A sample pre-installation message",
        "releaseVersion": 0,
        "selected": False,
        "tags": [
            "mesosphere",
            "example",
            "subcommand"
        ],
        "version": "0.1.0",
        "website": "https://github.com/mesosphere/dcos-helloworld"
    }

    returncode_, stdout_, stderr_ = exec_command(
        ['dcos', 'package', 'list', '--json'])
    assert stderr_ == b''
    assert returncode_ == 0
    output = json.loads(stdout_.decode('utf-8'))[0]
    assert output == stdout_json
    _uninstall_helloworld()


@pytest.mark.skip(reason=("Cosmos issue, see "
                          "https://jira.mesosphere.com/browse/DCOS_OSS-5529"))
def test_uninstall_multiple_apps():
    stderr = (
        b'This is a Community service. '
        b'Community services are not tested '
        b'for production environments. '
        b'There may be bugs, incomplete features, '
        b'incorrect documentation, or other discrepancies.\n'
        b'By Deploying, you agree to the Terms '
        b'and Conditions https://mesosphere.com/'
        b'catalog-terms-conditions/#community-services\n'
        b'A sample pre-installation message\n'
        b'Installing Marathon app for package [helloworld] version '
        b'[0.1.0]\n'
        b'A sample post-installation message\n'
    )

    with util.temptext(b'{"name": "/helloworld-1"}') as hello1, \
            util.temptext(b'{"name": "/helloworld-2"}') as hello2:

        _install_helloworld(
            ['--yes', '--options='+hello1[1], '--app'],
            stderr=stderr)

        _install_helloworld(
            ['--yes', '--options='+hello2[1], '--app'],
            stderr=stderr)

        stderr = (b"Multiple apps named [helloworld] are installed: "
                  b"[/helloworld-1, /helloworld-2].\n"
                  b"Please use --app-id to specify the ID of the app "
                  b"to uninstall, or use --all to uninstall all apps.\n")

        _uninstall_helloworld(stderr=stderr, returncode=1, uninstalled=b'')

        _uninstall_helloworld(args=['--all'])


@pytest.mark.parametrize("args", [
    '--json',
    'xyzzy --json',
    '--app-id=/xyzzy --json',
    '--json ceci-nest-pas-une-package',
    '--json --app-id=/' + 'ceci-nest-pas-une-package',
])
def test_list_empty(args):
    _list(args=args.split(' '), stdout=b'[]\n')


def test_list():
    with _helloworld():
        expected_output = file_json(
            'tests/data/package/json/test_list_helloworld.json', 4)
        _list(args=['--json'], stdout=expected_output)
        _list(args=['--json', 'helloworld'], stdout=expected_output)
        _list(args=['--json', '--app-id=/helloworld'], stdout=expected_output)


def test_list_table():
    with _helloworld():
        assert_lines(['dcos', 'package', 'list'], 2)


def test_install_yes():
    with open('tests/data/package/assume_yes.txt') as yes_file:
        _install_helloworld(
            args=[],
            stdin=yes_file,
            stdout=b'Continue installing? [yes/no] ',
            stderr=(
                b'This is a Community service. '
                b'Community services are not tested '
                b'for production environments. '
                b'There may be bugs, incomplete features, '
                b'incorrect documentation, or other discrepancies.\n'
                b'By Deploying, you agree to the Terms '
                b'and Conditions https://mesosphere.com/'
                b'catalog-terms-conditions/#community-services\n'
                b'A sample pre-installation message\n'
                b'Installing Marathon app for package [helloworld] version '
                b'[0.1.0]\n'
                b'Installing CLI subcommand for package [helloworld] '
                b'version [0.1.0]\n'
                b'New commands available: http\n'
                b'A sample post-installation message\n'
            )
        )
        _uninstall_helloworld()


def test_install_no():
    with open('tests/data/package/assume_no.txt') as no_file:
        _install_helloworld(
            args=[],
            stdin=no_file,
            stdout=b'Continue installing? [yes/no] ',
            stderr=(
                b'This is a Community service. '
                b'Community services are not tested '
                b'for production environments. '
                b'There may be bugs, incomplete features, '
                b'incorrect documentation, or other discrepancies.\n'
                b'By Deploying, you agree to the Terms '
                b'and Conditions https://mesosphere.com/'
                b'catalog-terms-conditions/#community-services\n'
                b'A sample pre-installation message\n'
                b"Error: couldn't get confirmation\n"
            ),
            returncode=1
        )


def test_list_cli():
    _install_helloworld()
    stdout = file_json(
        'tests/data/package/json/test_list_helloworld.json', 4)
    _list(args=['--json'], stdout=stdout)
    _uninstall_helloworld()

    stderr = (
        b'This is a Community service. '
        b'Community services are not tested '
        b'for production environments. '
        b'There may be bugs, incomplete features, '
        b'incorrect documentation, or other discrepancies.\n'
        b'By Deploying, you agree to the Terms '
        b'and Conditions https://mesosphere.com/'
        b'catalog-terms-conditions/#community-services\n'
        b'Installing CLI subcommand for package [helloworld] '
        b'version [0.1.0]\n'
        b'New commands available: http\n'
    )
    _install_helloworld(args=['--cli', '--yes'], stderr=stderr)

    stdout = file_json(
        'tests/data/package/json/test_list_helloworld_cli.json', 4)
    _list(args=['--json'], stdout=stdout)

    _uninstall_cli_helloworld()


def test_list_cli_only(env):
    helloworld_path = 'tests/data/package/json/test_list_helloworld_cli.json'
    helloworld_json = file_json(helloworld_path, 4)

    with _helloworld_cli(), \
            update_config('package.cosmos_url', 'http://nohost', env):
        assert_command(
            cmd=['dcos', 'package', 'list', '--json', '--cli'],
            stdout=helloworld_json)

        assert_command(
            cmd=['dcos', 'package', 'list', '--json', '--cli',
                 '--app-id=/helloworld'],
            stdout=helloworld_json)

        assert_command(
            cmd=['dcos', 'package', 'list', '--json', '--cli', 'helloworld'],
            stdout=helloworld_json)


def test_uninstall_multiple_frameworknames():
    retcode, _, _ = exec_command([
        'dcos', 'package', 'install', 'helloworld', '--app', '--yes',
        '--options=tests/data/package/helloworld-1.json'])
    assert retcode == 0

    retcode, _, _ = exec_command([
        'dcos', 'package', 'install', 'helloworld', '--app', '--yes',
        '--options=tests/data/package/helloworld-2.json'])
    assert retcode == 0

    watch_all_deployments()

    retcode, _, _ = exec_command([
        'dcos', 'package', 'uninstall',
        'helloworld', '--yes', '--app-id=hello-1'])
    assert retcode == 0

    retcode, _, _ = exec_command([
        'dcos', 'package', 'uninstall',
        'helloworld', '--yes', '--app-id=hello-2'])
    assert retcode == 0

    watch_all_deployments()


def test_search():
    returncode, stdout, stderr = exec_command(
        ['dcos', 'package', 'search', 'cron', '--json'])

    assert returncode == 0
    assert b'chronos' in stdout
    assert stderr == b''

    returncode, stdout, stderr = exec_command(
        ['dcos', 'package', 'search', 'xyzzy', '--json'])

    assert returncode == 0
    assert b'"packages": []' in stdout
    assert stderr == b''

    returncode, stdout, stderr = exec_command(
        ['dcos', 'package', 'search', 'xyzzy'])

    assert returncode == 1
    assert b'' == stdout
    assert b'no packages found' in stderr

    returncode, stdout, stderr = exec_command(
        ['dcos', 'package', 'search', '--json'])

    registries = json.loads(stdout.decode('utf-8'))
    # assert the number of packages is gte the number at the time
    # this test was written
    assert len(registries['packages']) >= 5

    assert returncode == 0
    assert stderr == b''


def test_search_table():
    returncode, stdout, stderr = exec_command(
        ['dcos', 'package', 'search'])

    assert returncode == 0
    assert b'chronos' in stdout
    assert len(stdout.decode('utf-8').split('\n')) > 5
    assert stderr == b''


def test_search_ends_with_wildcard():
    returncode, stdout, stderr = exec_command(
        ['dcos', 'package', 'search', 'c*', '--json'])

    assert returncode == 0
    assert b'chronos' in stdout
    assert b'cassandra' in stdout
    assert stderr == b''

    registries = json.loads(stdout.decode('utf-8'))
    # cosmos matches wildcards in name/description/tags
    # so will find more results (3 instead of 2)
    assert len(registries['packages']) >= 2


def test_search_start_with_wildcard():
    assert_command(['dcos', 'package', 'repo', 'remove', 'Universe'])
    assert_command(['dcos', 'package', 'repo', 'remove',
                    'Bootstrap Registry'])

    returncode, stdout, stderr = exec_command(
        ['dcos', 'package', 'search', '*world', '--json'])

    assert returncode == 0
    assert stderr == b''

    registries = json.loads(stdout.decode('utf-8'))
    assert len(registries['packages']) == 1
    assert registries['packages'][0]['name'] == 'helloworld'

    assert_command(
        ['dcos', 'package', 'repo', 'add', 'Universe', UNIVERSE_REPO])
    assert_command(['dcos', 'package', 'repo', 'add', 'Bootstrap Registry',
                    BOOTSTRAP_REGISTRY_REPO])


def test_search_middle_with_wildcard():
    assert_command(['dcos', 'package', 'repo', 'remove', 'Universe'])
    assert_command(['dcos', 'package', 'repo', 'remove',
                    'Bootstrap Registry'])

    returncode, stdout, stderr = exec_command(
        ['dcos', 'package', 'search', 'hellow*d', '--json'])

    assert returncode == 0
    assert stderr == b''

    registries = json.loads(stdout.decode('utf-8'))
    assert len(registries['packages']) == 1
    assert registries['packages'][0]['name'] == 'helloworld'

    assert_command(
        ['dcos', 'package', 'repo', 'add', 'Universe', UNIVERSE_REPO])
    assert_command(['dcos', 'package', 'repo', 'add', 'Bootstrap Registry',
                    BOOTSTRAP_REGISTRY_REPO])


def _get_app_labels(app_id):
    returncode, stdout, stderr = exec_command(
        ['dcos', 'marathon', 'app', 'show', app_id])

    assert returncode == 0
    assert stderr == b''

    app_json = json.loads(stdout.decode('utf-8'))
    return app_json.get('labels')


def _install_helloworld(
        args=None,
        stderr=(
            b'This is a Community service. '
            b'Community services are not tested '
            b'for production environments. '
            b'There may be bugs, incomplete features, '
            b'incorrect documentation, or other discrepancies.\n'
            b'By Deploying, you agree to the Terms '
            b'and Conditions https://mesosphere.com/'
            b'catalog-terms-conditions/#community-services\n'
            b'A sample pre-installation message\n'
            b'Installing Marathon app for package [helloworld] '
            b'version [0.1.0]\n'
            b'Installing CLI subcommand for package [helloworld] '
            b'version [0.1.0]\n'
            b'New commands available: http\n'
            b'A sample post-installation message\n'
        ),
        stdout=b'',
        returncode=0,
        stdin=None):
    if args is None:
        args = ['--yes']
    assert_command(
        ['dcos', 'package', 'install', 'helloworld'] + args,
        stdout=stdout,
        returncode=returncode,
        stdin=stdin,
        stderr=stderr)


def _uninstall_helloworld(
        args=[],
        stdout=b'',
        stderr=b'',
        returncode=0,
        uninstalled=b'Uninstalled package [helloworld] version [0.1.0]\n'):
    assert_command(['dcos', 'package', 'uninstall', 'helloworld',
                    '--yes'] + args,
                   stdout=stdout,
                   stderr=uninstalled+stderr,
                   returncode=returncode)

    watch_all_deployments()


def _uninstall_cli_helloworld(
        stdout=b'',
        stderr=b'',
        returncode=0):
    # TODO(janisz): Remove after DCOS_OSS-5619
    assert_command(['dcos', 'plugin', 'remove', 'dcos-http'],
                   stdout=stdout,
                   stderr=stderr,
                   returncode=returncode)
    # TODO(janisz): Uncomment after DCOS_OSS-5619
    # assert_command(['dcos', 'package', 'uninstall', 'helloworld',
    #                 '--cli'] + args,
    #                stdout=stdout,
    #                stderr=stderr,
    #                returncode=returncode)


def _uninstall_chronos(args=[], returncode=0, stdout=b'', stderr=''):
    result_returncode, result_stdout, result_stderr = exec_command(
        ['dcos', 'package', 'uninstall', 'chronos', '--yes'] + args)

    assert result_returncode == returncode
    assert result_stdout == stdout
    assert result_stderr.decode('utf-8').startswith(stderr)


def _install_bad_helloworld(
        args=['--yes'],
        stderr=(
            b'This is a Community service. '
            b'Community services are not tested '
            b'for production environments. '
            b'There may be bugs, incomplete features, '
            b'incorrect documentation, or other discrepancies.\n'
            b'By Deploying, you agree to the Terms '
            b'and Conditions https://mesosphere.com/'
            b'catalog-terms-conditions/#community-services\n'
            b'A sample pre-installation message\n'
            b'Installing Marathon app for package ['
            b'helloworld] version [0.1.0]\n'
        ),
        stdout=b''):
    cmd = ['dcos', 'package', 'install', 'helloworld'] + args
    returncode_, stdout_, stderr_ = exec_command(cmd)
    assert stdout_ == stdout
    assert stderr_ == stderr
    assert returncode_ == 1


def _list(args, stdout):
    assert_command(['dcos', 'package', 'list'] + args, stdout=stdout)


HELLOWORLD_CLI_STDOUT = (
    b'Installing CLI subcommand for package [helloworld] '
    b'version [0.1.0]\n'
    b'New commands available: http\n'
)


def _helloworld():
    stderr = (
        b'This is a Community service. '
        b'Community services are not tested '
        b'for production environments. '
        b'There may be bugs, incomplete features, '
        b'incorrect documentation, or other discrepancies.\n'
        b'By Deploying, you agree to the Terms '
        b'and Conditions https://mesosphere.com/'
        b'catalog-terms-conditions/#community-services\n'
        b'A sample pre-installation message\n'
        b'Installing Marathon app for package [helloworld] version '
        b'[0.1.0]\n' + HELLOWORLD_CLI_STDOUT +
        b'A sample post-installation message\n'
    )

    uninstall_stderr = b'Uninstalled package [helloworld] version [0.1.0]\n'
    return _package(name='helloworld',
                    args=['--yes'],
                    stderr=stderr,
                    uninstall_stderr=uninstall_stderr)


@contextlib.contextmanager
def _helloworld_cli():
    args = ['--yes', '--cli']
    command = ['dcos', 'package', 'install', 'helloworld'] + args

    installed = False
    timeout = http.DEFAULT_READ_TIMEOUT
    try:
        returncode_, _, _ = exec_command(command, timeout=timeout)
        installed = (returncode_ == 0)
        assert installed

        yield
    finally:
        if installed:
            # TODO(janisz): Cahnge back to package uninstall after DCOS_OSS-5619
            command = ['dcos', 'plugin', 'remove', 'dcos-http']
            assert_command(command)
            watch_all_deployments()


@contextlib.contextmanager
def _package(name,
             args,
             stdout=b'',
             stderr=b'',
             uninstall_confirmation=True,
             uninstall_app_id='',
             uninstall_stderr=b''):
    """Context manager that installs a package on entrance, and uninstalls it on
    exit.

    :param name: package name
    :type name: str
    :param args: extra CLI args
    :type args: [str]
    :param stdout: Expected stdout
    :param stderr: Expected stderr
    :type stdout: bytes
    :param uninstall_app_id: App id for uninstallation
    :type uninstall_app_id: string
    :param uninstall_stderr: Expected stderr
    :type uninstall_stderr: bytes
    :rtype: None
    """

    command = ['dcos', 'package', 'install', name] + args

    installed = False
    timeout = http.DEFAULT_READ_TIMEOUT
    try:
        returncode_, stdout_, stderr_ = exec_command(command, timeout=timeout)
        installed = (returncode_ == 0)

        assert stderr_ == stderr
        assert installed
        assert stdout_ == stdout

        yield
    finally:
        if installed:
            command = ['dcos', 'package', 'uninstall', name]
            if uninstall_confirmation:
                command.append('--yes')
            if uninstall_app_id:
                command.append('--app-id='+uninstall_app_id)
            assert_command(command, stderr=uninstall_stderr)
            watch_all_deployments()


def _repo_add(args=[], repo_list=[]):
    assert_command(['dcos', 'package', 'repo', 'add'] + args)
    assert_command(['dcos', 'package', 'repo', 'list'], stdout=repo_list)


def _repo_remove(args=[], repo_list=[]):
    assert_command(['dcos', 'package', 'repo', 'remove'] + args)
    assert_command(['dcos', 'package', 'repo', 'list'], stdout=repo_list)
