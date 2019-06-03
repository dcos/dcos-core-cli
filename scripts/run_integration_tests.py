#!/usr/bin/env python3

import os
import urllib.request
import socket
import subprocess
import sys
import time

from ipaddress import IPv4Address
from pathlib import Path
from urllib.parse import urlparse

import click
import pytest

from dcos_e2e.backends import AWS, Docker
from dcos_e2e.cluster import Cluster
from dcos_e2e.node import Node

from dcoscli.test.common import dcos_tempdir, exec_command
from passlib.hash import sha512_crypt


@click.command()
@click.option('--e2e-backend', type=click.Choice(['existing', 'dcos_launch', 'dcos_docker']), required=True)
@click.option('--installer-url', help='URL of the DC/OS installer.')
@click.option('--dcos-license', envvar='DCOS_TEST_LICENSE', help='Content of the license to use.')
@click.option('--dcos-url', envvar='DCOS_TEST_URL', help='Specifies the public URL or IP address of the master node.')
@click.option('--admin-username', help='Username for the admin user.', required=True)
@click.option('--admin-password', help='Password for the admin user.', required=True)
@click.option('--ssh-user', default='centos', help='SSH user for connecting to the cluster.')
@click.option('--ssh-key-path', type=click.Path(), help='Path to the private SSH key for connecting to the cluster.')
def run_tests(e2e_backend, installer_url, dcos_license, dcos_url, admin_username, admin_password, ssh_user, ssh_key_path):

    os.environ["CLI_TEST_SSH_USER"] = ssh_user
    os.environ["CLI_TEST_MASTER_PROXY"] = "1"
    os.environ["CLI_TEST_SSH_KEY_PATH"] = ssh_key_path

    # extra dcos_config (for dcos_launch and dcos_docker backends)
    extra_config = {
        'superuser_username': admin_username,
        'superuser_password_hash': sha512_crypt.hash(admin_password),
        'fault_domain_enabled': False,
        'license_key_contents': dcos_license,
    }

    if e2e_backend == 'dcos_launch':
        cluster_backend = AWS()

        with Cluster(cluster_backend=cluster_backend, agents=1) as cluster:
            dcos_config = {**cluster.base_config, **extra_config}

            cluster.install_dcos_from_url(
                dcos_installer=installer_url,
                dcos_config=dcos_config,
                ip_detect_path=AWS().ip_detect_path,
            )

            os.environ["CLI_TEST_SSH_KEY_PATH"] = str(cluster._cluster._ssh_key_path)

            _run_tests(cluster, admin_username, admin_password)
    elif e2e_backend == 'dcos_docker':
        dcos_ee_installer_filename = 'dcos_generate_config.ee.sh'
        dcos_ee_installer_path = Path.cwd() / Path(dcos_ee_installer_filename)

        if not dcos_ee_installer_path.exists():
            urllib.request.urlretrieve(installer_url, dcos_ee_installer_filename)

        with Cluster(cluster_backend=Docker(), agents=1) as cluster:
            dcos_config = {**cluster.base_config, **extra_config}

            cluster.install_dcos_from_path(
                dcos_installer=dcos_ee_installer_path,
                dcos_config=dcos_config,
                ip_detect_path=Docker().ip_detect_path,
            )

            _run_tests(cluster, admin_username, admin_password)
    elif e2e_backend == 'existing':
        try:
            dcos_ip = IPv4Address(dcos_url)
        except ValueError:
            parsed_dcos_url = urlparse(dcos_url)
            dcos_hostname = parsed_dcos_url.hostname
            dcos_ip = IPv4Address(socket.gethostbyname(dcos_hostname))

        masters = set([Node(
            public_ip_address=dcos_ip,
            private_ip_address=dcos_ip,
            ssh_key_path=Path(ssh_key_path),
            default_user=ssh_user,
        )])

        cluster = Cluster.from_nodes(
            masters=masters,
            agents=set(),
            public_agents=set(),
        )

        _run_tests(cluster, admin_username, admin_password)


def _run_tests(cluster, admin_username, admin_password):
    cluster.wait_for_dcos_ee(
        superuser_username=admin_username,
        superuser_password=admin_password,
    )

    master_node = next(iter(cluster.masters))
    master_ip = master_node.public_ip_address.exploded

    with dcos_tempdir():
        print(master_ip)
        exec_command(['dcos', 'cluster', 'setup', '--no-check', '--username',
                      admin_username, '--password', admin_password, master_ip])

        exec_command(['dcos', 'plugin', 'add', '-u', '../build/' + sys.platform + '/dcos-core-cli.zip'])

        os.chdir("../python/lib/dcoscli")

        retcode = pytest.main([
            '-vv',
            '-x',
            '--durations=10',
            '-p', 'no:cacheprovider',
            'tests/integrations'
        ])

    if retcode != 0:
        print("Sleeping for 5 minutes to leave room for manual debugging...")
        print(master_ip)
        time.sleep(300)

    sys.exit(retcode)

if __name__ == '__main__':
    run_tests(auto_envvar_prefix='DCOS_TEST')
