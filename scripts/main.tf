provider "aws" {
  region = "us-west-2"
}

module "dcos" {
  source  = "dcos-terraform/dcos/aws"
  version = "~> 0.2.0"

  cluster_name        = "dcos-core-cli-e2e-tests-"
  cluster_name_random_string = true

  ssh_public_key_file = "id_rsa.pub"
  admin_ips           = ["0.0.0.0/0"]

  num_masters        = "1"
  num_private_agents = "1"
  num_public_agents  = "0"

  dcos_version = "2.0.0"

  dcos_instance_os    = "centos_7.5"
  bootstrap_instance_type = "m4.large"
  masters_instance_type  = "m4.large"
  private_agents_instance_type = "m4.large"
  public_agents_instance_type = "m4.large"

  dcos_variant = "open"

  providers = {
    aws = "aws"
  }

  tags = {
    build_type_id = "CLI_Integration_Test"
  }
}

output "masters-ips" {
  value = module.dcos.masters-ips
}

output "cluster-address" {
  value = module.dcos.masters-loadbalancer
}

output "public-agents-loadbalancer" {
  value = module.dcos.public-agents-loadbalancer
}