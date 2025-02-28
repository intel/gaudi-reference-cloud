#!/bin/bash

for t in 1 2 3 31 4; do 
	TESTSET=$t ./run_llm_inference_test.sh
done
