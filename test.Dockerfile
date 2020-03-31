FROM ubuntu:latest

ENV TERRAFORM_VERSION=0.11.14

# Install required components & prepare environment
RUN set -x \
  && apt-get update \
  && apt-get install -y awscli \
     lsof \
     wget \
     jq \
     curl \
     rsync \
     openssh-client \
     unzip \
  && apt-get clean \
  && curl -o /usr/local/bin/dcos https://downloads.dcos.io/cli/testing/binaries/dcos/linux/x86-64/master/dcos \
  && chmod +x /usr/local/bin/dcos 

RUN cd /tmp \
  && wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip \
  && unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/bin