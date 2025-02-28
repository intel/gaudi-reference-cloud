#/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
dry_run=${DRYRUN:-0}
DOCKER=${DOCKER:-0}
VENV_NAME=${VENV_NAME:-bench_tf_resnet50_training}
GPUNUM=${GPUNUM:-"1 2 4 8"}

if [ "$DOCKER" != "0" ]; then
    LOG_DIR=${WORKSPACE}/bench_tf_resnet50_training_container_logs
    WORK_DIR=${WORKSPACE}/bench_tf_resnet50_training_container_workdir
else
    LOG_DIR=${WORKSPACE}/bench_tf_resnet50_training_logs
    WORK_DIR=${WORKSPACE}/bench_tf_resnet50_training_workdir
fi
mkdir -p ${LOG_DIR}
mkdir -p ${WORK_DIR}


print_prefix (){
    echo "-----------------------------------------------------------------"
    echo "# $1"
    echo "-----------------------------------------------------------------"
}

print_aftfix(){
    echo "-----------------------------------------------------------------"
    echo ""
}


num_gpu=$( lspci |grep Display |wc -l )

tiles=1
#Max 1550 device id 0bd5
device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
if [ ! -z "$device_type" ]; then
        tiles=2
fi
device_name=$(lspci -vmm |grep Max | head -n 1)

validate_ai_resnet50_tf_training(){
    #Use default test parameters
    PRECISION=bfloat16
    EPOCHS=${EPOCHS:-1}
    BATCH_SIZE=256
    NUMBER_OF_PROCESS=1
    PROCESS_PER_NODE=1
    DATASET_DUMMY=1

    #define threshold and refernce data
    threshold=0.05
    if [ $tiles == 1 ]; then
        #reference data for Max 1100
        num_of_card=(1 2 3 4 5 6 7 8)
        reference_data=(1300 2500 0 4800 0 0 0 9200)
    elif [ $tiles == 2 ]; then
        #reference dat for Max 1550
        num_of_card=(1 2 3 4 5 6 7 8)
        reference_data=(3050 5900 0 11700 0 0 0 23000)  
    fi
    perf_val=(0 0 0 0 0 0 0 0)

    #running in benchmark workdir folder
    test_id="tf_resnet50_training"
    cp -r ${SCRIPT_DIR}/../../workloads/AI/ResNet50_v1.5/tensorflow/training/* ${WORK_DIR}
    bench_script_dir=${WORK_DIR}
    date4log=""
    #$(date '+%Y%m%d_%H%M%S')
    if [ "$DOCKER" != "0" ]; then
        script2run=./tf_resnet50_training_container_run.sh
        log_file_prefix=${LOG_DIR}/tf_resnet50_training_container_run
	script2setup=./tf_resnet50_training_container_setup.sh
	log_file_setup=${LOG_DIR}/tf_resnet50_training_container_setup_${date4log}.log
    else
	script2run=./tf_resnet50_training_run.sh
	log_file_prefix=${LOG_DIR}/tf_resnet50_training_run
	script2setup=./tf_resnet50_training_setup.sh
	log_file_setup=${LOG_DIR}/tf_resnet50_training_setup_${date4log}.log
    fi

    cd $bench_script_dir
    #setup environments
    print_prefix "AI benchmark evaluation - Tensorflow Resnet50 v1.5 Training on $device_name"
    echo "### Set up working environment ..."
    if (( ! $dry_run ));then
        eval "VENV_NAME=${VENV_NAME} $script2setup" 2>&1 |tee $log_file_setup
    else
        echo "VENV_NAME=${VENV_NAME} $script2setup" 
    fi

    for i in ${GPUNUM}; do
      if (( $( echo "$num_gpu >= $i" | bc -l ) )); then
        echo "### Running benchmarks on $i GPU card"
        num_process=$( bc -l <<< "$tiles*$i" )
        log_file=${log_file_prefix}_${i}c${num_process}t.log
        epochs=1
        if [ "$num_process" -ge "8" ]; then
            epochs=4
        fi
        #overwrite by EPOCHS env
        if [ ${EPOCHS} -gt $epochs ] 2>/dev/null ; then
            epochs=${EPOCHS}
        fi
        if (( ! $dry_run )); then
            VENV_NAME=${VENV_NAME} EPOCHS=$epochs NUMBER_OF_PROCESS=$num_process PROCESS_PER_NODE=$num_process $script2run 2>&1 |tee $log_file
        else
    	    echo "VENV_NAME=${VENV_NAME} EPOCHS=$epochs NUMBER_OF_PROCESS=$num_process PROCESS_PER_NODE=$num_process $script2run"
        fi
        perf_val[$( bc -l <<< "$i -1")]=$( grep "Total Throughput" $log_file | awk -F= '{print $2}' |tr -d '\r\n' )
      fi
    done
    print_prefix "AI benchmark evaluation - Tensorflow Resnet50 v1.5 Training DATAVALIDATION on GPU $device_name"

    #single card perf as baseline for scalability
    val_base=${perf_val[0]}

    for i in $( seq 0 7); do
        val=${perf_val[$i]}
        if [ -z $val ];then
        val=0
        fi
        vali=$(bc -l <<< "$val != 0")
        if (( $vali )); then
        perf_data=$(printf '%.1f' ${perf_val[$i]})        
        if [ ! "$val_base" == "0" ]; then
            scale=$(printf '%.2f' $(bc -l <<< "$perf_data / $val_base /${num_of_card[$i]} * 100"))
        fi
        reference_diff=$( bc -l <<<  "${reference_data[$i]} *  $threshold" )
        reference=$( bc -l <<< "${reference_data[$i]} - $reference_diff" )
        result=$(echo "$perf_data >= $reference" |bc -l)
        ratio=$( bc -l <<< "$perf_data / ${reference_data[$i]} * 100")
        ratio=$( printf '%.2f' $ratio )
        threshold_percent=$( bc -l <<< "$threshold * 100")
        if (( $result )); then
            resultstr=PASSED
        else
            resultstr=WARNING
        fi
        echo "DATAVALIDATION $test_id on ${num_of_card[$i]} GPU (imgs/sec) - $resultstr : current $perf_data, reference ${reference_data[$i]}, ratio $ratio%, threshold $threshold_percent%, scalability $scale% "

       	csv_file="${test_id}.csv"
        if ! test -f $csv_file; then
            echo "Workload, #GPU, Result, Performance(imgs/sec), Reference, Ratio(%), Scale(%)" > $csv_file
        fi
        echo "${test_id},${num_of_card[$i]},${resultstr},${perf_data},${reference_data[$i]},${ratio},${scale}" >> $csv_file

        fi
    done

}

validate_ai_resnet50_tf_training

