# StableDiffusion 2.1 Gradio Demo WebUI on Intel Data Center GPU
The demo is based on the [stable-diffusion-demo](https://huggingface.co/spaces/stabilityai/stable-diffusion) project, updated to support for Intel Data Center GPU.   
This demo is ONLY for testing purpose.   

## Build the Docker image
Refer the steps in setup.sh. Once the image built, you can find the docker image openxla-sd-demo generated.
```
bash setup.sh
```

## Launch the Demo Instance
Once the instance launched, you will get the webUI link for the demo. By default, it is https://<your host ip>:8080    
Please update the run.sh to pass the right GPU device to the container. The number of GPU titles passed into the container will be the number of the images generated.   
```
bash run.sh
```

## Run the demo
On a system with Chrome browser, open the link https://<your host ip>:8080 and enjoy.

