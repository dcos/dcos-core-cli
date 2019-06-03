#!/usr/bin/env groovy

pipeline {
  agent none

  options {
    timeout(time: 6, unit: 'HOURS')
  }

  stages {
    stage("Build Go binary") {
      agent { label 'mesos-ubuntu' }

      steps {
          sh 'make linux'
          stash includes: 'build/linux/**', name: 'dcos-linux'
      }
    }

    stage("Run Linux integration tests") {
      agent { label 'py36' }

      steps {
        withCredentials([
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
          usernameVariable: 'DCOS_TEST_ADMIN_USERNAME',
          passwordVariable: 'DCOS_TEST_ADMIN_PASSWORD']
        ]) {
          unstash 'dcos-linux'

          sh '''
            bash -exc " \
              export DCOS_EXPERIMENTAL=1; \
              mkdir -p build/linux; \
              make plugin; \
              cd scripts; \
              python3 -m venv env; \
              source env/bin/activate; \
              pip install --upgrade pip==18.1 setuptools; \
              pip install -r requirements.txt; \
              wget -O env/bin/dcos https://downloads.dcos.io/cli/testing/binaries/dcos/linux/x86-64/master/dcos; \
              dcos cluster remove --all; \
              ./run_integration_tests.py --e2e-backend=dcos_launch"
          '''
        }
      }
    }
  }
}
