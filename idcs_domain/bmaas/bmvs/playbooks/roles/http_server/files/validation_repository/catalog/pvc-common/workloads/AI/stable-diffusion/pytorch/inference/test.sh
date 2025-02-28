#!/bin/bash

for model in CompVis/stable-diffusion-v1-4 \
	runwayml/stable-diffusion-v1-5 \
	stabilityai/stable-diffusion-2-1-base \
	stabilityai/stable-diffusion-2-1 \
	stabilityai/stable-diffusion-xl-base-1.0 ; do
echo $model
python  stable_diffusion_inference.py -m $model
done


python stable_diffusion_xl_base+refiner.py

