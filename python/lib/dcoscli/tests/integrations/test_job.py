import contextlib
import json
import os
import pytest
import retrying

from dcos import constants

from dcoscli.test.common import assert_command, exec_command, update_config
from dcoscli.test.job import job, show_job, show_job_schedule


def test_help():
    with open('dcoscli/data/help/job.txt') as content:
        assert_command(['dcos', 'job', '--help'],
                       stdout=content.read().encode('utf-8'))


@pytest.fixture
def env():
    r = os.environ.copy()
    r.update({constants.PATH_ENV: os.environ[constants.PATH_ENV]})

    return r


def test_job_list_unauthorized():
    with update_config('core.dcos_acs_token', None):
        assert_command(
            ['dcos', 'job', 'list'],
            stdout=b"",
            stderr=(b'Error: authentication failed, '
                    b'please run `dcos auth login`\n'),
            returncode=1)


def test_empty_list():
    _list_jobs()


def test_add_job():
    with _no_schedule_instance_job():
        _list_jobs('pikachu')


def test_add_job_with_schedule():
    with _schedule_instance_job():
        _list_jobs('snorlax')


def test_add_job_with_env():
    with _env_instance_job():
        _list_jobs('job-env-secret')


def test_show_job_schedule():
    with _schedule_instance_job():
        show_job_schedule('snorlax', 'snore-nightly')


def test_add_job_bad_resource():
    returncode, stdout, stderr = exec_command(
        ['dcos', 'job', 'add', 'bad_resource'])

    assert returncode == 1
    assert "Error: open bad_resource:" in stderr.decode('utf-8')


def test_add_bad_json_job():
    returncode, stdout, stderr = exec_command(
        ['dcos', 'job', 'add', 'tests/data/metronome/jobs/bad.json'])

    assert returncode == 1
    assert stderr.decode('utf-8').startswith(
        'Error: unexpected end of JSON input')


def test_show_job():
    with _no_schedule_instance_job():
        show_job('pikachu')


def test_show_job_with_blank_jobname():
    returncode, stdout, stderr = exec_command(
        ['dcos', 'job', 'show'])

    assert returncode == 1
    assert "Error: accepts 1 arg(s)" in stderr.decode('utf-8')


def test_show_job_with_invalid_jobname():
    assert_command(
        ['dcos', 'job', 'show', 'invalid'],
        stdout=b'',
        stderr=b'Error: job "invalid" does not exist\n',
        returncode=1)


def test_show_job_runs_blank_jobname():
    returncode, stdout, stderr = exec_command(
        ['dcos', 'job', 'show', 'runs'])

    assert returncode == 1
    assert "Error: accepts 1 arg(s)" in stderr.decode('utf-8')


def test_show_schedule_blank_jobname():
    returncode, stdout, stderr = exec_command(
        ['dcos', 'job', 'schedule', 'show'])

    assert returncode == 1
    assert stderr.decode('utf-8').startswith('Error: accepts 1 arg(s)')


def test_show_job_queue_blank():
    returncode, stdout, stderr = exec_command(
        ['dcos', 'job', 'queue', '--json'])

    result = json.loads(stdout.decode('utf-8'))

    assert len(result) == 0


def test_show_job_queue_blank_for_job():
    assert_command(
        ['dcos', 'job', 'queue', 'magikarp'],
        stdout=b"",
        stderr=b'Error: job "magikarp" does not exist\n',
        returncode=1)


def test_show_schedule_invalid_jobname():
    assert_command(
        ['dcos', 'job', 'schedule', 'show', 'invalid'],
        stdout=b'',
        stderr=b'Error: job "invalid" does not exist\n',
        returncode=1)


def test_remove_job():
    with _no_schedule_instance_job():
        pass
    _list_jobs()


def test_update_job():
    with _no_schedule_instance_job():

        original = show_job('pikachu')
        _update_job(
            'pikachu',
            'tests/data/metronome/jobs/update-pikachu.json')

        result = show_job('pikachu')
        assert original['run']['cmd'] != result['run']['cmd']


def test_no_history():

    returncode, stdout, stderr = exec_command(
        ['dcos', 'job', 'history', 'BAD'])

    assert returncode == 1


def test_no_history_with_job():
    with _no_schedule_instance_job():

        returncode, stdout, stderr = exec_command(
            ['dcos', 'job', 'history', 'pikachu'])

        assert returncode == 0


def test_show_runs():
    with _no_schedule_instance_job():

        _run_job('pikachu')

        returncode, stdout, stderr = exec_command(
            ['dcos', 'job', 'show', 'runs', 'pikachu'])

        assert returncode == 0
        assert 'JOB ID' in stdout.decode('utf-8')
        assert 'pikachu' in stdout.decode('utf-8')


def test_run_json():
    with _no_schedule_instance_job():
        returncode, stdout, stderr = exec_command(
            ['dcos', 'job', 'run', '--json', 'pikachu'])
        assert returncode == 0
        assert '"jobId": "pikachu",' in stdout.decode('utf-8')


def _run_job(job_id):
    returncode, stdout, stderr = exec_command(
        ['dcos', 'job', 'run', job_id])

    assert returncode == 0


def test_show_queue():
    with _no_schedule_instance_large_job():
        _run_job('gyarados')

        @retrying.retry(stop_max_attempt_number=5, wait_fixed=2000)
        def dcos_job_queue_check():
            returncode, stdout, stderr = exec_command(
                ['dcos', 'job', 'queue'])

            assert returncode == 0
            assert 'JOB ID' in stdout.decode('utf-8')
            assert 'RUN ID' in stdout.decode('utf-8')
            assert 'gyarados' in stdout.decode('utf-8')

        dcos_job_queue_check()


@contextlib.contextmanager
def _no_schedule_instance_job():
    with job('tests/data/metronome/jobs/pikachu.json',
             'pikachu'):
        yield


@contextlib.contextmanager
def _schedule_instance_job():
    with job('tests/data/metronome/jobs/snorlax.json',
             'snorlax'):
        yield


@contextlib.contextmanager
def _env_instance_job():
    with job('tests/data/metronome/jobs/job-env.json',
             'job-env-secret'):
        yield


def _update_job(app_id, file_path):
    assert_command(['dcos', 'job', 'update', file_path])


@contextlib.contextmanager
def _no_schedule_instance_large_job():
    with job('tests/data/metronome/jobs/gyarados.json',
             'gyarados'):
        yield


def _list_jobs(app_id=None):
    returncode, stdout, stderr = exec_command(
        ['dcos', 'job', 'list', '--json'])

    result = json.loads(stdout.decode('utf-8'))

    if app_id is None:
        assert len(result) == 0
    else:
        assert len(result) == 1
        assert result[0]['id'] == app_id

    assert returncode == 0
    assert stderr == b''

    return result
