#!/bin/bash

# Define the test directories manually
vm_directories=(
    "./public_apis/vmaas"
    "./business_usecases/vmaas/vmaas_cloud_init"
    "./business_usecases/vmaas/vmaas_metering_monitoring"
    "./business_usecases/vmaas/vmaas_quota_enforcement"
    "./business_usecases/vmaas/vmaas_sshtest"
    "./business_usecases/vmaas/vmaas_multi_sshkeys"
    "./business_usecases/vmaas/vmaas_crd_validation"
    "./business_usecases/vmaas/vmaas_create_all_instancetypes"
    "./business_usecases/vmaas/vmaas_node_pool"
    "./business_usecases/vmaas/vmaas_anti_affinity"
    "./business_usecases/vmaas/vmaas_multitenancy"
    "./business_usecases/vmaas/vmaas_with_diff_machine_images"
    # Add more directories as needed
)

bm_directories=(
    "./public_apis/bmaas"
    "./business_usecases/bmaas/bmaas_cloud_init"
    "./business_usecases/bmaas/bmaas_metering_monitoring"
    "./business_usecases/bmaas/bmaas_quota_enforcement"
    "./business_usecases/bmaas/validation_operator"
    "./business_usecases/bmaas/bmaas_sshtest"
    "./business_usecases/bmaas/bmaas_multi_sshkeys"
    "./business_usecases/bmaas/bmaas_crd_validation"
    "./business_usecases/bmaas/bmaas_create_all_instancetypes"
    "./business_usecases/bmaas/bmaas_bgp"
    "./business_usecases/bmaas/bmaas_gaudi3"
    # Add more directories as needed
)

lb_directories=(
    "./public_apis/loadbalancer"
    #"./business_usecases/loadbalancer/lb-patch-usecases"
    #"./business_usecases/loadbalancer/lb-quota-validation"
    #"./business_usecases/loadbalancer/lb-with-different-monitor-types"
    #"./business_usecases/loadbalancer/lb-with-instance-resource-ids"
    #"./business_usecases/loadbalancer/lb-with-instance-selectors"
    #"./business_usecases/loadbalancer/lb-with-multi-listeners"
    #"./business_usecases/loadbalancer/lb-with-specific-ips"
)

# Check if a valid argument is provided
if [ $# -lt 1 ]; then
    echo "Usage: $0 <test_directories> [additional flags]"
    exit 1
fi

# Determine which set of test directories to use based on the provided argument
case "$1" in
    "vm_directories")
        test_directories=("${vm_directories[@]}")
        ;;
    "bm_directories")
        test_directories=("${bm_directories[@]}")
        ;;
    "lb_directories")
        test_directories=("${lb_directories[@]}")
        ;;
    *)
        echo "Invalid argument: $1"
        echo "Usage: $0 <test_directories> [additional flags]"
        exit 1
        ;;
esac

# Shift to process additional flags
shift

# Collect additional flags
additional_flags="$@"

failed=false

# # Run Ginkgo for each test directory with the specified label filter
parallel_suites=("vmaas_create_all_instancetypes" "vmaas_with_diff_machine_images" "bmaas_create_all_instancetypes" "bmaas_with_diff_machine_images")
for dir in "${test_directories[@]}"; do
    matched=false
    for name in "${parallel_suites[@]}"; do
        if [[ "$dir" == *"$name"* ]]; then
            matched=true
            echo "directory is matching"
            break
        fi
    done
    # Run Ginkgo for the current test directory
    if $matched; then
        ginkgo -nodes=1 --output-interceptor-mode=none --label-filter=compute "$dir" -- $additional_flags
    else
        ginkgo --label-filter=compute "$dir" -- $additional_flags
    fi
    # Check if the Ginkgo command failed
    if [ $? -ne 0 ]; then
        failed=true
    fi
done

# Remove the existing report.html file
rm -f report.html
rm -f junit_results.xml

# Merge all report.html files from the test directories into a single file
for dir in "${test_directories[@]}"; do
    if [ -f "$dir/report.html" ]; then
        cat "$dir/report.html" >> report.html
        rm "$dir/report.html"  # Remove the report.html file after concatenation
    fi
done

# Merge all junit.xml files from the test directories into a single file
for dir in "${test_directories[@]}"; do
    if [ -f "$dir/junit_results.xml" ]; then
        echo >> junit_results.xml
        cat "$dir/junit_results.xml" >> junit_results.xml
        rm "$dir/junit_results.xml"  # Remove the junit.xml file after concatenation
    fi
done

# Input and output files
input_file="junit_results.xml"
temp_file=$(mktemp)
output_file="processed_results.xml"

# Step 1: Remove all but the first occurrence of <?xml ...>
awk '
    /^<\?xml/ {
        if (xml_found == 0) {
            xml_found = 1
            print
        }
        next
    }
    { print }
' "$input_file" > "$temp_file"

# Step 2: Remove all but the first occurrence of <testsuites ...>
awk '
    BEGIN { testsuites_found = 0 }
    /^[[:space:]]*<testsuites[^>]*>/ {
        if (testsuites_found == 0) {
            testsuites_found = 1
            sub(/^[[:space:]]*<testsuites[^>]*>/, "  <testsuites>")
            print
        }
        next
    }
    { print }
' "$temp_file" > "$input_file"

# Step 3: Keep only the last occurrence of </testsuites>
awk '
    /^[[:space:]]*<\/testsuites>/ {
        last_testsuites_line = NR
        next
    }
    { print }
    END {
        if (last_testsuites_line > 0) {
            print "</testsuites>"
        }
    }
' "$input_file" > "$temp_file"

# Step 4: Delete the first line if it is empty
awk 'NR == 1 { if ($0 != "") print; next } { print }' "$temp_file" > "$input_file"

# Move the temp file to the output file
mv "$input_file" "$output_file"
mv "$output_file" "$input_file"


# If any test case fails during the entire execution, exit with a non-zero exit code
if [ "$failed" = true ]; then
    exit 1
fi
