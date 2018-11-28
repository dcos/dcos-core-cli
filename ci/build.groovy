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

        stage('Build Windows binary') {
          agent {
            node {
              label 'winpy354'
              customWorkspace 'C:\\windows\\workspace'
            }
          }

          steps {
            bat '''
              bash -exc " \
                cd python/lib/dcoscli; \
                make binary"
            '''

            bat '''
              bash -exc " \
                mkdir -p build/windows/python; \
                cp python/lib/dcoscli/dist/dcos.exe build/windows/python/"
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
                cd python/lib/dcos; \
                make env; \
                ./env/bin/tox -e py35-syntax; \
                ./env/bin/tox -e py35-unit"
            '''

            sh '''
              bash -exc " \
                cd python/lib/dcoscli; \
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
                cd python/lib/dcos; \
                make env; \
                ./env/bin/tox -e py35-syntax; \
                ./env/bin/tox -e py35-unit"
            '''

            sh '''
              bash -exc " \
                cd python/lib/dcoscli; \
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
                cd python/lib/dcos; \
                make env; \
                ./env/Scripts/tox -e py35-syntax; \
                ./env/Scripts/tox -e py35-unit"
            '''

            bat '''
              bash -exc " \
                cd python/lib/dcoscli; \
                make env; \
                ./env/Scripts/tox -e py35-syntax"
            '''
          }
        }
      }
    }

    stage("Publish binaries and plugins to S3") {
      when {
        expression { env.CHANGE_ID == null }
      }

      agent { label 'py36' }

      steps {
        withCredentials([
            string(credentialsId: "8b793652-f26a-422f-a9ba-0d1e47eb9d89", variable: "SLACK_API_TOKEN"),
            string(credentialsId: "e270aa3f-4825-480c-a3ec-18a541c4e2d1",variable: "AWS_ACCESS_KEY_ID"),
            string(credentialsId: "cd616d55-78eb-45de-b7a8-e5bc5ccce4c7",variable: "AWS_SECRET_ACCESS_KEY"),
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
                pip install --upgrade pip setuptools; \
                pip install -r requirements.txt; \
                ./publish_plugins.py"
            '''
        }
      }
    }
  }
}
