#!/bin/bash

# GitHub organization/repository details
org_name="intel-innersource"
repo_name="frameworks.cloud.devcloud.services.idc"
release_id="$1"
username="$2"
password="$3"
# GitHub release ID (you can find this from the GitHub UI or through the API)

# Function to update release notes with regression status
update_release_notes() {
    local release_id="$1"
    local username="$2"
    local password="$3"
    local org_name="$4"
    local repo_name="$5"
    local status="$6"

    # Fetch release information
    release_info=$(curl -s -u "$username:$password" \
        "https://api.github.com/repos/$org_name/$repo_name/releases/$release_id")


    # Extract existing release notes from the response
    existing_notes=$(echo "$release_info" | jq -r '.body')
    #echo "$existing_notes"

    new_notes="$existing_notes"$'\n\n'"Regression test status: $status"

    # Escape double quotes in new_notes
    new_notes_escaped=$(echo "$new_notes" | sed 's/"/\\"/g')

    # Construct JSON payload with properly escaped new_notes using jq
    payload=$(jq -n --arg new_notes "$new_notes_escaped" '{"body": $new_notes}')

    # JSON payload to update release body
    #payload="{\"body\": \"$new_notes\"}"

    # Update the release via GitHub Releases API
    response=$(curl -s -u "$username:$password" -X PATCH \
        -H "Content-Type: application/json" \
        -d "$payload" \
        "https://api.github.com/repos/$org_name/$repo_name/releases/$release_id")

    echo "$response"
}

# Update release notes with regression status
output=$(update_release_notes "$release_id" "$username" "$password" "$org_name" "$repo_name" "passed")

if [ "$output" == *"message"* ]; then
    printf "%s\n"  "Error: Failed to update release. Response: $output\n"
    exit 1
else
    echo "Release $release_id notes updated successfully with regression result"
fi

