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
        stage("Build Go binaries") {
        agent { label 'mesos-ubuntu' }

        steps {
          sh 'make test darwin linux windows'
          junit 'report.xml'
          stash includes: 'build/**', name: 'dcos-core-go'
        }
        }
        stage('Build Linux binary') {
          agent {
            node {
              label 'mesos'
              customWorkspace '/workspace'
            }
          }

          steps {

            sh '''
                docker run --rm -v $PWD:/usr/src -w /usr/src \
                    python:3.7-stretch bash -exc " \
                      cd python/lib/dcoscli; \
                      make binary"
            '''

            sh '''
              bash -exc "
                mkdir -p build/linux/python; \
                cp python/lib/dcoscli/dist/dcos build/linux/python/"
            '''

            stash includes: "build/**", name: "dcos-linux"
          }
        }

        stage('Build macOS binary') {
          agent { label 'mac-hh-yosemite' }

          steps {
            sh '''
              bash -exc " \
                export PYTHON=python3.7; \
                cd python/lib/dcoscli; \
                make binary"
            '''

            sh '''
              bash -exc " \
                mkdir -p build/darwin/python; \
                cp python/lib/dcoscli/dist/dcos build/darwin/python/"
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
              label 'mesos'
              customWorkspace '/workspace'
            }
          }

          steps {
            sh '''
              docker run --rm -v $PWD:/usr/src -w /usr/src \
                python:3.7-stretch bash -exc " \
                  cd python/lib/dcos; \
                  make env; \
                  ./env/bin/tox -e py35-syntax; \
                  ./env/bin/tox -e py35-unit"
            '''

            sh '''
              docker run --rm -v $PWD:/usr/src -w /usr/src \
                python:3.7-stretch bash -exc " \
                  export PYTHON=python3.7; \
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
                export PYTHON=python3.7; \
                cd python/lib/dcos; \
                make env; \
                ./env/bin/tox -e py35-syntax; \
                ./env/bin/tox -e py35-unit"
            '''

            sh '''
              bash -exc " \
                export PYTHON=python3.7; \
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
            unstash "dcos-core-go"

            sh '''
              bash -exc " \
                ls build; \
                cd scripts; \
                python -m venv env; \
                source env/bin/activate; \
                pip install --upgrade pip==18.1 setuptools; \
                pip install -r requirements.txt; \
                ./publish_plugins.py"
            '''
        }
      }
    }
  }
}
