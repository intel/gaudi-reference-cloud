FROM vault.habana.ai/gaudi-docker/1.14.0/ubuntu22.04/habanalabs/pytorch-installer-2.1.1:1.14.0-493

RUN apt-get update && apt-get install -y pdsh tmux htop

# set python environment variable
ENV PYTHON /usr/bin/python3.8

# upgrade pip
RUN $PYTHON -m pip install --upgrade pip

# Install packages
RUN $PYTHON -m pip install git+https://github.com/HabanaAI/DeepSpeed.git@1.14.0
RUN $PYTHON -m pip install optimum[habana]

# Docker ssh port setup
RUN sed -i 's/#Port 22/Port 3022/g' /etc/ssh/sshd_config && \
   sed -i 's/#   Port 22/    Port 3022/g' /etc/ssh/ssh_config && \
   sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
   service ssh restart

COPY /tmp/run_load.sh /root/run_load.sh
COPY /tmp/.deepspeed_env /root/.deepspeed_env

ENTRYPOINT service ssh restart && /bin/bash