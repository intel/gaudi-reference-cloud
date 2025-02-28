#!/bin/bash

# set python environment variable
PYTHON=/usr/bin/python3.10

cd /optimum-habana/examples/contrastive-image-text

python ../gaudi_spawn.py --use_mpi run_bridgetower.py \
--output_dir /tmp/bridgetower-test \
--model_name_or_path BridgeTower/bridgetower-large-itm-mlm-itc \
--dataset_name jmhessel/newyorker_caption_contest --dataset_config_name matching \
--dataset_revision 3c6c4f6c0ff7e902833d3afa5f8f3875c2b036e6 \
--image_column image --caption_column image_description \
--remove_unused_columns=False \
--do_train --do_eval --do_predict \
--per_device_train_batch_size="40" --per_device_eval_batch_size="16" \
--num_train_epochs 5 \
--learning_rate="1e-5" \
--overwrite_output_dir \
--save_strategy no \
--use_habana --use_lazy_mode --use_hpu_graphs_for_inference --gaudi_config_name Habana/clip \
--throughput_warmup_steps 3 \
--logging_steps 10 \
--dataloader_num_workers 1 \
--mediapipe_dataloader \
--bf16