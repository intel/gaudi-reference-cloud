#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

LMP_MPI_RANKS=${NUMBER_OF_PROCESS:-16}
LMP_MPI_PERNODE=${PROCESS_PER_NODE:-${LMP_MPI_RANKS}}
LMP_OMP_THREADS=${OMP_NUM_THREADS:-1}
LMP_TILES=${TILES:-1}
LMP_GPUS=${LMP_TILES}
TESTCASE=${TESTCASE:-lc}

docker run -it --rm --device /dev/dri:/dev/dri \
--ipc=host -v /dev/dri/by-path:/dev/dri/by-path \
--workdir /lammps \
--env TESTCASE=${TESTCASE} \
--env NUMBER_OF_PROCESS=${LMP_MPI_RANKS} \
--env PROCESS_PER_NODE=${LMP_MPI_THREADS} \
--env TILES=${LMP_TILES} \
lammps2023 \
/bin/bash lammps_run.sh

