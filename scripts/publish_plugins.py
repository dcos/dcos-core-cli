#!/usr/bin/env python3

import boto3

from plugin import package_plugin
import os

version = os.environ.get("TAG_NAME", '1.12')

build_path = os.path.dirname(os.path.realpath(__file__)) + "/../build"

platforms = ['linux', 'darwin', 'windows']

for platform in platforms:
    plugin_path = build_path + '/' + platform + '/plugin'

    package_plugin(plugin_path, platform)


s3_client = boto3.resource('s3', region_name='us-west-2').meta.client

for platform in platforms:
    zip_file = platform + "/dcos-core-cli.zip"
    bucket_key = 'cli/plugins/dcos-core-cli/{}/{}/x86-64/dcos-core-cli.zip'.format(version, platform)
    s3_client.upload_file(build_path + "/" + zip_file, "downloads.dcos.io", bucket_key)
