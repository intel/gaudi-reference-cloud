#! /bin/bash

#docker stop btower_$(hostname)
#docker rm btower_$(hostname)

sudo apt update -y && sudo apt upgrade -y
sudo add-apt-repository ppa:deadsnakes/ppa -y
sudo apt update -y
sudo apt install python3.10 -y

sudo docker build -t btower -f /tmp/Dockerfile . > /tmp/dockeroutput1.txt 2>&1

sudo docker run --env=HABANA_VISIBLE_DEVICES=all --env=OMPI_MCA_btl_vader_single_copy_mechanism=none --cap-add=sys_nice --network=host --restart=no --device=/dev/infiniband --runtime=habana --shm-size=64g -v /home:/home --name btower -t -d btower:latest


sudo docker exec -it btower bash -c "\
	  git clone https://github.com/huggingface/optimum-habana.git /optimum-habana
          pip install --upgrade-strategy eager optimum[habana] && \
          pip install git+https://github.com/HabanaAI/DeepSpeed.git@1.14.0 && \
          pip install git+https://github.com/huggingface/optimum-habana.git && \
          cd /optimum-habana/examples/contrastive-image-text/ && \
          pip install -r requirements.txt && \
          git checkout v1.10.1 && \
          pip install --upgrade transformers==4.37.0 && \
          cd /root/ && \
          sh run_load.sh >/tmp/dockeroutput2.txt 2>&1"