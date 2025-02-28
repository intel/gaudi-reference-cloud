#!/bin/bash

# Usage: ./addregressiontoreleasenotes.sh <releaseTag> <username> <password> <regressionDetails>

releaseId=$1
username=$2
password=$3
regressionDetails=$4

# # Hardcoded values
repoOwner="intel-innersource"
repoName="frameworks.cloud.devcloud.services.idc"

# GitHub API URL
githubApiUrl="https://api.github.com"

# Step 1: Fetch the current release notes using the GitHub API
releaseResponse=$(curl -s -u "${username}:${password}" \
    "${githubApiUrl}/repos/${repoOwner}/${repoName}/releases/${releaseId}")

# Check if the release exists
if [[ "$(echo "$releaseResponse" | jq -r '.id')" == "null" ]]; then
    echo "Error: Release not found for ID ${releaseId}"
    exit 1
fi

# Fetch the current release notes
currentReleaseNotes=$(echo "$releaseResponse" | jq -r '.body')

echo "Regression Details: ${regressionDetails}"

# Use the provided regression details string
# Ensure the regressionDetails is not empty
if [[ -z "$regressionDetails" ]]; then
    echo "Error: Regression details string is empty!"
    exit 1
fi

# Combine the existing release notes with the regression table
updatedReleaseNotes="${currentReleaseNotes}\n\n${regressionDetails}"

# # Escape special characters in the updated release notes for JSON
# escapedReleaseNotes=$(echo -e "$updatedReleaseNotes" | jq -Rs .)

# Update the release notes via the GitHub API
updateResponse=$(curl -s -X PATCH -u "${username}:${password}" \
    -d "{\"body\": \"${updatedReleaseNotes}\"}" \
    "${githubApiUrl}/repos/${repoOwner}/${repoName}/releases/${releaseId}")

# Check if the update was successful
if echo "$updateResponse" | grep -q '"id":'; then
    echo "Release notes updated successfully!"
else
    echo "Error updating release notes: $updateResponse"
    exit 1
fi
