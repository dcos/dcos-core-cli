#!/bin/bash

CURRDIR=$(dirname "${0}")
source ${CURRDIR}/common.sh

if [ ! -d "${BUILDDIR}/${VENV}" ]; then
    # Check for required prerequisites.
    echo "Checking prerequisites..."
    if [ ! "$(command -v ${PYTHON})" ]; then
      echo "Cannot find python. Exiting..."
      exit 1
    fi

    PYTHON_MAJOR=$(${PYTHON} -c 'import sys; print(sys.version_info[0])')
    PYTHON_MINOR=$(${PYTHON} -c 'import sys; print(sys.version_info[1])')

    : "${DCOS_EXPERIMENTAL:=""}"
    if [ "${DCOS_EXPERIMENTAL}" = "" ]; then

      # On Windows, our build scripts rely on virtualenv instead of venv.
      # This highlighted issues when trying to upgrade to Python 3.7.
      # Given that the issue solved by upgrading to Python 3.7 is UNIX specific
      # (https://jira.mesosphere.com/browse/DCOS-52180), and that we'll have the Python
      # codebase completely removed for DC/OS 1.15, we're sticking to Python 3.5 on Windows
      # instead of taking more time to try to upgrade to Python 3.7.
      if [[ "$(uname)" =~ MINGW64_NT ]]; then
        if [ "${PYTHON_MAJOR}" != "3" ] || [ "${PYTHON_MINOR}" != "5" ]; then
            echo "Cannot find supported python version 3.5. Exiting..."
            exit 1
        fi
      else
        if [ "${PYTHON_MAJOR}" != "3" ] || [ "${PYTHON_MINOR}" != "7" ]; then
            echo "Cannot find supported python version 3.7. Exiting..."
            exit 1
        fi
      fi
    fi
    echo "Prerequisite checks passed."

    # Create the virtualenv.
    echo "Creating virtualenv..."
      ${PYTHON} -m venv ${BUILDDIR}/${VENV}
      sed -i'' -e "s#(${VENV}) #${PROMPT}#g" ${BUILDDIR}/${VENV}/${BIN}/activate
    echo "Virtualenv created: ${BUILDDIR}/${VENV}"

    # Install all requirements into the virtualenv.
    echo "Installing virtualenv requirements..."
    if [[ "$(uname)" =~ MINGW64_NT ]]; then
      ${PYTHON} -m pip install -U pip==18.1
    else
      ${BUILDDIR}/${VENV}/${BIN}/pip${EXE} install --upgrade pip==18.1
    fi
    ${BUILDDIR}/${VENV}/${BIN}/pip${EXE} install -r ${BASEDIR}/requirements.txt
    ${BUILDDIR}/${VENV}/${BIN}/pip${EXE} install -e ${BASEDIR}
    echo "Virtualenv requirements installed."
else
    echo "Virtualenv already exists: '${BUILDDIR}/${VENV}'"
fi
