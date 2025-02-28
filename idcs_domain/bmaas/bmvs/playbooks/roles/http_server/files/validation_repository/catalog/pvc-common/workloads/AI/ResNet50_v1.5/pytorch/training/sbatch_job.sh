#!/bin/bash

nodes=$1
jobname=$(basename $PWD)

echo "generate hostfile from $1"
echo $1 | tr "," "\n" > hostfile

nodes_num=$(wc -l < hostfile)
echo "Required Nodes: ${nodes_num}"

sbatch \
  -N ${nodes_num} \
  -w ${nodes} \
  --job-name=${jobname} \
  --output=$(date +"%Y%m%d")-${nodes}-${jobname}-$(date +"%s").txt \
  --export=node=${nodes} \
  work.sh

