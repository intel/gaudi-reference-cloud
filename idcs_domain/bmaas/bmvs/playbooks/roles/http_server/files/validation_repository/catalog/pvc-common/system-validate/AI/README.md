<!-- TOC -->

- [Validate with AI Workload for Intel Data Center GPU Max Series PVC](#validate-with-ai-workload-for-intel-data-center-gpu-max-series-pvc)
    - [Tensorflow](#tensorflow)
        - [Resnet50 v1.5 Training](#resnet50-v15-training)
        - [Resnet50 v1.5 Inference](#resnet50-v15-inference)
        - [BertLarge Training](#bertlarge-training)
        - [BertLarge Inference](#bertlarge-inference)
    - [PyTorch](#pytorch)
        - [Resnet50 v1.5 Training](#resnet50-v15-training)
        - [Resnet50 v1.5 Inference](#resnet50-v15-inference)
        - [BertLarge Training](#bertlarge-training)
        - [BertLarge Inference](#bertlarge-inference)

<!-- /TOC -->

# Validate with AI Workload for Intel Data Center GPU Max Series (PVC)
The scripts provide the automated tests based on Intel Extension for Pytorch and Intel Extension for Tensorflow, using Resnet50, BertLarge with both training and inference.   
For the detailed instructions on running these workload on PVC system, please go to corresponding folder in [workload/AI](../../workloads/AI/).
The scripts are designed for test and validation with predefined configuration and expected performance data. The different result is expected if the test is done on a different system configuration.
The test can support conda env, python venv or docker env. It is recommended to use conda env to run the test. Please make sure you have conda installed.
```bash
# make sure the conda binary can be found in the current bash shell before you run the test
which conda
```
With a successful test, two folders are created: the log files are saved in the bench*_logs folder and the working folder in folder bench*_workdir. A csv file with test result is generated in the workdir for offline reference.   
use DRYRUN=1 bash \<script name\>.sh to print the command for each test and you can go to bench*_workdir to run the same test command separately if needed.   

## Tensorflow
The tests are based on [Intel Extension for Tensorflow v2.14.0.1](https://github.com/intel/intel-extension-for-tensorflow/releases/tag/v2.14.0.1). 
### Resnet50 v1.5 Training
- DataSet: The test uses synthetic ImageNet2012 dataset for validation purpose. The real Imagenet2012 dataset not needed for this test.   
- Batch Size: 256
- Data Type: bfloat16
- EPOCHS: 4
- Number of GPU to test: 1 2 4 8
```bash
bash sys_val_tf_resnet50_training.sh
```

### Resnet50 v1.5 Inference
- DataSet: The test uses synthetic ImageNet2012 dataset for validation purpose. The real Imagenet2012 dataset not needed for this test.   
- Batch Size: 1024
- Data Type: INT8
- Number of GPU to test: 1 2 4 8
```bash
bash sys_val_tf_resnet50_inference.sh
```

### BertLarge Training
- DataSet: Dummy dataset.   
- Batch Size: 32
- Data Type: bfloat16
- Number of GPU to test: 1 2 4 8
```bash
bash sys_val_tf_bertlarge_training.sh
```

### BertLarge Inference
- DataSet: SQuAD1.0   
- Batch Size: 64
- Data Type: float16
- Number of GPU to test: 1 2 4 8
```bash
bash sys_val_tf_bertlarge_inference.sh
```

## PyTorch
The tests are based on [Intel Extension for PyTorch v2.1.10](https://github.com/intel/intel-extension-for-pytorch/releases/tag/v2.1.10+xpu).
### Resnet50 v1.5 Training
- DataSet: The test uses synthetic ImageNet2012 dataset for validation purpose. The real Imagenet2012 dataset not needed for this test.   
- Batch Size: 256
- Data Type: bfloat16
- EPOCHS: 2
- Number of GPU to test: 1 2 4 8
```bash
bash sys_val_pt_resnet50_training.sh
```

### Resnet50 v1.5 Inference
- DataSet: The test uses synthetic ImageNet2012 dataset for validation purpose. The real Imagenet2012 dataset not needed for this test.   
- Batch Size: 1024
- Data Type: INT8
- Number of GPU to test: 1 2 4 8
```bash
bash sys_val_pt_resnet50_inference.sh
```

### BertLarge Training
- DataSet: [MLCommons BERT](https://drive.google.com/drive/folders/1cywmDnAsrP5-2vsr8GDc6QUc7VWe-M3v).   
- Batch Size: 32
- Data Type: bfloat16
- Number of GPU to test: 1 2 4 8
Please prepare the dataset based on [instructions](https://github.com/IntelAI/models/blob/master/quickstart/language_modeling/pytorch/bert_large/training/gpu/DEVCATALOG.md#datasets). Use env DATASET_DIR and PROCESSED_DATASET_DIR to set the original downloaded dataset and the processed dataset. Create the empty folder if you want the scripts to download the dataset and process it automatically. Please make sure you have enough disk space to store the processed dataset, which is around 538G for processed hdf5 files.
```bash
DATASET_DIR=<path to dataset dir> PROCESSED_DATASET_DIR=<path to processed dataset dir> bash sys_val_pt_bertlarge_training.sh
```

### BertLarge Inference
- DataSet: SQuAD1.0   
- Batch Size: 64
- Data Type: float16
- Number of GPU to test: 1 2 4 8
```bash
bash sys_val_pt_bertlarge_inference.sh
```
