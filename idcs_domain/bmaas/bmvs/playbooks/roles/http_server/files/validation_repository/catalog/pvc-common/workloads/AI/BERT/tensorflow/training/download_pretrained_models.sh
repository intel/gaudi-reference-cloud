# Download Pretrained Models
if [ ! -f wwm_uncased_L-24_H-1024_A-16/bert_model.ckpt.data-00000-of-00001 ]; then
   wget https://storage.googleapis.com/bert_models/2019_05_30/wwm_uncased_L-24_H-1024_A-16.zip
   unzip wwm_uncased_L-24_H-1024_A-16.zip
else
   echo "Pretrained models already existed"
fi

