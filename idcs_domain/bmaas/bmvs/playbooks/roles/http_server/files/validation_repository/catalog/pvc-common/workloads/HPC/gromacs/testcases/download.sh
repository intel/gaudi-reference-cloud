#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $SCRIPT_DIR

echo "Downloading benchmarks from https://www.mpinat.mpg.de/grubmueller/bench"

if [ ! -e "benchMEM.tpr" ]; then
    wget -O benchMEM.zip https://www.mpinat.mpg.de/benchMEM
    unzip benchMEM.zip
fi

if [ ! -e "benchPEP.tpr" ]; then
    wget -O benchPEP.zip https://www.mpinat.mpg.de/benchPEP
    unzip benchPEP.zip
fi

if [ ! -e "benchPEP-h.tpr" ]; then
    wget -O benchPEP-h.zip https://www.mpinat.mpg.de/benchPEP-h
    unzip benchPEP-h.zip
fi

echo "Downloading benchmarks from https://zenodo.org/record/3893789"
if [ ! -e "topol.tpr" ]; then
    wget -O GROMACS_heterogeneous_parallelization_benchmark_info_and_systems_JCP.tar.gz https://zenodo.org/record/3893789/files/GROMACS_heterogeneous_parallelization_benchmark_info_and_systems_JCP.tar.gz
    tar xf GROMACS_heterogeneous_parallelization_benchmark_info_and_systems_JCP.tar.gz
    ln -s GROMACS_heterogeneous_parallelization_benchmark_info_and_systems_JCP/stmv/topol.tpr
fi


