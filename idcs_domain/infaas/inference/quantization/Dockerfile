# General container to build OptimumHabana env to run quantization scripts
FROM vault.habana.ai/gaudi-docker/1.18.0/ubuntu22.04/habanalabs/pytorch-installer-2.4.1:1.18.0-524-20241212

WORKDIR /app

RUN apt-get update && apt-get install -y \
    build-essential \
    python3-dev \
    git \
    cmake \
    libopenmpi-dev

RUN pip install --upgrade pip

RUN pip install git+https://github.com/huggingface/optimum-habana.git@v1.14-release

RUN git clone https://github.com/huggingface/optimum-habana /app/optimum-habana
RUN pip install -r /app/optimum-habana/examples/text-generation/requirements.txt
RUN pip install -r /app/optimum-habana/examples/text-generation/requirements_lm_eval.txt
RUN pip install datasets==2.19.2 sentencepiece

COPY /quantization/startup.sh /app/startup.sh

# Make the script executable
RUN chmod +x /app/startup.sh
CMD ["/app/startup.sh"]