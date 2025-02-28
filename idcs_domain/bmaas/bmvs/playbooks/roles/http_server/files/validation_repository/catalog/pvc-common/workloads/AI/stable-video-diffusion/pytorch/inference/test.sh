#!/bin/bash

for model in stabilityai/stable-video-diffusion-img2vid \
        stabilityai/stable-video-diffusion-img2vid-xt; do
echo $model
python stable_video_diffusion_inference.py -m $model
done

