## Stable Diffusion inference on Intel Data Center GPU Max Series

### Run Stable Diffusion in Bare Metal
#### Create a conda or python venv environment
* If you have conda/miniconda installed
```
conda create -y -n py_sv python=3.9
conda activate py_sv
```

* or use python venv to create a virtual environment
```
sudo apt install python3-virtualenv
python3 -m venv py_sv
source py_sv/bin/activate
```

#### Install requirements and setup inference environment in one time
#### Make sure correct oneAPI is installed (oneAPI 2024.0 is needed for ipex 2.1.10)
```
bash setup.sh
```

#### Run the Stable Diffusion inference
```
source /opt/intel/oneapi/setvars.sh
python  stable_diffusion_inference.py
```
