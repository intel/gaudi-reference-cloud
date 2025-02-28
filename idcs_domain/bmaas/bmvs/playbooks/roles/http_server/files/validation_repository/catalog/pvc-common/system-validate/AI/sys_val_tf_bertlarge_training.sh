#/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
dry_run=${DRYRUN:-0}
DOCKER=${DOCKER:-0}
VENV_NAME=${VENV_NAME:-bench_tf_bertlarge_training}

if [ "$DOCKER" != "0" ]; then
    LOG_DIR=${WORKSPACE}/bench_tf_bertlarge_training_container_logs
    WORK_DIR=${WORKSPACE}/bench_tf_bertlarge_training_container_workdir
else
    LOG_DIR=${WORKSPACE}/bench_tf_bertlarge_training_logs
    WORK_DIR=${WORKSPACE}/bench_tf_bertlarge_training_workdir
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
else
    tiles=1
fi
device_name=$(lspci -vmm |grep Max | head -n 1)

validate_tf_bertlarge_training(){
    #define threshold and refernce data
    threshold=0.05
    if [ $tiles == 1 ]; then
        #reference data for Max 1100
        num_of_card=(1 2 3 4 5 6 7 8)
        reference_data=(45 85 0 160 0 0 0 270)
    elif [ $tiles == 2 ]; then
        #reference dat for Max 1550
        num_of_card=(1 2 3 4 5 6 7 8)
        reference_data=(110 210 0 420 0 0 0 850)
    fi
    perf_val=(0 0 0 0 0 0 0 0)

    #running in benchmark folder
    test_id="tf_bertlarge_training"
    cp -r ${SCRIPT_DIR}/../../workloads/AI/BERT/tensorflow/training/* ${WORK_DIR}
    bench_script_dir=${WORK_DIR}
    date4log=""
    #$(date '+%Y%m%d_%H%M%S')
    if [ "$DOCKER" != "0" ]; then
        script2run=./tf_bertlarge_training_container_run.sh
        log_file_prefix=${LOG_DIR}/tf_bertlarge_training_container_run
    script2setup=./tf_bertlarge_training_container_setup.sh
    log_file_setup=${LOG_DIR}/tf_bertlarge_training_container_setup_${date4log}.log
    else
    script2run=./tf_bertlarge_training_run.sh
        log_file_prefix=${LOG_DIR}/tf_bertlarge_training_run
    script2setup=./tf_bertlarge_training_setup.sh
    log_file_setup=${LOG_DIR}/tf_bertlarge_training_setup_${date4log}.log
    fi
    cd $bench_script_dir

    #setup environments
    print_prefix "AI benchmark evaluation - Tensorflow Bert Large Training on $device_name"
    echo "### Set up working environment ..."    
    if (( ! $dry_run ));then
        ./download_pretrained_models.sh
        VENV_NAME=$VENV_NAME $script2setup 2>&1 |tee $log_file_setup
    else
        echo "./download_pretrained_models.sh"
        echo "VENV_NAME=$VENV_NAME $script2setup"
    fi

    for i in 1 2 4 8; do
        if (( $( echo "$num_gpu >= $i" | bc -l ) )); then
            echo "### Running benchmarks on $i GPU card"
            num_process=$( bc -l <<< "$tiles*$i" )
            log_file=${log_file_prefix}_${i}c${num_process}t.log
            if (( ! $dry_run )); then
                VENV_NAME=$VENV_NAME NUMBER_OF_PROCESS=$num_process PROCESS_PER_NODE=$num_process $script2run 2>&1 |tee $log_file
            else
                echo "VENV_NAME=$VENV_NAME NUMBER_OF_PROCESS=$num_process PROCESS_PER_NODE=$num_process $script2run"
            fi
            perf_val[$( bc -l <<< "$i -1")]=$( grep "Total Throughput" $log_file | awk -F' ' '{ print $3 }' |tr -d '\r\n' )
       fi
    done
    print_prefix "AI benchmark evaluation - Tensorflow Bert Large Training DATAVALIDATION on GPU $device_name"

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
        scale=$(printf '%.2f' $(bc -l <<< "$perf_data / $val_base /${num_of_card[$i]} * 100"))
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
        echo "DATAVALIDATION $test_id on ${num_of_card[$i]} GPU (sentences/s) - $resultstr : current $perf_data, reference ${reference_data[$i]}, ratio $ratio%, threshold $threshold_percent%, scalability $scale% "

        csv_file="${test_id}.csv"
        if ! test -f $csv_file; then
            echo "Workload, #GPU, Result, Performance(sentences/s), Reference, Ratio(%), Scale(%)" > $csv_file
        fi
        echo "${test_id},${num_of_card[$i]},${resultstr},${perf_data},${reference_data[$i]},${ratio},${scale}" >> $csv_file

        fi
    done

}

validate_tf_bertlarge_training

