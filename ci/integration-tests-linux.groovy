#!/usr/bin/env groovy

def credentials = [
        [$class           : 'AmazonWebServicesCredentialsBinding',
         credentialsId    : 'a20fbd60-2528-4e00-9175-ebe2287906cf',
         accessKeyVariable: 'AWS_ACCESS_KEY_ID',
         secretKeyVariable: 'AWS_SECRET_ACCESS_KEY'],
        [$class       : 'FileBinding',
         credentialsId: '23743034-1ac4-49f7-b2e6-a661aee2d11b',
         variable     : 'CLI_TEST_SSH_KEY_PATH'],
        [$class       : 'StringBinding',
         credentialsId: '0b513aad-e0e0-4a82-95f4-309a80a02ff9',
         variable     : 'DCOS_TEST_INSTALLER_URL'],
        [$class       : 'StringBinding',
         credentialsId: 'ca159ad3-7323-4564-818c-46a8f03e1389',
         variable     : 'DCOS_TEST_LICENSE'],
        [$class          : 'UsernamePasswordMultiBinding',
         credentialsId   : '323df884-742b-4099-b8b7-d764e5eb9674',
         usernameVariable: 'DCOS_USERNAME',
         passwordVariable: 'DCOS_PASSWORD']
]

def master_ip = 'UNKNOWN'
def os = 'linux'

pipeline {
    agent { label 'mesos' }

    options {
        timeout(time: 6, unit: 'HOURS')
    }

    stages {
        stage("Build Go binary") {
            steps {
                sh "make ${os}"
                stash includes: "build/${os}/**", name: "dcos-${os}"
            }
        }

        stage("Launch AWS Cluster") {
            steps {
                withCredentials(credentials) {
                    script {
                        sh(script: '''
                                  cd scripts ; \
                                  export AWS_REGION="us-east-1" ; \
                                  export TF_VAR_dcos_user=$DCOS_USERNAME  ; \
                                  export TF_VAR_dcos_pass_hash=${mkpasswd -m sha-512 $DCOS_PASSWORD } ; \
                                  export TF_VAR_dcos_license_key_contents=$DCOS_TEST_LICENSE ; \
                                  export TF_VAR_custom_dcos_download_path=$DCOS_TEST_INSTALLER_URL ; \
                                  export CLI_TEST_SSH_KEY_PATH ; \
                                  export TF_INPUT=false ; \
                                  export TF_IN_AUTOMATION=1 ; \
                                  wget -nv https://releases.hashicorp.com/terraform/0.11.14/terraform_0.11.14_linux_amd64.zip ; \
                                  unzip -o terraform_0.11.14_linux_amd64.zip ; \
                                  mkdir -p $HOME/.ssh
                                  eval $(ssh-agent) ; \
                                  ssh-add $CLI_TEST_SSH_KEY_PATH ; \
                                  ssh-keygen -y -f $CLI_TEST_SSH_KEY_PATH > $HOME/.ssh/id_rsa.pub ; \
                                  ./terraform init ; \
                                  ./terraform  apply -auto-approve -no-color''')
                        master_ip = sh(script: 'cd scripts && ./terraform output masters_public_ip', returnStdout: true).trim()
                    }
                    stash includes: 'scripts', name: 'terraform'
                }
            }
        }

        stage("Run integration tests") {
            agent { label 'py37' }
            steps {
                unstash "dcos-${os}"
                withEnv(["DCOS_TEST_URL=${master_ip}", "OS=${os}"]) {
                    withCredentials(credentials) {
                        sh '''scripts/run_integration_tests.sh'''
                    }
                }
            }
        }
    }

    post {
        failure {
            echo 'Generate diagnostics bundle'
            withEnv(["DCOS_TEST_URL=${master_ip}"]) {
                withCredentials(credentials) {
                    dir("build/linux/") {
                        sh 'wget -qO ./dcos https://downloads.dcos.io/cli/testing/binaries/dcos/linux/x86-64/master/dcos'
                        sh 'chmod +x dcos'
                        sh './dcos cluster setup --no-check ${DCOS_TEST_URL}'
                        sh './dcos diagnostics create'
                        sh './dcos diagnostics wait'
                        sh './dcos diagnostics download'
                        archiveArtifacts artifacts: '*.zip', fingerprint: true
                    }
                }
            }
        }

        cleanup {
            echo 'Delete AWS Cluster'
            unstash 'terraform'
            withCredentials(credentials) {
                sh('''
                  cd scripts ; \
                  export AWS_REGION="us-east-1" ; \
                  export TF_INPUT=false ; \
                  export TF_IN_AUTOMATION=1 ; \
                  ./terraform destroy -auto-approve -no-color''')
            }
        }
    }

}