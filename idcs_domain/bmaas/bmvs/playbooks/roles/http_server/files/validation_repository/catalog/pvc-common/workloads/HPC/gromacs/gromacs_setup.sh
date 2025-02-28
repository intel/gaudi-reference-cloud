#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
ONEAPI_ROOT=${ONEAPI_ROOT:-/opt/intel/oneapi}
APPS_ROOT=${SCRIPT_DIR}/apps
GMX_ROOT=${APPS_ROOT}/gromacs
GMX_INSTALL=${APPS_ROOT}/__install
TESTCASE_DIR=${SCRIPT_DIR}/testcases
TESTCASE=${TESTCASE:-1}
MULTI_NODES=${MULTI_NODES:-0}

GMX_BUILD_OPTION="-DCMAKE_C_COMPILER=icx -DCMAKE_CXX_COMPILER=icpx -DCMAKE_INSTALL_PREFIX=$GMX_INSTALL -DGMXAPI=OFF -DGMX_GPU=SYCL -DGMX_FFT_LIBRARY=mkl -DGMX_GPU_NB_CLUSTER_SIZE=8 -DGMX_GPU_NB_NUM_CLUSTER_PER_CELL_X=1"
if (( $MULTI_NODES ));then
    GMX_BUILD_OPTION="$GMX_BUILD_OPTION -DGMX_MPI=on"
fi

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
    if [ -e "gromacs/README" ]; then
	echo "$GMX_ROOT folder exist. Reuse the existing gromacs source to build"
	echo "Remove the folder $GMX_ROOT to have a fresh download"
	#rm -rf gromacs
    else
        git clone https://gitlab.com/gromacs/gromacs.git
        res "Gromacs Source code download"
    fi
}

function download_testcases() {
    if (( $TESTCASE )); then
	cd $TESTCASE_DIR
	bash download.sh
    fi
}

function install_gromacs() {
    if [ -e $GMX_INSTALL ]; then
	echo "$GMX_INSTALL folder exist, gromacs install cancelled. Please remove the folder to rebuild if needed"
	return
    fi
    rm -rf $GMX_ROOT/build
    #rm -rf $GMX_INSTALL

    cd $GMX_ROOT
    mkdir -p build ; cd build
    pwd
    echo "cmake $GMX_BUILD_OPTION .."
    cmake $GMX_BUILD_OPTION .. 2>&1 |tee cmake.log
    make -j 2>&1 |tee make.log

    make install 2>&1 |tee install.log
    res "Gromacs Build"
}


download_source

install_gromacs

download_testcases


