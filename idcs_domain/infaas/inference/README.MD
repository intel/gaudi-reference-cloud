# MaaS - Model as a Service 🧠
This repo has the main logic for the MaaS inference service proxy, a python gRPC server to handle traffic to and from an LLM server.

## 📋 Repo components:

| Component         | Description                                                          |
| ----------------- | -------------------------------------------------------------------- |
| [src](https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/tree/main/idcs_domain/infaas/inference/src) | main logic, has the gRPC files. protos, Dockerfile, etc.|
| [scripts](https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/tree/main/idcs_domain/infaas/inference/scripts) | standalone running scripts, for example HuggingFace to S3 model copy. |
| [quantization](https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/tree/main/idcs_domain/infaas/inference/quantization) | standalone Dockerfile, scripts and insturction to prefom a model FP8 quantization on Gaudi. |
| [configs](https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/tree/main/idcs_domain/infaas/inference/configs) | internal configs, for example server's dev config. |
| [benchmarks](https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/tree/main/idcs_domain/infaas/inference/benchmarks) | benchmarks: standalone Dockerfile, k8s pod file & test logic to measure server's performance. |
| [tests](https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/tree/main/idcs_domain/infaas/inference/tests) | pytest scripts for unit-testing. |

## ▶️ Getting started:
1. Install ```pipx```:
```bash
brew install pipx
```

2. Using ```pipx``` install ```poetry``` (with export plugin)
```bash
pipx install poetry
```

3. Restart the terminal and run:
```bash
poetry self add poetry-plugin-export
```

4. Navigate to poetry's root:
```bash
cd idcs_domain/infaas/inference
```

5. Run poetry install:
```bash
poetry install
```

* Before moving on to next sections, please make sure that `poetry install` is finished without installation errors.

6. Make sure Intel VPN is on and export proxies:
```bash
export HTTP_PROXY=http://proxy-dmz.intel.com:912                 
export HTTPS_PROXY=http://proxy-dmz.intel.com:912
export NO_PROXY=intel.com,.intel.com,10.0.0.0/8,192.168.0.0/16,localhost,127.0.0.0/8,134.134.0.0/16,172.16.0.0/16,192.168.150.0/24,.kind.local
```

7. Export HuggingFace token as ```HF_TOKEN``` (use presonal / get one from MaaS team):
```bash
export HF_TOKEN=<token>
```

8. Run unit-tests to make sure evertyhing is working:
```bash
poetry run pytest
```

## ⚙️ Set poerty as virtual environment (venv):

1. Make sure that poetry has an active venv:
```bash
poetry env info
```

* Result should look like the following:
```bash
Virtualenv
Python:         3.12.8
Implementation: CPython
Path:           /Users/user/Library/Caches/pypoetry/virtualenvs/inference-6Rb_cK8F-py3.12
Executable:     /Users/user/Library/Caches/pypoetry/virtualenvs/inference-6Rb_cK8F-py3.12/bin/python
Valid:          True

Base
Platform:   darwin
OS:         posix
Python:     3.12.8
Path:       /opt/homebrew/opt/python@3.12/Frameworks/Python.framework/Versions/3.12
Executable: /opt/homebrew/opt/python@3.12/Frameworks/Python.framework/Versions/3.12/bin/python3.12
```

* If your result looks like the following please continue to read the `*` steps, if your result is valid you can skip to `2`.
```bash
Virtualenv
Python:         3.12.8
Implementation: CPython
Path:           NA
Executable:     NA
```

* Run the following command using the python executable path from the valid `Base` section example:
```bash
poetry env use <Base Executable>
```

* Run again the info command and make sure that the result is now valid.

2. Activate poetry's venv:
```bash
source $(poetry env info --path)/bin/activate
```

3. Get venv python dir to set as python interpreter (OPTIONAL - relevant only if you want to run `py` files directly using IDE's interpreter):
```bash
poetry env info --path
```

* Setting the python intepreter is different in different IDEs, in all of them you need to open the environment setting manager and enter the previous section path.
