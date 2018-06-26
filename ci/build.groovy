#!/usr/bin/env groovy

@Library('sec_ci_libs@v2-latest') _

pipeline {
  agent none

  options {
    timeout(time: 2, unit: 'HOURS')
  }

  stages {
    stage('Check authorization') {
      when {
        expression { env.CHANGE_ID != null }
      }

      steps {
        user_is_authorized([] as String[], '8b793652-f26a-422f-a9ba-0d1e47eb9d89', '#dcos-cli-ci')
      }
    }

    stage('Build binaries') {
      parallel {
        stage('Build Linux binary') {
          agent {
            node {
              label 'py35'
              customWorkspace '/workspace'
            }
          }

          steps {
            sh '''
              bash -exc " \
                cd cli; \
                make binary"
            '''

            sh '''
              bash -exc "
                mkdir -p build/linux; \
                cp cli/dist/dcos build/linux/"
            '''

            stash includes: "build/**", name: "dcos-linux"
          }
        }

        stage('Build macOS binary') {
          agent { label 'mac-hh-yosemite' }

          steps {
            sh '''
              bash -exc " \
                cd cli; \
                make binary"
            '''

            sh '''
              bash -exc " \
                mkdir -p build/darwin; \
                cp cli/dist/dcos build/darwin/"
            '''

            stash includes: "build/**", name: "dcos-darwin"
          }
        }

        stage('Build Windows binary') {
          agent {
            node {
              label 'windows'
              customWorkspace 'C:\\windows\\workspace'
            }
          }

          steps {
            bat '''
              bash -exc " \
                cd cli; \
                make binary"
            '''

            bat '''
              bash -exc " \
                mkdir -p build/windows; \
                cp cli/dist/dcos.exe build/windows/"
            '''

            stash includes: "build/**", name: "dcos-windows"
          }
        }
      }
    }

    stage('Run tests') {
      when {
        expression { env.TAG_NAME == null }
      }
      parallel {
        stage('Run Linux tests') {
          agent {
            node {
              label 'py35'
              customWorkspace '/workspace'
            }
          }

          steps {
            sh '''
              bash -exc " \
                make env; \
                ./env/bin/tox -e py35-syntax; \
                ./env/bin/tox -e py35-unit"
            '''

            sh '''
              bash -exc " \
                cd cli; \
                make env; \
                ./env/bin/tox -e py35-syntax"
            '''
          }
        }

        stage('Run macOS tests') {
          agent { label 'mac-hh-yosemite' }

          steps {
            sh '''
              bash -exc " \
                make env; \
                ./env/bin/tox -e py35-syntax; \
                ./env/bin/tox -e py35-unit"
            '''

            sh '''
              bash -exc " \
                cd cli; \
                make env; \
                ./env/bin/tox -e py35-syntax"
            '''
          }
        }

        stage('Run Windows tests') {
          agent {
            node {
              label 'windows'
              customWorkspace 'C:\\windows\\workspace'
            }
          }

          steps {
            bat 'bash -c "rm -rf ${HOME}/.dcos"'

            bat '''
              bash -exc " \
                make env; \
                ./env/Scripts/tox -e py35-syntax; \
                ./env/Scripts/tox -e py35-unit"
            '''

            bat '''
              bash -exc " \
                cd cli; \
                make env; \
                ./env/Scripts/tox -e py35-syntax"
            '''
          }
        }
      }
    }

    stage("Publish binaries to S3") {
      when {
        anyOf {
          branch 'master'
          expression { env.TAG_NAME != null }
        }
      }

      agent { label 'py35' }

      steps {
        withCredentials([
            string(credentialsId: "8b793652-f26a-422f-a9ba-0d1e47eb9d89", variable: "SLACK_API_TOKEN"),
            string(credentialsId: "3f0dbb48-de33-431f-b91c-2366d2f0e1cf",variable: "AWS_ACCESS_KEY_ID"),
            string(credentialsId: "f585ec9a-3c38-4f67-8bdb-79e5d4761937",variable: "AWS_SECRET_ACCESS_KEY"),
        ]) {

            unstash "dcos-linux"
            unstash "dcos-darwin"
            unstash "dcos-windows"

            sh '''
              bash -exc " \
                ls build; \
                cd scripts; \
                python -m venv env; \
                source env/bin/activate; \
                pip install -r requirements.txt; \
                ./publish_binaries.py"
            '''
        }
      }
    }
  }
}
