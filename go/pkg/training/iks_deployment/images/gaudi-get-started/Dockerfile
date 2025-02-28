FROM icir.cps.intel.com/idc-training/trainings/gaudi-base:0.1.0@sha256:e94a86cdf772195b255c94d8cce59167bdef247db2e8dffb127ef1fe5a0ae7ae

ARG PT_VERSION
ARG VERSION
ARG REVISION
ARG BASE_NAME
ARG ARTIFACTORY_URL

USER root

RUN apt-get update && apt-get install -y --no-install-recommends \
        curl \
        libcurl4 \
        moreutils \
        iproute2 \
        libhdf5-dev \
        libnuma-dev \
        libjpeg-dev \
        liblapack-dev \
        libopenblas-dev \
        numactl \
        pdsh \
        libgoogle-perftools-dev && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

USER ${NB_UID}

# RUN bash -c "\
#         case $BASE_NAME in \
#             *ubuntu22.04*) \
#                update-alternatives --install /usr/bin/python python /usr/bin/python3.10 1 \
#             ;; \
#         esac"

# Download wheel files
ARG pt_package_name="pytorch_modules-v${PT_VERSION}_${VERSION}_${REVISION}.tgz"
ARG pt_artifact_path="https://${ARTIFACTORY_URL}/artifactory/gaudi-pt-modules/${VERSION}/${REVISION}/pytorch/ubuntu2204"
ARG tmp_path=/tmp/installations
RUN mkdir "${tmp_path}"

RUN wget --no-verbose \
    "https://${ARTIFACTORY_URL}/artifactory/gaudi-pt-modules/${VERSION}/${REVISION}/pytorch/ubuntu2204/pytorch_modules-v${PT_VERSION}_${VERSION}_${REVISION}.tgz"
RUN tar -xf "pytorch_modules-v${PT_VERSION}_${VERSION}_${REVISION}.tgz" -C "${tmp_path}"/.

# Overwrite the install.sh that comes with it
COPY ./install.sh /tmp/installations/install.sh

# COPY ./installations/ /tmp/installations/

RUN /tmp/installations/install.sh ${VERSION} ${REVISION} && echo "source /etc/profile.d/habanalabs.sh" >> ~/.bashrc

USER root
RUN rm -rf /tmp/installations/ "${pt_package_name}" && /sbin/ldconfig
USER ${NB_UID}

# COPY install_packages.sh .

# RUN ./install_packages.sh && rm -f install_packages.sh && \
#     /sbin/ldconfig && echo "source /etc/profile.d/habanalabs.sh" >> ~/.bashrc

ENV LD_PRELOAD=/lib/x86_64-linux-gnu/libtcmalloc.so.4
ENV TCMALLOC_LARGE_ALLOC_REPORT_THRESHOLD=7516192768
