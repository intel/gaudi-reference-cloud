#!/bin/bash
# Example to run on x8 Max1550 system
result_file=test_smcx8oam_report.log
echo "" >> $result_file

echo "Test Start Time [$(date)]" >> $result_file
test=(gemm stream zepeer zebandwidth onecclbench hpl hpl_ai tf_resnet50_training pt_bertlarge_infer)
passnum=(160 16 57 144 8 12 10 4 4)
for i in $(seq 0 8); do
      EXPECTED_PASSNUM=${passnum[$i]} TEST2RUN=${test[$i]} ./test2run.sh
      testresult=$?
      if [ $testresult -eq 0 ]; then
            echo TEST ${test[$i]} PASSED >> $result_file
      else
            echo TEST ${test[$i]} FAILED >> $result_file
      fi
done
echo "Test End Time [$(date)]" >> $result_file


#EXPECTED_PASSNUM=160 TEST2RUN=gemm ./test2run.sh
#EXPECTED_PASSNUM=16 TEST2RUN=stream ./test2run.sh
#EXPECTED_PASSNUM=57 TEST2RUN=zepeer ./test2run.sh
#EXPECTED_PASSNUM=144 TEST2RUN=zebandwidth ./test2run.sh
#EXPECTED_PASSNUM=8 TEST2RUN=onecclbench ./test2run.sh
#EXPECTED_PASSNUM=12 TEST2RUN=hpl ./test2run.sh
#EXPECTED_PASSNUM=10 TEST2RUN=hpl_ai ./test2run.sh
#EXPECTED_PASSNUM=4 TEST2RUN=tf_resnet50_training ./test2run.sh
#EXPECTED_PASSNUM=4 TEST2RUN=pt_bertlarge_infer ./test2run.sh
