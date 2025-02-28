#!/bin/bash

suites=("../public_apis/loadbalancer")

# Check if a valid argument is provided
if [ $# -lt 1 ]; then
    echo "Usage: $0 <suites> [additional flags]"
    exit 1
fi

max_procs=5

additional_flags="$@"

# Function to run a suite
run_suite() {
    local suite=$1
    echo "Running $suite..."
    ginkgo -r ./$suite -- $additional_flags
}

export -f run_suite
export additional_flags

echo "Running suites in parallel with $max_procs processes..."
parallel -j $max_procs run_suite ::: "${suites[@]}"

# Remove the existing report.html file
rm -f report.html
rm -f junit_results.xml

# Merge all report.html files from the test directories into a single file
for dir in "${suites[@]}"; do
    if [ -f "$dir/report.html" ]; then
        cat "$dir/report.html" >> report.html
        rm "$dir/report.html"  # Remove the report.html file after concatenation
    fi
done

# Merge all junit.xml files from the test directories into a single file
for dir in "${suites[@]}"; do
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
