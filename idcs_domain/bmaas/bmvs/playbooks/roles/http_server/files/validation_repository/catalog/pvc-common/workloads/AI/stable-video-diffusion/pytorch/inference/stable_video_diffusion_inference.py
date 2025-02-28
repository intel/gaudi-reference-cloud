import time
import argparse
import torch
import torch.utils.benchmark as benchmark
from diffusers import StableVideoDiffusionPipeline
from diffusers.utils import load_image, export_to_video

def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument('-m', '--model-id',
        type=str,
        choices=["stabilityai/stable-video-diffusion-img2vid", 
            "stabilityai/stable-video-diffusion-img2vid-xt"],
        default='stabilityai/stable-video-diffusion-img2vid',
    )
    parser.add_argument("--batch-size", "--batch_size", default=1, type=int, help="batch size")
    parser.add_argument("--num-iter", default=3, type=int, help="num iter")
    parser.add_argument("--num-warmup", default=1, type=int, help="num warmup")
    parser.add_argument('--num-decode-frames', default=8, type=int, help='number of decode frames')
    parser.add_argument('--seed', default=42, type=int, help='generator seed')
    parser.add_argument('--export-video', action='store_true', default=False, help='export generated video')
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
    parser.add_argument('--input-image', default='rocket.png', type=str, help='the input image for pipeline')

    return parser.parse_args()

args = get_args()
print(args)

if args.device == "xpu":
    import intel_extension_for_pytorch as ipex

if args.dtype == "bfloat16":
    infer_dtype = torch.bfloat16
elif args.dtype == "float16":
    infer_dtype = torch.float16
elif args.dtype == "float32":
    infer_dtype = torch.float32

model_id = args.model_id
print(f"Input Image: {args.input_image}")
print(f"Loading model {model_id}")
print(f"Batch size: {args.batch_size}")

pipe = StableVideoDiffusionPipeline.from_pretrained(
    model_id, torch_dtype=infer_dtype, variant="fp16"
)

pipe = pipe.to(args.device)

# optimize with ipex
if args.device == "xpu":
    pipe.unet = torch.xpu.optimize(pipe.unet.eval(), dtype=infer_dtype, inplace=True)
    pipe.vae = torch.xpu.optimize(pipe.vae.eval(), dtype=infer_dtype, inplace=True)
    pipe.image_encoder = torch.xpu.optimize(pipe.image_encoder.eval(), dtype=infer_dtype, inplace=True)

def elapsed_time(pipeline, num_warmup=1, number_iter=3, num_decode_frames=8, batch_size=1):
    global frames

    # Load the conditioning image
    #image = load_image("https://huggingface.co/datasets/huggingface/documentation-images/resolve/main/diffusers/svd/rocket.png?download=true")
    image = load_image(args.input_image)
    image = image.resize((1024, 576))

    generator = torch.manual_seed(args.seed)
    # warmup
    print(f'warmup {num_warmup} times')
    for _ in range(num_warmup):
        frames = pipeline(image, decode_chunk_size=num_decode_frames, generator=generator).frames[0]

    total_time = 0
    for n in range(number_iter):
        start = time.time()
        frames = pipeline(image, decode_chunk_size=num_decode_frames, generator=generator).frames[0]
        end = time.time()
        duration = end - start
        print(f'infer: {n}, Latency: {duration:.3f} seconds')
        total_time += duration
    
    return total_time / number_iter

def export_output(frames):
    if args.export_video:
            for i in range(args.batch_size):
                file_name = "{}_bs{}_video{}.mp4".format(args.dtype, args.batch_size, i)
                export_to_video(frames, file_name, fps=7)

latency = elapsed_time(pipe, args.num_warmup, number_iter=args.num_iter, num_decode_frames=args.num_decode_frames, batch_size=args.batch_size)
print(f"Input Image: {args.input_image}")
print(f"Data type: {args.dtype}")
print(f"Batch Size: {args.batch_size}")
print(f"Model: {args.model_id}")
print(f"Average Inference Latency: {latency:.3f} seconds")
print(f"Average Inference Throughput: {args.batch_size/latency:.3f} samples/second")

export_output(frames)
