import json
import os
import sys

import pytest
import retrying
import six

import dcos.util as util
from dcos import mesos
from dcos.util import create_schema

from dcoscli.test.common import (assert_command, assert_lines, exec_command,
                                 fetch_valid_json, ssh_output)
from ..fixtures.node import slave_fixture


def test_help():
    with open('dcoscli/data/help/node.txt') as content:
        planned_stdout = six.b(content.read())
    returncode, stdout, stderr = exec_command(['dcos', 'node', '--help'])
    assert returncode == 0
    assert stderr == b''
    assert stdout == planned_stdout


def test_node():
    returncode, stdout, stderr = exec_command(
        ['dcos', 'node', 'list', '--json'])

    assert returncode == 0
    assert stderr == b''

    nodes = json.loads(stdout.decode('utf-8'))
    assert len(nodes) > 0
    slave_nodes = [node for node in nodes if node['type'] == 'agent']
    schema = _get_schema(slave_fixture())
    for node in slave_nodes:
        assert util.validate_json(node, schema)


def test_node_table():
    returncode, stdout, stderr = exec_command(['dcos', 'node', 'list'])

    assert returncode == 0
    assert stderr == b''
    assert len(stdout.decode('utf-8').split('\n')) > 2


def test_node_table_field_option():
    returncode, stdout, stderr = exec_command(
        ['dcos', 'node', 'list', '--field=used_resources.disk'])

    assert returncode == 0
    assert stderr == b''
    lines = stdout.decode('utf-8').splitlines()
    assert len(lines) > 2
    assert lines[0].split() == ["HOSTNAME", "IP", "PUBLIC", "IP(S)", "ID",
                                "TYPE", "STATUS", "REGION", "ZONE", "USED",
                                "RESOURCES", "DISK"]
    assert stdout.decode('utf-8').count("agent (public)") == 1


def test_node_table_uppercase_field_option():
    returncode, stdout, stderr = exec_command(
        ['dcos', 'node', 'list', '--field=TASK_RUNNING'])

    assert returncode == 0
    assert stderr == b''
    lines = stdout.decode('utf-8').splitlines()
    assert len(lines) > 2
    assert lines[0].split() == ["HOSTNAME", "IP", "PUBLIC", "IP(S)", "ID",
                                "TYPE", "STATUS", "REGION", "ZONE", "TASK",
                                "RUNNING"]


def test_node_log_empty():
    stderr = b"Error: '--leader' or '<mesos-id>' must be provided\n"
    assert_command(['dcos', 'node', 'log'], returncode=1, stderr=stderr)


def test_node_log_leader():
    assert_lines(['dcos', 'node', 'log', '--leader'], 10, greater_than=True)


def test_node_log_slave():
    slave_id = _node()[0]['id']
    assert_lines(
        ['dcos', 'node', 'log', slave_id],
        10,
        greater_than=True)


def test_node_log_missing_slave():
    returncode, _, stderr = exec_command(
        ['dcos', 'node', 'log', 'bogus'])

    assert returncode == 1
    assert b"'bogus' not found" in stderr


def test_node_log_lines():
    # since we are getting system logs, it's not guaranteed to get back
    # exactly 4 log entries. It must be >= 4
    assert_lines(
        ['dcos', 'node', 'log', '--leader', '--lines=4'],
        4,
        greater_than=True)


def test_node_log_invalid_lines():
    returncode, stdout, stderr = exec_command(
        ['dcos', 'node', 'log', '--leader', '--lines=bogus'])

    assert returncode == 1
    assert b'invalid argument' in stderr


def test_node_metrics_agent_summary():
    first_node_id = _node()[0]['id']

    @retrying.retry(stop_max_delay=30 * 1000)
    def _test_node_metrics_agent_summary():
        assert_lines(
            ['dcos', 'node', 'metrics', 'summary', first_node_id],
            2
        )
    _test_node_metrics_agent_summary()


def test_node_metrics_agent_summary_json():
    first_node_id = _node()[0]['id']

    @retrying.retry(stop_max_delay=30 * 1000)
    def _fetch_valid_json():
        node_json = fetch_valid_json(
            ['dcos', 'node', 'metrics', 'summary', first_node_id, '--json']
        )
        assert len(node_json) > 0
        return node_json

    node_json = _fetch_valid_json()

    metrics = [
        'cpu.total',
        'filesystem.capacity.total',
        'filesystem.capacity.used',
        'load.1min',
        'memory.total',
        'memory.free',
    ]

    for d in node_json:
        assert 'name' in d
        assert d['name'] in metrics


def test_node_metrics_agent_details():
    first_node_id = _node()[0]['id']

    @retrying.retry(stop_max_delay=30 * 1000)
    def _test_node_metrics_agent_details():
        assert_lines(
            ['dcos', 'node', 'metrics', 'details', first_node_id],
            100,
            greater_than=True
        )
    _test_node_metrics_agent_details()


def test_node_metrics_agent_details_json():
    first_node_id = _node()[0]['id']

    @retrying.retry(stop_max_delay=30 * 1000)
    def _fetch_valid_json():
        node_json = fetch_valid_json(
            ['dcos', 'node', 'metrics', 'details', first_node_id, '--json']
        )
        assert len(node_json) > 100
        return node_json

    node_json = _fetch_valid_json()

    metrics = [
        'cpu.idle',
        'cpu.system',
        'cpu.total',
        'cpu.user',
        'cpu.wait',
        'filesystem.capacity.free',
        'filesystem.capacity.total',
        'filesystem.capacity.used',
        'filesystem.inode.free',
        'filesystem.inode.total',
        'filesystem.inode.used',
        'load.15min',
        'load.1min',
        'load.5min',
        'memory.buffers',
        'memory.cached',
        'memory.free',
        'memory.total',
        'network.in',
        'network.in.dropped',
        'network.in.errors',
        'network.in.packets',
        'network.out',
        'network.out.dropped',
        'network.out.errors',
        'network.out.errors',
        'network.out.packets',
        'process.count',
        'swap.free',
        'swap.total',
        'swap.used',
        'system.uptime',
    ]

    for d in node_json:
        assert 'name' in d
        assert d['name'] in metrics


def test_node_dns():
    returncode, stdout, stderr = exec_command(
        ['dcos', 'node', 'dns', 'marathon.mesos', '--json'])

    result = json.loads(stdout.decode('utf-8'))

    assert returncode == 0
    assert stderr == b''
    assert result[0]['host'] == "marathon.mesos."
    assert 'ip' in result[0]


@pytest.mark.skipif(sys.platform == 'win32',
                    reason='No pseudo terminal on windows')
def test_node_ssh_leader():
    _node_ssh(['--leader'])


@pytest.mark.skipif(sys.platform == 'win32',
                    reason='No pseudo terminal on windows')
def test_node_ssh_slave():
    slave_id = mesos.DCOSClient().get_state_summary()['slaves'][0]['id']
    _node_ssh(['--mesos-id={}'.format(slave_id), '--master-proxy'])


@pytest.mark.skipif(sys.platform == 'win32',
                    reason='No pseudo terminal on windows')
def test_node_ssh_slave_with_private_ip():
    slave_ip = mesos.DCOSClient().get_state_summary()['slaves'][0]['hostname']
    _node_ssh(['--private-ip={}'.format(slave_ip), '--master-proxy'])


@pytest.mark.skipif(sys.platform == 'win32',
                    reason='No pseudo terminal on windows')
def test_node_ssh_option():
    stdout, stderr, _ = _node_ssh_output(
        ['--leader', '--option', 'Protocol=0'])
    assert b'ignoring bad proto spec' in stdout


@pytest.mark.skipif(sys.platform == 'win32',
                    reason='No pseudo terminal on windows')
def test_node_ssh_config_file():
    stdout, stderr, _ = _node_ssh_output(
        ['--leader', '--config-file', 'tests/data/node/ssh_config'])
    assert stdout == b''


@pytest.mark.skipif(sys.platform == 'win32',
                    reason='No pseudo terminal on windows')
def test_node_ssh_user():
    stdout, stderr, _ = _node_ssh_output(
        ['--master-proxy', '--leader', '--user=bogus', '--option',
         'BatchMode=yes'])
    assert stdout == b''
    assert b'Permission denied' in stderr


@pytest.mark.skipif(sys.platform == 'win32',
                    reason='No pseudo terminal on windows')
def test_node_ssh_master_proxy_no_agent():
    env = os.environ.copy()
    env.pop('SSH_AUTH_SOCK', None)

    returncode, stdout, stderr = exec_command(
        ['dcos', 'node', 'ssh', '--master-proxy', '--leader'],
        env=env)

    assert returncode == 1
    assert b'Permission denied' in stderr


@pytest.mark.skipif(sys.platform == 'win32',
                    reason='No pseudo terminal on windows')
def test_node_ssh_master_proxy():
    _node_ssh(['--leader', '--master-proxy'])


@pytest.mark.skipif(sys.platform == 'win32',
                    reason='No pseudo terminal on windows')
def test_node_ssh_with_command():
    leader_hostname = mesos.DCOSClient().get_state_summary()['hostname']
    _node_ssh(['--leader', '--master-proxy', '/opt/mesosphere/bin/detect_ip'],
              0, leader_hostname)


@pytest.mark.skipif(sys.platform == 'win32',
                    reason='No pseudo terminal on windows')
def test_node_ssh_slave_with_command():
    slave = mesos.DCOSClient().get_state_summary()['slaves'][0]
    _node_ssh(['--mesos-id={}'.format(slave['id']), '--master-proxy',
               '/opt/mesosphere/bin/detect_ip'], 0, slave['hostname'])


@pytest.mark.skipif(sys.platform == 'win32',
                    reason='No pseudo terminal on windows')
def test_node_ssh_slave_with_separated_command():
    slave = mesos.DCOSClient().get_state_summary()['slaves'][0]
    _node_ssh(['--mesos-id={}'.format(slave['id']), '--master-proxy', '--user',
               os.environ.get('CLI_TEST_SSH_USER'), '--',
               '/opt/mesosphere/bin/detect_ip'], 0, slave['hostname'])


@pytest.mark.skipif(True, reason='The agent should be recommissioned,'
                                 ' but that feature does not exist yet.')
def test_node_decommission():
    agents = mesos.DCOSClient().get_state_summary()['slaves']
    agents_count = len(agents)
    assert agents_count > 0

    agent_id = agents[0]['id']

    returncode, stdout, stderr = exec_command([
        'dcos', 'node', 'decommission', agent_id])

    exp_stdout = "Agent {} has been marked as gone.\n".format(agent_id)

    assert returncode == 0
    assert stdout.decode('utf-8') == exp_stdout
    assert stderr == b''

    new_agents = mesos.DCOSClient().get_state_summary()['slaves']
    assert (agents_count - 1) == len(new_agents)


def test_node_decommission_unexisting_agent():
    returncode, stdout, stderr = exec_command([
        'dcos', 'node', 'decommission', 'not-a-mesos-id'])

    assert returncode == 1
    assert stdout == b''
    assert b"not mark agent 'not-a-mesos-id' as gone" in stderr


def _node_ssh_output(args):
    cli_test_ssh_key_path = os.environ['CLI_TEST_SSH_KEY_PATH']

    if os.environ.get('CLI_TEST_SSH_USER') and \
            not any("--user" in a for a in args):
        args.extend(['--user', os.environ.get('CLI_TEST_SSH_USER')])

    if os.environ.get('CLI_TEST_MASTER_PROXY') and \
            '--master-proxy' not in args:
        args.append('--master-proxy')

    cmd = ('ssh-agent /bin/bash -c "ssh-add {} 2> /dev/null && ' +
           'dcos node ssh --option StrictHostKeyChecking=no ' +
           '    --option ConnectTimeout=5 {}"').format(
        cli_test_ssh_key_path,
        ' '.join(args))

    return ssh_output(cmd)


def _node_ssh(args, expected_returncode=None, expected_stdout=None):
    stdout, stderr, returncode = _node_ssh_output(args)
    assert returncode is expected_returncode, \
        'returncode = %r; stdout: = %s; stderr = %s' % (
            returncode, stdout, stderr)
    if expected_stdout is not None:
        assert stdout.decode('utf-8').startswith(expected_stdout)


def _get_schema(slave):
    schema = create_schema(slave, True)
    schema['required'].remove('reregistered_time')

    schema['required'].remove('reserved_resources')
    schema['properties']['reserved_resources']['required'] = []

    schema['required'].remove('unreserved_resources')
    schema['properties']['unreserved_resources']['required'] = []

    schema['properties']['used_resources']['required'].remove('ports')
    schema['properties']['offered_resources']['required'].remove('ports')

    schema['required'].remove('version')
    return schema


def _node():
    returncode, stdout, stderr = exec_command(['dcos', 'node', '--json'])

    assert returncode == 0

    return [n for n in json.loads(stdout.decode('utf-8'))
            if n['type'] == 'agent']
