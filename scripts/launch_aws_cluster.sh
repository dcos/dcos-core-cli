#!/bin/sh

export AWS_REGION="us-east-1"
export TF_VAR_dcos_user=$DCOS_USERNAME
export TF_VAR_dcos_pass_hash=$(perl -e 'print crypt($ENV{DCOS_PASSWORD},"\$6\$1234567890\$")')
export TF_VAR_dcos_license_key_contents=$DCOS_TEST_LICENSE
export TF_VAR_custom_dcos_download_path=$DCOS_TEST_INSTALLER_URL
export CLI_TEST_SSH_KEY_PATH
export TF_INPUT=false
export TF_IN_AUTOMATION=1
wget -q https://releases.hashicorp.com/terraform/0.11.14/terraform_0.11.14_linux_amd64.zip
unzip -qq -o terraform_0.11.14_linux_amd64.zip
mkdir -p $HOME/.ssh
eval $(ssh-agent)
ssh-add $CLI_TEST_SSH_KEY_PATH
ssh-keygen -y -f $CLI_TEST_SSH_KEY_PATH > $HOME/.ssh/id_rsa.pub
./terraform init -no-color
./terraform  apply -auto-approve -no-color
./terraform output masters_public_ip