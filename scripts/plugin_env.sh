#!/usr/bin/env bash

echo "export DCOS_URL=$(dcos config show core.dcos_url)"
echo "export DCOS_ACS_TOKEN=$(dcos config show core.dcos_acs_token)"
echo "export DCOS_TLS_INSECURE=1"
