#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
: ${1?"Usage: $0 major.minor.patch revision"}
: ${2?"Usage: $0 major.minor.patch revision"}
set -e
# note
# script params
HABANA_RELEASE_VERSION=$1
HABANA_RELEASE_ID=$2
# env params
# MIN_PYTHON_VER="${PYTHON_VERSION:-3}"
PIP_PYTHON_OPTIONS="${PYTHON_OPTIONS:-}"
EXTRA_INDEX_URL="${PYTHON_INDEX_URL:-}"
PYTHON_MPI_VERSION="${MPI_VERSION:-3.1.6}"
# define constants
PILLOW_SIMD_VERSION="9.5.0.post1"

"${CONDA_DIR}/envs/gaudi/bin/pip" install mpi4py=="${PYTHON_MPI_VERSION}" ${PIP_PYTHON_OPTIONS}
"${CONDA_DIR}/envs/gaudi/bin/pip" install habana-pyhlml=="${HABANA_RELEASE_VERSION}"."${HABANA_RELEASE_ID}" ${PIP_PYTHON_OPTIONS} ${EXTRA_INDEX_URL}
"${CONDA_DIR}/envs/gaudi/bin/pip" install /tmp/installations/*.whl -r /tmp/installations/requirements-pytorch.txt ${PIP_PYTHON_OPTIONS} --disable-pip-version-check --no-warn-script-location
"${CONDA_DIR}/envs/gaudi/bin/pip" uninstall -y pillow 2>/dev/null || echo "Skip uninstalling pillow. Need SUDO permissions."
"${CONDA_DIR}/envs/gaudi/bin/pip" uninstall -y pillow-simd 2>/dev/null || echo "Skip uninstalling pillow-simd. Need SUDO permissions."
"${CONDA_DIR}/envs/gaudi/bin/pip" install pillow-simd==${PILLOW_SIMD_VERSION} ${PIP_PYTHON_OPTIONS} --disable-pip-version-check
