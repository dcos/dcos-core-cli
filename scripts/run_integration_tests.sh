#!/usr/bin/env bash

set -euxo pipefail

# prepare environment
export PYTHONIOENCODING=utf-8
export CLI_TEST_SSH_USER=centos
export CLI_TEST_MASTER_PROXY=1
export DCOS_DIR=$(mktemp -d /tmp/dcos.XXXXXXXXXX)
export PYTHON=python3.7
export LANG=en_US.utf-8
export LC_ALL=en_US.utf-8

test -f $CLI_TEST_SSH_KEY_PATH

# build the plugin
make plugin
cd python/lib/dcoscli
source env/bin/activate

# connect to cluster and install built plugin
wget -qO env/bin/dcos https://downloads.dcos.io/cli/testing/binaries/dcos/${OS}/x86-64/master/dcos
chmod +x env/bin/dcos
dcos cluster setup --no-check ${DCOS_TEST_URL} 2> /dev/null
dcos plugin add -u ../../../build/$OS/dcos-core-cli.zip

# run the tests
py.test -vv -x --durations=10 -p no:cacheprovider tests/integrations --junitxml=tests.xml

# clean up locally (cluster is left for cloudcleaner to clean up)
rm -rf $DCOS_DIR
