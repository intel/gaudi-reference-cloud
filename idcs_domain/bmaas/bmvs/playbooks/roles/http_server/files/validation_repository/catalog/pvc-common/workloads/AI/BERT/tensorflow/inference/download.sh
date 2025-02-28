# Download Pretrained Models
echo "Download pretrained models"
if [ ! -f wwm_uncased_L-24_H-1024_A-16/bert_model.ckpt.data-00000-of-00001 ]; then
    wget https://storage.googleapis.com/bert_models/2019_05_30/wwm_uncased_L-24_H-1024_A-16.zip
    unzip wwm_uncased_L-24_H-1024_A-16.zip
fi

if [ ! -f wwm_uncased_L-24_H-1024_A-16/dev-v1.1.json ]; then
    wget https://rajpurkar.github.io/SQuAD-explorer/dataset/dev-v1.1.json -P wwm_uncased_L-24_H-1024_A-16
fi

# Download frozen graph model

if [ ! -f fp32_bert_squad.pb ]; then
    wget https://storage.googleapis.com/intel-optimized-tensorflow/models/v2_7_0/fp32_bert_squad.pb
fi

# Download dataset
mkdir -p SQuAD1.0
echo "Download dataset, SQuAD1.0 for Bert large inference..."
cd SQuAD1.0
if [ ! -f train-v1.1.json ]; then
    wget https://rajpurkar.github.io/SQuAD-explorer/dataset/train-v1.1.json
    wget https://rajpurkar.github.io/SQuAD-explorer/dataset/dev-v1.1.json
    #wget https://github.com/allenai/bi-att-flow/blob/master/squad/evaluate-v1.1.py
    wget https://raw.githubusercontent.com/allenai/bi-att-flow/master/squad/evaluate-v1.1.py
fi
