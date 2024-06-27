#!/bin/bash

# this is the original shell script that inspired the go cli application.
# It is a bash script that clones all repositories from a github organization and generates the modules.xml file for the intellij project.
# the app covers a lot more features and is more user friendly and flexible than the bash script.

set -euo pipefail

usage() {
  echo "Usage: $0 [OPTIONS]"
  echo
  echo "Options:"
  echo "  -h, --help             Show this help message and exit"
  echo "  --pulls-on-existing    Perform 'git pull' on existing repositories"
  echo "  --clone-depth N        Clone repositories with depth N (default: full clone)"
  echo
  echo "Environment Variables:"
  echo "  GITHUB_TOKEN     GitHub personal access token for accessing private repositories"
  echo "                   If not provided, only public repositories will be processed"
  echo "                   https://github.com/settings/tokens"
  echo
}

# Function to fetch repositories from GitHub API
# Appends the repos array from the response to .all.json
fetch_repos() {
  local page=$1
  local token=$2
  local url="$GITHUB_API_URL?sort=updated&type=all&page=$page&per_page=$PER_PAGE"

  if [[ -z "$token" ]]; then
    curl -s "$url"
  else
    curl -s -H "Authorization: token $token" "$url"
  fi | jq -r '.[]' > .all.json.tmp
  if [[ -f .all.json ]]; then
    jq --argjson repos '["\(.[])"]' '.all.json |= . + $repos' .all.json.tmp > .all.json
  else
    mv .all.json.tmp .all.json
  fi
}

# Function to process a repository.
# It will clone or pull the repository if it already exists.
# It will save the repo github json (part of the list reponse) to a file inside repo/.github/repo.json
process_repo() {
  local repo_url=$1
  local pulls_on_existing=$2
  local clone_depth=$3
  local repo_name

  repo_name=$(basename "$repo_url" .git)

  if [[ -d "$repo_name" ]]; then
    if $pulls_on_existing; then
      echo "[$((i+1))] $repo_name already exists. Performing git pull..."
      cd "$repo_name"
      git pull
      cd ..
    else
      echo "[$((i+1))] $repo_name already exists. Skipping..."
    fi
  else
    echo "[$((i+1))] $repo_name: Cloning..."
    if [[ -z "$clone_depth" ]]; then
      git clone "$repo_url"
    else
      git clone --depth "$clone_depth" "$repo_url"
    fi
  fi

  jq --arg url "$repo_url" '.[] | select(.clone_url == $url)' .all.json > "$repo_name/.github/repo.json"

  # Add repo name to the modules array
  modules+=("$repo_name")
}

# Check if modules.xml exists and has a project block
check_modules_file() {
  local modules_file=$1
  if [[ -f "$modules_file" && $(grep -q '<project>' "$modules_file") ]]; then
    return 0
  else
    return 1
  fi
}

# Check if module for repo_name already exists
check_module_exists() {
  local modules_file=$1
  local repo_name=$2
  if grep -q "<module .*filepath=\".*$repo_name.*\".*/>" "$modules_file"; then
    return 0
  else
    return 1
  fi
}

# Add module for repo_name
add_module() {
  local modules_file=$1
  local repo_name=$2
  local temp_file
  temp_file=$(mktemp)
  # Check if mktemp failed to create a temporary file
  if [[ ! -f "$temp_file" ]]; then
    echo "Failed to create a temporary file"
    return 1
  fi

  # Add the new module to the temporary file
  sed "/<\/project>/i \  <module fileurl=\"file://\$PROJECT_DIR\$/$repo_name\" filepath=\"\$PROJECT_DIR\$/$repo_name\" />" "$modules_file" > "$temp_file"

  # Replace the original modules.xml file with the temporary file
  mv "$temp_file" "$modules_file"
}

# Generate modules for all repos in the modules array
generate_modules() {
  local modules_file=".idea/_modules.xml"
  local repo_name

  if check_modules_file "$modules_file"; then
    for repo_name in "${modules[@]}"; do
      if check_module_exists "$modules_file" "$repo_name"; then
        echo "Module for $repo_name already exists. Skipping..."
      else
        echo "Adding module for $repo_name..."
        add_module "$modules_file" "$repo_name"
      fi
    done
  else
    echo "Creating modules.xml file and adding modules..."
    echo '<?xml version="1.0" encoding="UTF-8"?>' > "$modules_file"
    echo '<project version="4">' >> "$modules_file"
    for repo_name in "${modules[@]}"; do
      add_module "$modules_file" "$repo_name"
    done
    echo '</project>' >> "$modules_file"
  fi
}

ORG_NAME="gymondo-git"
GITHUB_API_URL="https://api.github.com/orgs/$ORG_NAME/repos"
PER_PAGE=100

PULLS_ON_EXISTING=false
CLONE_DEPTH=
modules=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    --pulls-on-existing)
      PULLS_ON_EXISTING=true
      shift
      ;;
    --clone-depth)
      CLONE_DEPTH="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      usage
      exit 1
      ;;
  esac
done

GITHUB_TOKEN=${GITHUB_TOKEN:-}

PAGE=1
i=0

while true; do
  RESPONSE=$(fetch_repos "$PAGE" "$GITHUB_TOKEN")

  CURRENT_REPOS=$(echo "$RESPONSE" | jq -r '.[].clone_url')

  for repo_url in $CURRENT_REPOS; do
    process_repo "$repo_url" "$PULLS_ON_EXISTING" "$CLONE_DEPTH"
    ((i++))
  done

  # Check if there are more pages
  if [[ -z "$CLONE_DEPTH" ]]; then
    if [[ $(echo "$RESPONSE" | jq -r 'length') -lt $PER_PAGE ]]; then
      break
    fi
  fi

  PAGE=$((PAGE + 1))
done

# Generate modules after processing all repos
generate_modules
