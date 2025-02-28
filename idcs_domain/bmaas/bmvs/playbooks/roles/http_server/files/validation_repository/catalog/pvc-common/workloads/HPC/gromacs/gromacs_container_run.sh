#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

TESTCASE=${TESTCASE:-benchMEM}
GMX_MPI_RANKS=${NTMPI:-1}
GMX_OMP_THREADS=${NTOMP:-16}

docker run -it --rm --device /dev/dri:/dev/dri \
--ipc=host -v /dev/dri/by-path:/dev/dri/by-path \
--workdir /gromacs \
--env TESTCASE=${TESTCASE} \
--env NTMPI=${GMX_MPI_RANKS} \
--env NTOMP=${GMX_MPI_THREADS} \
gromacs2023 \
/bin/bash gromacs_run.sh

#--volume ${SCRIPT_DIR}/testcases:/gromacs/testcases \
