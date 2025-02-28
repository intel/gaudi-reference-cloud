#!/bin/bash

# Function to create a tag and return its ID
createtag() {
    local username="$1"
    local password="$2"
    local org_name="$3"
    local repo_name="$4"
    local commit_sha="$5"
    local tag_name="$6"

    tag_payload="{\"tag\": \"$tag_name\", \"object\": \"$commit_sha\", \"body\": \"$commit_log_output\", \"type\": \"commit\", \"message\": \"Creating tag $tag_name\"}"

    # Create tag via GitHub API
    response=$(curl -s -u "$username:$password" -X POST \
            -d "$tag_payload" \
            "https://api.github.com/repos/$org_name/$repo_name/git/tags")

    # Check response status
    status=$(echo "$response" | jq -r '.message')
    if [ "$status" = "Bad credentials" ]; then
        printf "Error: Invalid credentials..."
    elif [ "$status" = "Not Found" ]; then
        printf "Error: Repository not found. Please check the repository details."
    else
        #printf "%s\n" "Tag $tag_name created successfully."

        # Extract the tag SHA from the response
        tag_sha=$(echo "$response" | jq -r '.sha')
        #printf "%s\n" "Tag SHA: $tag_sha"

        # Push the tag to the remote repository
        if [ -n "$tag_sha" ]; then
            tag_ref_response=$(curl -s -u "$username:$password" -X POST \
                    -d "{\"ref\": \"refs/tags/$tag_name\", \"sha\": \"$tag_sha\"}" \
                    "https://api.github.com/repos/$org_name/$repo_name/git/refs")
            #printf "%s\n" "Tag $tag_name pushed successfully."
        else
            printf "Error: Failed to push tag."
        fi
    fi

    # Return tag SHA
    echo "$tag_sha"
}


# Function to create a release and return its ID
createrelease() {
    local username="$1"
    local password="$2"
    local org_name="$3"
    local repo_name="$4"
    local tag_name="$5"
    local formatted_notes="$6"

    # Construct the release notes JSON string
    release_notes="{\"tag_name\":\"$tag_name\",\"name\":\"$tag_name\",\"body\":\"$formatted_notes\"}"

    # Create the release using GitHub Releases API
    response=$(curl -s -u "$username:$password" -X POST \
        -H "Content-Type: application/json" \
        -d "$release_notes" \
        "https://api.github.com/repos/$org_name/$repo_name/releases")

    #printf "%s\n" "$response"


    # Check if release creation was successful
    if [ "$response" == *"message"* ]; then
        printf "%s\n"  "Error: Failed to create release. Response: $response\n"
        exit 1
    else
        release_id=$(echo "$response" | jq -r '.id')
        #printf "%s\n" "Tag $tag_name with release notes pushed successfully."
        #printf "%s\n" " Release id : $release_id\n"
    fi
    echo "$release_id"
}

# GitHub organization/repository details
org_name="intel-innersource"
repo_name="frameworks.cloud.devcloud.services.idc"
username="$1"
password="$2"
#last_prod_commit_sha="$3"

# Tag details
current_date_time=$(date +"%Y-%m-%dT%H-%M-%S-%3N")
tag_name="idc-auto-tag-$current_date_time"
branch_name="origin/main"

# Get the latest commit SHA of the specified branch
commit_sha=$(git rev-parse "$branch_name")

if [ -z "$commit_sha" ]; then
  printf "%s\n" "Error: Unable to get the latest commit SHA of the '$branch_name' branch."
  exit 1
fi

printf "%s\n" "commit id of main is '$commit_sha'\n"
#printf "%s\n" "current production's commit id is '$last_prod_commit_sha'\n"

# Call the createtag function to create the tag and capture the tag SHA
tag_sha=$(createtag "$username" "$password" "$org_name" "$repo_name" "$commit_sha" "$tag_name")

printf "%s\n" "Tag created successfully with SHA: $tag_sha \n"

# Call the commit_log.sh script and capture its output
#commit_log_output=$(./generatereleasenotes.sh "$last_prod_commit_sha" "$commit_sha")
commit_log_output=$(./generatereleasenotes.sh)

formatted_notes=$(echo "$commit_log_output" | sed 's/"/\\"/g; s/$/\\n/' | tr -d '\n')

# Call the createrelease function to create the release and capture the release ID
release_id=$(createrelease "$username" "$password" "$org_name" "$repo_name" "$tag_name" "$formatted_notes")

if [ -z "$release_id" ]; then
  printf "%s\n" "Error: Failed to create release."
  exit 1
else
  printf "%s\n" "Release - $release_id created for the tag name $tag_name successfully\n"
  echo "Release id is $release_id \n"
  echo "Tag name is $tag_name \n"
fi



