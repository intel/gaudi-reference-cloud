#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
ONEAPI_ROOT=${ONEAPI_ROOT:-/opt/intel/oneapi}
APPS_ROOT=${SCRIPT_DIR}/apps
LAMMPS_ROOT=${APPS_ROOT}/lammps

source ${ONEAPI_ROOT}/setvars.sh

function res () {
    if [ $? -eq 0 ]
    then
        echo -e "\033[32m $@ sucessed. \033[0m"
    else
        echo -e "\033[41;37m $@ failed. \033[0m"
        exit
    fi
}

function download_source() {
    mkdir -p ${APPS_ROOT} ; cd ${APPS_ROOT}
    echo "In folder `pwd`"
    if [ -e "cal/README.md" ]; then
	echo "The compute-aggregation-layer folder already exists. Reuse the existing cal source to build"
	echo "Remove the folder ${APPS_ROOT}/cal to have a fresh download"
	#rm -rf cal
    else
        git clone https://github.com/intel/compute-aggregation-layer.git cal
	res "Compute Aggregation Layer source code download"
    fi

    if [ -e "lammps/README" ]; then
	echo "The Lammps already exists. Reuse the existing lammps source to build"
	echo "Remove the folder ${APPS_ROOT}/lammps to have a fresh download"
	#rm -rf lammps     
    else
        git clone https://www.github.com/lammps/lammps lammps -b develop --depth 1
        res "Lammps source code download"
    fi
}

function install_cal() {
    cd ${APPS_ROOT}/cal
    echo "In folder `pwd`"
    if [ -e "build/calrun" ]; then
	echo "Find existing calrun. Skip the cal build"
	echo "remove folder ${APPS_ROOT}/cal/build folder to rebuild"
    else
	mkdir build; cd build
        cmake ../
        make -j
        res "CAL build"
    fi
}

function install_lammps() {
    export PATH=${APPS_ROOT}/cal/build:$PATH
    cd ${LAMMPS_ROOT}/src
    echo "In folder `pwd`"
    if [ -e "lmp_oneapi" ]; then
	echo "Find existing lammps binary ${LAMMPS_ROOT}/src/lmp_oneapi. Skip the lammps build"
	echo "Remove file lmp_oneapi and make clean_all to clean the build"
    else
        make yes-asphere yes-kspace yes-manybody yes-misc
        make yes-molecule yes-rigid yes-dpd-basic yes-gpu
        cd ${LAMMPS_ROOT}/lib/gpu
        make -f Makefile.oneapi -j
        cd ${LAMMPS_ROOT}/src
        make oneapi -j
        res "Lammps build"
	cp -r ${LAMMPS_ROOT}/src/INTEL/TEST ${APPS_ROOT}/
    fi
}

download_source

install_cal

install_lammps

