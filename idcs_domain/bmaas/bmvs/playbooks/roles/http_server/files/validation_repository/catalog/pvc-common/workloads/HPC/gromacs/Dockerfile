FROM intel/oneapi-basekit:latest as build


FROM intel/oneapi-basekit:latest

RUN apt-get update && apt-get upgrade -y && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    ca-certificates cmake curl unzip python3-pip python3-mako vim clinfo intel-opencl-icd && \
    mkdir -p /gromacs
  
COPY gromacs_setup.sh /gromacs
COPY gromacs_run.sh /gromacs
ADD testcases /gromacs/testcases

WORKDIR /gromacs
RUN ./gromacs_setup.sh
