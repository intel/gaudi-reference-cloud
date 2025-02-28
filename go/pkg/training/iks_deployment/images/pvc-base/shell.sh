# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
"${CONDA_DIR}/envs/pvc/bin/python" -m ipykernel install --user \
    --name="pvc" --display-name "PyTorch 2.5" && \
    fix-permissions "${CONDA_DIR}" && \
    fix-permissions "/home/${NB_USER}"