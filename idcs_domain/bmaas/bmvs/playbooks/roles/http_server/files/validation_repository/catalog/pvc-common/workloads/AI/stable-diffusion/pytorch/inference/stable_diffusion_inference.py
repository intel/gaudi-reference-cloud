import time
import argparse
import torch
import torch.utils.benchmark as benchmark
from diffusers import DPMSolverMultistepScheduler, StableDiffusionPipeline, StableDiffusionXLPipeline

import intel_extension_for_pytorch as ipex

def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument('-m', '--model-id',
        type=str,
        choices=["CompVis/stable-diffusion-v1-4", "runwayml/stable-diffusion-v1-5",
            "stabilityai/stable-diffusion-2-1-base", "stabilityai/stable-diffusion-2-1",
            "stabilityai/stable-diffusion-xl-base-1.0"],
        default='stabilityai/stable-diffusion-xl-base-1.0',
    )
    parser.add_argument("--batch-size", "--batch_size", default=1, type=int, help="batch size")
    parser.add_argument("--num-iter", default=3, type=int, help="num iter")
    parser.add_argument("--num-warmup", default=1, type=int, help="num warmup")
    parser.add_argument('--num-steps', default=50, type=int, help='number of diffusion steps')
    parser.add_argument('--save_image', action='store_true', default=False, help='save image')
    parser.add_argument('--device',
        type=str,
        choices=["cpu", "xpu"],
        help="cpu or xpu",
        default='xpu',
    )
    parser.add_argument(
        "--dtype",
        type=str,
        help="bfloat16 or float32 or float16",
        choices=["float32", "bfloat16", "float16"],
        default="float16",
    )
    parser.add_argument('--prompt', default='photograph of an astronaut riding a horse', type=str, help='the text prompt to put into the text encoder')

    return parser.parse_args()

args = get_args()
print(args)

if args.device == "xpu":
    args.ipex = True


if args.dtype == "bfloat16":
    infer_dtype = torch.bfloat16
elif args.dtype == "float16":
    infer_dtype = torch.float16
elif args.dtype == "float32":
    infer_dtype = torch.float32

model_id = args.model_id
prompt = args.prompt
print(f"Prompt: {prompt}")
print(f"Loading model {model_id}")
print(f"Batch size: {args.batch_size}")
print(f"Prompt: {prompt}")

if "xl" in args.model_id.lower():
    pipe = StableDiffusionXLPipeline.from_pretrained(model_id, torch_dtype=infer_dtype, safety_checker=None)
else:
    pipe = StableDiffusionPipeline.from_pretrained(model_id, torch_dtype=infer_dtype, safety_checker=None)
pipe.scheduler = DPMSolverMultistepScheduler.from_config(pipe.scheduler.config)
#pipe.set_progress_bar_config(disable=True)

pipe = pipe.to(args.device)

if args.batch_size >= 32:
    pipe.enable_vae_slicing()

# optimize with ipex
pipe.unet = ipex.optimize(pipe.unet.eval(), dtype=infer_dtype, inplace=True)
pipe.vae = ipex.optimize(pipe.vae.eval(), dtype=infer_dtype, inplace=True)
pipe.text_encoder = ipex.optimize(pipe.text_encoder.eval(), dtype=infer_dtype, inplace=True)

def elapsed_time(pipeline, num_warmup=1, number_iter=10, num_inference_steps=20, batch_size=1):
    global images
    # warmup
    print(f'warmup {num_warmup} times')
    for _ in range(num_warmup):
        images = pipeline(prompt, num_inference_steps=num_inference_steps, num_images_per_prompt=batch_size).images

    total_time = 0
    for n in range(number_iter):
        start = time.time()
        images = pipeline(prompt, num_inference_steps=num_inference_steps, num_images_per_prompt=batch_size).images
        end = time.time()
        duration = end - start
        print(f'infer: {n}, Latency: {duration:.3f} seconds')
        total_time += duration
    
    return total_time / number_iter

def save_images(images):
    if args.save_image:
           for i in range(args.batch_size):
               file_name = "{}_steps{}_bs{}_imgs{}.png".format(args.dtype, args.num_steps, args.batch_size, i)
               images[i].save(file_name)

latency = elapsed_time(pipe, args.num_warmup, number_iter=args.num_iter, num_inference_steps=args.num_steps, batch_size=args.batch_size)
print(f"Prompt: {prompt}")
print(f"Data type: {args.dtype}")
print(f"Batch Size: {args.batch_size}")
print(f"Model: {args.model_id}")
print(f"Average Inference Latency: {latency:.3f} seconds")
print(f"Average Inference Throughput: {args.batch_size/latency:.3f} samples/second")

save_images(images)
