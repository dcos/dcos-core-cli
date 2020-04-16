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
def os = 'windows'

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
//                withCredentials(credentials) {
//                    script {
//                        master_ip = sh(script: 'cd scripts && ./launch_aws_cluster.sh', returnStdout: true).trim()
//                    }
//                    stash includes: 'scripts/**/*', name: 'terraform'
//                }
            }
        }

        stage("Stash dcos.exe") {
            steps {
                sh "wget -q https://downloads.dcos.io/cli/testing/binaries/dcos/${os}/x86-64/master/dcos.exe"
                stash includes: 'dcos.exe', name: 'dcos-exe'
            }
        }

        stage("Run integration tests") {
            agent { label 'windows' }
            steps {
                unstash "dcos-exe"
                unstash "dcos-${os}"
                withEnv(["DCOS_TEST_URL=${master_ip}", "OS=${os}"]) {
                    withCredentials(credentials) {
                        bat '''
                            bash -exc " \
                            export PYTHONIOENCODING=utf-8; \
                            export CLI_TEST_SSH_USER=centos; \
                            export CLI_TEST_MASTER_PROXY=1; \
                            mkdir -p build/windows; \
                            make python; \
                            python scripts/plugin/package_plugin.py; \
                            cd python/lib/dcoscli; \
                            make env; \
                            rm -f ./env/Scripts/dcos.exe; \
                            mv ../../../dcos.exe dist; \
                            PATH=$PWD/dist:$PATH; \
                            dcos cluster remove --all; \
                            dcos cluster setup --no-check ${DCOS_TEST_URL}; \
                            dcos plugin add -u ../../../build/$OS/dcos-core-cli.zip; \
                            ./env/Scripts/pytest -vv -x --durations=10 -p no:cacheprovider tests/integrations"'''
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
                  cd scripts && \
                  export AWS_REGION="us-east-1" && \
                  export TF_INPUT=false && \
                  export TF_IN_AUTOMATION=1 && \
                  ./terraform destroy -auto-approve -no-color''')
            }
        }
    }

}