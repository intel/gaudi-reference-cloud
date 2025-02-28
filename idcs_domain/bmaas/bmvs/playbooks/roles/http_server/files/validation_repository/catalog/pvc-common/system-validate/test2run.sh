#!/bin/bash
# A wrapper for test, default gemm
# gemm stream zepeer zebandwidth onecclbench hpl hpl_ai tf_resnet50_training pt_bertlarge_infer
expected_passnum=${EXPECTED_PASSNUM:-0}
test2run=${TEST2RUN:-gemm}

case $test2run in
    gemm)
    testscripts=sys_val_gemm
    ;;
    stream)
    testscripts=sys_val_stream
    ;;
    zepeer)
    testscripts=sys_val_zepeer
    ;;
    zebandwidth)
    testscripts=sys_val_zebandwidth
    ;;
    onecclbench)
    testscripts=sys_val_onecclbench
    ;;
    hpl)
    testscripts=sys_val_hpl
    ;;
    hpl_ai)
    testscripts=sys_val_hpl_ai
    ;;
    tf_resnet50_training)
    testscripts=AI/sys_val_tf_resnet50_training
    ;;
    pt_bertlarge_infer)
    testscripts=AI/sys_val_pt_bertlarge_infer
    ;;
    *)
    echo "Unknown test case"
    exit -1
esac

logfile=$testscripts.log
bash $testscripts.sh 2>&1 |tee $logfile

testwarning=$( grep DATAVALIDATION $logfile | grep WARNING )
testwarningnum=$( grep DATAVALIDATION $logfile | grep WARNING |wc -l)
testpass=$( grep DATAVALIDATION $logfile |grep PASSED )
testpassnum=$( grep DATAVALIDATION $logfile |grep PASSED |wc -l)

if [ $expected_passnum -gt 0 ]; then
    if [ ! $testpassnum -eq $expected_passnum ]; then
        echo "Test ${test2run} FAILED, expect $expected_passnum PASS but find $testpassnum"
        exit -1
    fi
fi

if [ $testwarningnum -gt 0 ]; then
    echo "$testwarning"
    echo "Test ${test2run} FAILED with $testwarningnum WARNINGS"
    exit -1
fi

echo "Test ${test2run} PASSED with $testpassnum test items PASSED"


