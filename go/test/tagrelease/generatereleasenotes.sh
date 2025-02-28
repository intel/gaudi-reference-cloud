#!/bin/bash

# Function to generate the commit log table
generate_commit_log() {
    #local commit1="$1"
    #local commit2="$2"

    local VMaaS_authors="(Claudio Fahey|GopeshKh|nishyt|malsbat|nmmani)"
    local BMaaS_authors="(keedyandre|rmanijacob|Sandeep|tewit|Sanatan Shrivastava|adduarte|pnhowe-intel|Saiteja-Garlapati)"
    local LB_authors="(stevesloka)"

    echo "### Virtual Machine as a Service"
    echo "-----------------------------"
    local vmaas_git_log_output=$(git log origin/main --pretty=format:"| %h | %an | %ad | %s |" --since="yesterday" --until="now" --perl-regexp --author="$VMaaS_authors")
    echo "$vmaas_git_log_output"
    echo ""

    echo "### Bare Metal as a Service"
    echo "------------------------"
    local bmaas_git_log_output=$(git log origin/main --pretty=format:"| %h | %an | %ad | %s |" --since="yesterday" --until="now" --perl-regexp --author="$BMaaS_authors")
    echo "$bmaas_git_log_output"
    echo ""

    # Run git log command and capture the output into a variable
    echo "### Load Balancer as a Service"
    echo "---------------------------"
    local lb_git_log_output=$(git log origin/main --pretty=format:"| %h | %an | %ad | %s |" --since="yesterday" --until="now" --perl-regexp --author="$LB_authors")
    echo "$lb_git_log_output"
    echo ""
}

# Check if commit hashes are provided as parameters
#if [ "$#" -ne 2 ]; then
#    echo "Usage: $0 <commit1> <commit2>"
#    exit 1
#fi

# Call the function with provided parameters and capture the output
commit_log_output=$(generate_commit_log)

# Output the captured commit log
echo "$commit_log_output"
