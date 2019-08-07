from dcoscli.test.common import assert_command


def test_help_job():
    with open('dcoscli/data/help/job.txt') as content:
        assert_command(['dcos', 'help', 'job'],
                       stdout=content.read().encode('utf-8'))


def test_help_marathon():
    with open('dcoscli/data/help/marathon.txt') as content:
        assert_command(['dcos', 'help', 'marathon'],
                       stdout=content.read().encode('utf-8'))


def test_help_node():
    with open('dcoscli/data/help/node.txt') as content:
        assert_command(['dcos', 'help', 'node'],
                       stdout=content.read().encode('utf-8'))


# def test_help_package():
#     with open('dcoscli/data/help/package.txt') as content:
#         assert_command(['dcos', 'help', 'package'],
#                        stdout=content.read().encode('utf-8'))


def test_help_service():
    with open('dcoscli/data/help/service.txt') as content:
        assert_command(['dcos', 'help', 'service'],
                       stdout=content.read().encode('utf-8'))


# def test_help_task():
#     with open('dcoscli/data/help/task.txt') as content:
#         assert_command(['dcos', 'help', 'task'],
#                        stdout=content.read().encode('utf-8'))
