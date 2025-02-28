#!/bin/bash

# Install habana-deepspeed
# Needed if running a multi-device quantization
pip install git+https://github.com/HabanaAI/DeepSpeed.git@1.18.0

# Keep the container alive
sleep infinity