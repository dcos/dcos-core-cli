#!/usr/bin/env groovy

def builders = [:]

builders['Linux integration tests'] = {
    build 'integration-tests-linux'
}

builders['Mac integration tests'] = {
    build 'integration-tests-mac'
}

builders['Windows integration tests'] = {
    build 'integration-tests-windows'
}

node('mesos') {
    try {
        parallel builders
    } catch (Exception e) {
        withCredentials([
            [$class: 'StringBinding',
             credentialsId: '8b793652-f26a-422f-a9ba-0d1e47eb9d89',
             variable: 'SLACK_TOKEN']
        ]) {
            slackSend (
                channel: "#dcos-cli-ci",
                color: "danger",
                message: "*dcos-core-cli*\nNightly integration tests failed... :disappointed:\n${env.RUN_DISPLAY_URL}",
                teamDomain: "mesosphere",
                token: "${env.SLACK_TOKEN}",
            )
        }
        throw e;
    }
}
