from diffusers import StableDiffusionXLPipeline, StableDiffusionXLImg2ImgPipeline, DiffusionPipeline, StableDiffusionPipeline, DPMSolverMultistepScheduler
from diffusers.utils import load_image
import torch
import time
import numpy as np
import intel_extension_for_pytorch as ipex

# check Intel GPU
print(ipex.xpu.get_device_name(0))

prompt = "photograph of an astronaut riding a horse"
datatype = torch.float16
variant = "fp16"
n_steps = 50
high_noise_frac = 0.8

num_iter = 3
num_warmup = 1

base = DiffusionPipeline.from_pretrained("stabilityai/stable-diffusion-xl-base-1.0", \
        torch_dtype=datatype, use_safetensors=True, variant=variant)
base.scheduler = DPMSolverMultistepScheduler.from_config(base.scheduler.config)
#base.set_progress_bar_config(disable=True)

base.to('xpu')
base.unet = torch.xpu.optimize(base.unet.eval(), dtype=datatype, inplace=True)
base.vae = torch.xpu.optimize(base.vae.eval(), dtype=datatype, inplace=True)
base.text_encoder = torch.xpu.optimize(base.text_encoder.eval(), dtype=datatype, inplace=True)

refiner = DiffusionPipeline.from_pretrained("stabilityai/stable-diffusion-xl-refiner-1.0",  \
        text_encoder_2=base.text_encoder_2, vae=base.vae, \
        torch_dtype=datatype, use_safetensors=True, variant=variant)
refiner.scheduler = DPMSolverMultistepScheduler.from_config(refiner.scheduler.config)
#base.set_progress_bar_config(disable=True)

refiner.to('xpu')
refiner.unet = torch.xpu.optimize(refiner.unet.eval(), dtype=datatype, inplace=True)
refiner.vae = torch.xpu.optimize(refiner.vae.eval(), dtype=datatype, inplace=True)
refiner.text_encoder_2 = torch.xpu.optimize(refiner.text_encoder_2.eval(), dtype=datatype, inplace=True)
print("---- use IPEX optimized model")

#generator = torch.xpu.Generator().manual_seed(0)
generator = torch.Generator(device="xpu").manual_seed(0)

# Warm up
print("Warm up......")
for i in range(num_warmup):
    image = base(
        prompt=prompt,
        num_inference_steps=n_steps,
        denoising_end=high_noise_frac,
        output_type="latent",
        generator = generator
    ).images
    image = refiner(
        prompt=prompt,
        num_inference_steps=n_steps,
        denoising_start=high_noise_frac,
        image=image,
        generator = generator
    ).images[0]

# Inerence
print("Inference......")
total_time = 0
for i in range(num_iter):
    start_time = time.time()
    image = base(
        prompt=prompt,
        num_inference_steps=n_steps,
        denoising_end=high_noise_frac,
        output_type="latent",
        generator = generator
    ).images
    image = refiner(
        prompt=prompt,
        num_inference_steps=n_steps,
        denoising_start=high_noise_frac,
        image=image,
        generator = generator
    ).images[0]

    duration = time.time() - start_time
    total_time += duration
    print(f'iteration {i}: {duration:2.3f} sec')

latency = total_time/num_iter
throughput = 1/latency
print(f'Average end to end Inference Latency: {latency:.3f} seconds')
print(f'Average end to end Inference Throughput: {throughput:.3f} samples/second')

image.save(f'sdxl_base_refiner_fp16.png')

