#!/usr/bin/env groovy

def credentials = [
  [$class: 'AmazonWebServicesCredentialsBinding',
   credentialsId: 'a20fbd60-2528-4e00-9175-ebe2287906cf',
   accessKeyVariable: 'AWS_ACCESS_KEY_ID',
   secretKeyVariable: 'AWS_SECRET_ACCESS_KEY'],
  [$class: 'FileBinding',
   credentialsId: '23743034-1ac4-49f7-b2e6-a661aee2d11b',
   variable: 'DCOS_TEST_SSH_KEY_PATH'],
  [$class: 'StringBinding',
   credentialsId: '0b513aad-e0e0-4a82-95f4-309a80a02ff9',
   variable: 'DCOS_TEST_INSTALLER_URL'],
  [$class: 'StringBinding',
   credentialsId: 'ca159ad3-7323-4564-818c-46a8f03e1389',
   variable: 'DCOS_TEST_LICENSE'],
  [$class: 'UsernamePasswordMultiBinding',
   credentialsId: '323df884-742b-4099-b8b7-d764e5eb9674',
   usernameVariable: 'DCOS_USERNAME',
   passwordVariable: 'DCOS_PASSWORD']
]

node('py36') {
    ws('/jenkins/workspace') {

    dir("dcos-core-cli") {
        checkout scm
    }

    stash includes: "dcos-core-cli/**", name: "dcos-core-cli"

    dir("dcos-core-cli/scripts") {
        withCredentials(credentials) {
            def master_ip = sh(
                script: '''bash -ec " \
                    python -m venv env >&2; \
                    source env/bin/activate; \
                    pip install --upgrade pip >&2; \
                    pip install -r requirements.txt >&2; \
                    ./launch_aws_cluster.py"
                ''',
                returnStdout: true
            ).trim()

            print(master_ip)

            withEnv(["DCOS_TEST_URL=${master_ip}"]) {
                node('windows') {
                    ws('C:\\windows\\workspace') {
                        dir("dcos-core-cli") {
                            stage('Cleanup workspace') {
                                deleteDir()
                            }
                        }

                        // Cannot `checkout scm` here.
                        // https://mesosphere.slack.com/archives/C5U03P5T6/p1527867956000237
                        unstash "dcos-core-cli"

                        // The run_integration_tests.py script triggers
                        // `ImportError: No module named 'termios'` on Windows.
                        // There were a workaround for this issue in the licensing
                        // CLI tests, but it is not addressed in dcos e2e.
                        dir('dcos-core-cli') {
                            bat '''
                                bash -exc " \
                                export PYTHONIOENCODING=utf-8; \
                                export CLI_TEST_SSH_USER=centos; \
                                export CLI_TEST_SSH_KEY_PATH=${DCOS_TEST_SSH_KEY_PATH}; \
                                export CLI_TEST_MASTER_PROXY=true; \
                                mkdir -p build/windows; \
                                make python; \
                                python scripts/plugin/package_plugin.py; \
                                cd python/lib/dcoscli; \
                                python -c 'import urllib.request; urllib.request.urlretrieve(\\\"https://downloads.dcos.io/binaries/cli/windows/x86-64/latest/dcos.exe\\\", \\\"dist/dcos.exe\\\")'; \
                                dist/dcos cluster remove --all; \
                                dist/dcos cluster setup ${DCOS_TEST_URL} --insecure; \
                                dist/dcos plugin add ../../../build/windows/dcos-core-cli.zip; \
                                PATH=$PWD/dist:$PATH; \
                                make env; \
                                ./env/Scripts/tox -e py35-integration"
                            '''
                        }
                    }
                }
            }
        }
    }
    }
}
