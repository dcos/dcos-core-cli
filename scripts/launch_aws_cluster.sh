#!/bin/bash

set +x

bash <(curl -s https://raw.githubusercontent.com/dcos/dcos/master/test_util/terraform_init.sh)

./terraform init
./terraform apply

MASTER_PUBLIC_IP=$(./terraform output --json -module dcos.dcos-infrastructure masters.public_ips | jq -r '.value[0]')
echo ${MASTER_PUBLIC_IP}
