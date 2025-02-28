# Use the base image
FROM vault.habana.ai/gaudi-docker/1.13.0/ubuntu20.04/habanalabs/pytorch-installer-2.1.0:latest

# Set environment variables
ENV HABANA_VISIBLE_DEVICES=all
ENV OMPI_MCA_btl_vader_single_copy_mechanism=none

# Add capability
RUN apt-get update && apt-get install -y libcap2
RUN setcap cap_sys_nice=ep /usr/bin/python3.8

# Clone the repository and navigate to the directory
RUN git clone https://github.com/HabanaAI/Model-References.git && \
    cd Model-References/PyTorch/examples/computer_vision/hello_world/

# Set PYTHONPATH
ENV PYTHONPATH=$PYTHONPATH:/Model-References

# Set Python executable path
ENV PYTHON=/usr/bin/python3.8

# Define the command to execute the training load
CMD ["sh", "-c", "export PYTHONPATH=$PYTHONPATH:/Model-References && \
                   export PYTHON=/usr/bin/python3.8 && \
                   $PYTHON /Model-References/PyTorch/examples/computer_vision/hello_world/mnist.py --batch-size=64 --epochs=18 --lr=1.0 --gamma=0.7 --hpu --autocast"]
