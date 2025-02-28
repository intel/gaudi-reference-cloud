#!/bin/bash

# File containing the list of directories
DIRECTORIES=("../public_apis/vmaas" 
"../business_usecases/vmaas/vmaas_anti_affinity" 
"../business_usecases/vmaas/vmaas_cloud_init" 
"../business_usecases/vmaas/vmaas_create_all_instancetypes" 
"../business_usecases/vmaas/vmaas_metering_monitoring" 
"../business_usecases/vmaas/vmaas_multi_sshkeys" 
"../business_usecases/vmaas/vmaas_multitenancy" 
"../business_usecases/vmaas/vmaas_node_pool" 
"../business_usecases/vmaas/vmaas_quota_enforcement" 
"../business_usecases/vmaas/vmaas_sshtest" 
"../business_usecases/vmaas/vmaas_with_diff_machine_images"
"../public_apis/bmaas"
"../business_usecases/bmaas/bmaas_cloud_init"
"../business_usecases/bmaas/bmaas_metering_monitoring"
"../business_usecases/bmaas/bmaas_quota_enforcement"
"../business_usecases/bmaas/validation_operator"
"../business_usecases/bmaas/bmaas_sshtest"
"../business_usecases/bmaas/bmaas_multi_sshkeys"
"../business_usecases/bmaas/bmaas_crd_validation"
"../business_usecases/bmaas/bmaas_create_all_instancetypes"
"../business_usecases/bmaas/bmaas_bgp"
"../business_usecases/bmaas/bmaas_gaudi3"
"../public_apis/loadbalancer")

# Files to be removed
FILES_TO_REMOVE=("*log.txt", "*.html", "*.xml") # Add file names here

# Loop through each directory in the hardcoded list
for dir in "${DIRECTORIES[@]}"; do
    # Check if the directory exists
    if [[ -d "$dir" ]]; then
        echo "Processing directory: $dir"
        for pattern in "${FILES_TO_REMOVE[@]}"; do
            # Find and remove files matching the pattern
            for file in "$dir"/$pattern; do
                if [[ -f "$file" ]]; then
                    echo "Removing $file"
                    rm "$file"
                else
                    echo "No files matching $pattern found in $dir."
                fi
            done
        done
    else
        echo "Directory $dir does not exist. Skipping."
    fi
done

echo "Completed."
