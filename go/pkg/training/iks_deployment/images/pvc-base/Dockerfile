FROM icir.cps.intel.com/idc-training/trainings/training-base:0.1.0@sha256:747e7986e48b9679c88dab984e38cbaa8fcde108a22657e8a9585f6555ca20af

USER root

ARG env_name=pvc
ARG py_ver=3.12

COPY training-files /tmp
COPY shell.sh /usr/local/bin/before-notebook.d

# Install GPU drivers
RUN apt update && apt install -y gpg-agent wget
RUN apt install gnupg -y
RUN wget -qO - https://repositories.intel.com/gpu/intel-graphics.key | gpg --yes --dearmor --output /usr/share/keyrings/intel-graphics.gpg
RUN . /etc/os-release && \
    if [[ ! " jammy " =~ " ${VERSION_CODENAME} " ]]; then \
        echo "Ubuntu version ${VERSION_CODENAME} not supported"; \
    else \
        wget -qO - https://repositories.intel.com/gpu/intel-graphics.key | \
        gpg --yes --dearmor --output /usr/share/keyrings/intel-graphics.gpg && \
        echo "deb [arch=amd64 signed-by=/usr/share/keyrings/intel-graphics.gpg] https://repositories.intel.com/gpu/ubuntu ${VERSION_CODENAME}/lts/2350 unified" | \
        tee /etc/apt/sources.list.d/intel-gpu-${VERSION_CODENAME}.list && \
        apt update && \
        apt install -y \
            linux-headers-5.15.0-73-generic \
            linux-modules-extra-5.15.0-73-generic \
            flex bison \
            intel-fw-gpu xpu-smi && \
        # apt install -y intel-i915-dkms && \
        apt install -y \
            intel-opencl-icd intel-level-zero-gpu level-zero \
            intel-media-va-driver-non-free libmfx1 libmfxgen1 libvpl2 \
            libegl-mesa0 libegl1-mesa libegl1-mesa-dev libgbm1 libgl1-mesa-dev libgl1-mesa-dri \
            libglapi-mesa libgles2-mesa-dev libglx-mesa0 libigdgmm12 libxatracker2 mesa-va-drivers \
            mesa-vdpau-drivers mesa-vulkan-drivers va-driver-all vainfo hwinfo clinfo && \
        apt-get install -y --no-install-recommends \
            libigc-dev intel-igc-cm libigdfcl-dev libigfxcmrt-dev level-zero-dev; \
    fi


# Add user to the same group as the GPU cards
RUN usermod -aG render ${NB_USER}
RUN groupadd -g 110 graphics
RUN usermod -aG graphics ${NB_USER}

# Create the environment with mamba
RUN mamba create --yes -p "${CONDA_DIR}/envs/${env_name}" \
    python=${py_ver} \
    'ipykernel' \
    'jupyterlab' \
    'tqdm==4.67.1' \
    'Pillow==11.1.0' && \
    mamba clean --all -f -y

# Install PyTorch and related packages using pip with the specified index URL
RUN "${CONDA_DIR}/envs/${env_name}/bin/pip" install \
    --index-url https://download.pytorch.org/whl/test/xpu \
    'torch==2.5.0' \
    'torchaudio==2.5.0' \
    'torchvision==0.20.0'

# Set up the Jupyter kernel for the new environment
RUN "${CONDA_DIR}/envs/${env_name}/bin/python" -m ipykernel install --user \
    --name="${env_name}" --display-name "PyTorch 2.5" && \
    fix-permissions "${CONDA_DIR}" && \
    fix-permissions "/home/${NB_USER}"


# Creating the oneAPI installations
ARG ONEAPI_INSTALLER_URL='https://registrationcenter-download.intel.com/akdlm/IRC_NAS/96aa5993-5b22-4a9b-91ab-da679f422594/intel-oneapi-base-toolkit-2025.0.0.885_offline.sh'
ENV DEBIAN_FRONTEND=noninteractive

# ============================================
# Intel® oneAPI Base Toolkit: Offline Installer
# https://www.intel.com/content/www/us/en/docs/oneapi-base-toolkit/get-started-guide-linux/2025-0/overview.html
# https://www.intel.com/content/www/us/en/developer/tools/oneapi/base-toolkit-download.html?packages=oneapi-toolkit&oneapi-toolkit-os=linux&oneapi-lin=offline
# ============================================

# Download and install offline oneAPI package  
RUN wget --quiet --output-document oneapi.sh ${ONEAPI_INSTALLER_URL} && \
    sh ./oneapi.sh -a --silent --cli --eula accept && \
    rm oneapi.sh && \
    rm -rf intel

# ============================================
# Configure System
# https://www.intel.com/content/www/us/en/docs/oneapi-base-toolkit/get-started-guide-linux/2025-0/before-you-begin.html
# ============================================

# 1. Install CMake*, pkg-config, and GNU* Dev Tools to Build Samples
RUN apt-get update && \
    apt-get install --yes --no-install-recommends \
    build-essential \
    cmake \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# 2. Set Environment Variables for CLI Development
# Extends the default CMD behavior by first running the oneAPI environment variable script then continuing to the Jupyter startup
CMD ["/bin/bash", "-c", "source /opt/intel/oneapi/setvars.sh; /usr/local/bin/start-notebook.py"]

USER ${NB_USER}