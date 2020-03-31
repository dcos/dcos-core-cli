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
                        master_ip = sh(script: '''docker build -t cluster-starter:test -f test.Dockerfile ./;\
                                  docker run --rm -v $PWD:/usr/src -w /usr/src \
                                    -v $(pwd)/scripts:/mnt/scripts \
                                    -v ${CLI_TEST_SSH_KEY_PATH}:${CLI_TEST_SSH_KEY_PATH} \
                                    -e DCOS_TEST_INSTALLER_URL \
                                    -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY \
                                    -e DCOS_TEST_LICENSE -e CLI_TEST_SSH_KEY_PATH \
                                    -e "TF_VAR_variant=open" \
                                    cluster-starter:test bash -c " \
                                        export TF_VAR_cluster_name='ui-\$(date +%s)'; \
                                        echo $TF_VAR_cluster_name > /tmp/cluster_name-open; \
                                        mkdir -p /tmp/ssh; \
                                        cd scripts/terraform && ./up.sh | tail -n1"
                                    ''', returnStdout: true).trim()
                    }
                    // script{
                    //     master_ip = sh(script: '''
                    //         ENV TERRAFORM_VERSION=0.11.14

                    //         RUN cd /tmp \
                    //         && apt-get install -y unzip
                    //         && wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip \
                    //         && unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/bin
                    //         export TF_VAR_cluster_name="core-cli-\$(date +%s)"
                    //         echo $TF_VAR_cluster_name > /tmp/cluster_name-open
                    //         cd scripts/terraform && ./up.sh | tail -n1
                    //     ''', returnStdout: true).trim()
                    // }
                    // script {
                    //     // master_ip = sh(script: '''docker run --rm -v $PWD:/usr/src -w /usr/src \
                    //     //               -v ${CLI_TEST_SSH_KEY_PATH}:${CLI_TEST_SSH_KEY_PATH} \
                    //     //               -e DCOS_TEST_INSTALLER_URL \
                    //     //               -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY \
                    //     //               -e DCOS_USERNAME -e DCOS_PASSWORD \
                    //     //               -e DCOS_TEST_LICENSE -e CLI_TEST_SSH_KEY_PATH \
                    //     //               python:3.7-stretch bash -c " \
                    //     //                 cd scripts; \
                    //     //                 python3 -m venv env; \
                    //     //                 source env/bin/activate; \
                    //     //                 pip -q install --upgrade pip==18.1 setuptools; \
                    //     //                 pip -q install -r requirements.txt; \
                    //     //                 ./launch_aws_cluster.py create"''',
                    //     //         returnStdout: true).trim()
                    // }
                    // stash includes: 'scripts/config.json', name: 'aws-config'
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
        always {
            withCredentials(credentials) {
            sh '''
                export TF_VAR_cluster_name=$(cat /tmp/cluster_name-open)
                cd scripts/terraform && ./down.sh
            '''
            }
        }
//TODO(janisz): Uncomment once we have proper permisions in our CI to perform cleanup
//        cleanup {
//            echo 'Delete AWS Cluster'
//            unstash 'aws-config'
//            withCredentials(credentials) {
//                sh('''docker run --rm -v $PWD:/usr/src -w /usr/src \
//                      -v ${CLI_TEST_SSH_KEY_PATH}:${CLI_TEST_SSH_KEY_PATH} \
//                      python:3.7-stretch bash -c " \
//                        cd scripts; \
//                        python3 -m venv env; \
//                        source env/bin/activate; \
//                        pip -q install --upgrade pip==18.1 setuptools; \
//                        pip -q install -r requirements.txt; \
//                        ./launch_aws_cluster.py delete"''')
//            }
//        }
    }

}
