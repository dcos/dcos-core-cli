#!/usr/bin/env bash

export PYTHONIOENCODING=utf-8
export CLI_TEST_SSH_USER=centos
export CLI_TEST_MASTER_PROXY=1
DCOS_EXPERIMENTAL=1 make plugin
cd python/lib/dcoscli
source env/bin/activate
wget -qO env/bin/dcos https://downloads.dcos.io/cli/testing/binaries/dcos/${OS}/x86-64/master/dcos
chmod +x env/bin/dcos
dcos cluster remove --all
dcos cluster setup --no-check ${DCOS_TEST_URL}
dcos plugin add -u ../../../build/$OS/dcos-core-cli.zip
py.test -vv -x --durations=10 -p no:cacheprovider tests/integrations