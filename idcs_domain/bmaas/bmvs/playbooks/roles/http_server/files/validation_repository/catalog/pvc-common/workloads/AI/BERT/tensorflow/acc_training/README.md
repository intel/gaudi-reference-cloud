## Accelerate BERT-Large Pretraining on Intel Data Center GPU Max Series


### Prepare model for pretraining

We will follow the ACC BERT training instuction (https://github.com/intel/intel-extension-for-tensorflow/blob/main/examples/pretrain_bert/README.md) to set up BERT-Large pretraining based on nvidia-bert. It was optimized nvidia-bert, for example, using custom kernels, fusing some ops to reduce op number, and adding bf16 mode for the model. 

To get better performance, instead of installing official nvidia-bert, you can clone nvidia-bert, apply the patch, then install it as shown here:

```
git clone https://github.com/NVIDIA/DeepLearningExamples.git
cd DeepLearningExamples/TensorFlow2/LanguageModeling/BERT
git apply itex.patch
```

### Setup Running Environment

* Enable oneAPI Running Envionment

```
source /opt/intel/oneapi/setvars.sh
```

* Creat a python venv envrionment

```
sudo apt install python3.10-venv
python3.10 -m venv acc_tf_bert
source acc_tf_bert/bin/activate
```

* Install requirements and setup the tensorflow training envionment

```bash
./tf_bertlarge_training_setup.sh
```

### Prepare Dataset

nvidia-bert repository provides scripts to download, verify, and extract the SQuAD dataset and pretrained weights for fine-tuning as well as Wikipedia and BookCorpus dataset for pre-training. You can run below to download datasets for fine-tuning and pretraining.

```
bash scripts/data_download.sh all
```

For more details about downloading and processing the dataset, you can reference [downloading](https://github.com/NVIDIA/DeepLearningExamples/tree/master/TensorFlow2/LanguageModeling/BERT#quick-start-guide) and [processing](https://github.com/NVIDIA/DeepLearningExamples/tree/master/TensorFlow2/LanguageModeling/BERT#getting-the-data) part. After downloading and processing, the datasets are supposed in the following locations by default

- SQuAD v1.1 - `data/download/squad/v1.1`
- SQuAD v2.0 - `data/download/squad/v2.0`
- BERT-Large - `data/download/google_pretrained_weights/uncased_L-24_H-1024_A-16`
- BERT-Base - `data/download/google_pretrained_weights/uncased_L-12_H-768_A-12`
- Wikipedia + BookCorpus TFRecords - `data/tfrecords/books_wiki_en_corpus`

## Execute the Example

Bert pretraining is very time-consuming, as nvidia-bert repository says, training BERT-Large from scratch on 16 V100 using FP16 datatype takes around 4.5 days. So Here we only provide single-tile pretraining scripts within a day to show performance.

#### Pretraining Command

Assume current_dir is `examples/pretrain_bert/DeepLearningExamples/TensorFlow2/LanguageModeling/BERT`

+ BFloat16 DataType

```
DATATYPE=bf16
```

+ Float32 DataType

```
DATATYPE=fp32
```

+ TF32 DataType

```
export ITEX_FP32_MATH_MODE=TF32
DATATYPE=fp32
```

**Run BERT_Large Tensorflow Pretraining**

+ We use [LAMB](https://arxiv.org/pdf/1904.00962.pdf) as the optimizer and pretraining has two phases. The maximum sequence length of phase1 and phase2 is 128 and 512, respectively. For the whole process of pretraining, you can use scripts in [nvidia-bert](https://github.com/NVIDIA/DeepLearningExamples/tree/master/TensorFlow2/LanguageModeling/BERT#training-process).

```
#PWD: working dir 
cd ./DeepLearningExamples/TensorFlow2/LanguageModeling/BERT
bash ./run_lamb.sh
```

#### Finetune Command

Assume current_dir is `examples/pretrain_bert/DeepLearningExamples/TensorFlow2/LanguageModeling/BERT`. After getting the pretraining checkpoint, you can use it for finetuning.

+ BFloat16 DataType

```
DATATYPE=bf16
```

+ Float32 DataType

```
DATATYPE=fp32
```

+ TF32 DataType

```
export ITEX_FP32_MATH_MODE=TF32
DATATYPE=fp32
```

**Run BERT_Large Tensorflow Finetuning **

```
#PWD: working dir 
cd ./DeepLearningExamples/TensorFlow2/LanguageModeling/BERT
bash ./run_finetune.sh
```
