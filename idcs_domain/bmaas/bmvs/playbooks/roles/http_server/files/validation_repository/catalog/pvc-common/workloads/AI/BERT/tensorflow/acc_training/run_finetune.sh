NUM_GPUS=1
BATCH_SIZE_PER_GPU=12
LEARNING_RATE_PER_GPU=5e-6
DATATYPE=$DATATYPE
USE_XLA=false
SQUAD_VERSION=1.1
EPOCHS=2
USE_MYTRAIN=true
BERT_MODEL=large
#PRETRAIN_PATH=./results/tf_bert_pretraining_lamb_large_bf16_gbs1_7200_gbs2_1024/phase_2/pretrained/bert_model.ckpt-1
PRETRAIN_PATH=$PRETRAIN_RESULT_DIR/phase_2/pretrained/bert_model.ckpt-1
DATA_DIR=./data
RESULT_DIR=./results/tf_bert_finetune_${BERT_MODEL}_${$DATATYPE}
bash scripts/run_squad.sh \
    $NUM_GPUS \
    $BATCH_SIZE_PER_GPU \
    $LEARNING_RATE_PER_GPU \
    $DATATYPE \
    $USE_XLA \
    $BERT_MODEL \
    $SQUAD_VERSION \
    $EPOCHS \
    $USE_MYTRAIN \
    $PRETRAIN_PATH \
    $DATA_DIR \
    $RESULT_DIR \
    |& tee finetune.log
