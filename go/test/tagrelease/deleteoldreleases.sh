#!/bin/bash

# Usage: ./deleteoldreleases.sh <username> <password>

username="$1"
password="$2"
days_threshold="$3"

# Hardcoded values
repoOwner="intel-innersource"
repoName="frameworks.cloud.devcloud.services.idc"

# GitHub API URL
githubApiUrl="https://api.github.com"

# Get releases information
releases_json=$(curl -s -u "${username}:${password}" "${githubApiUrl}/repos/${repoOwner}/${repoName}/releases")

# Get the current date in Unix time format
current_time=$(date +%s)

# Loop through each release and check age
echo "$releases_json" | jq -c '.[]' | while read -r release; do
    release_id=$(echo "$release" | jq -r '.id')
    tag_name=$(echo "$release" | jq -r '.tag_name')
    created_at=$(echo "$release" | jq -r '.created_at')

    # Convert created_at to Unix time
    created_time=$(date -d "$created_at" +%s)

    # Calculate the age in days
    age_days=$(( (current_time - created_time) / 86400 ))

    # Delete if older than days defined in days_threshold
    if [ "$age_days" -gt "$days_threshold" ]; then
        # Delete the release
        curl -s -u "${username}:${password}" -X DELETE "${githubApiUrl}/repos/${repoOwner}/${repoName}/releases/${release_id}"

        # Delete the associated tag
        curl -s -u "${username}:${password}" -X DELETE "${githubApiUrl}/repos/${repoOwner}/${repoName}/git/refs/tags/${tag_name}"

        echo "Deletion complete: Release and tag $tag_name have been removed."
    else
        echo "Skipping release $tag_name, created $age_days days ago."
    fi
done
