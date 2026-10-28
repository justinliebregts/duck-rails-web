#!/bin/bash

#!/bin/bash
#set -euo pipefail # Exit immediately on error, treat unset variables as error, fail if any command in a pipeline fails.

# --- 1. Variable Definitions and Directory Setup ---

# Get the directory where the script is located, regardless of current working directory
SCRIPT_ROOT="$(dirname "$0")"

# Save the current directory to return to later
P_WD=$(pwd)

# Change to the script's root directory
cd "${SCRIPT_ROOT}"

# --- 2. Environment Setup Calls ---

# Call shared environment cleanup and setup scripts
# NOTE: These .bat files need to be translated to .sh scripts or replaced with their Bash equivalents.
# Assuming the translated files are sharedEnvClean.sh and sharedEnv.sh
# For simplicity, we just call them here, but the called scripts must be Bash.

# forget all variable
# "./shared/sharedEnvClean.sh"

# Load environment variables (tencentBase, tcrNamespace, tcrRepo, tcrUsername, tcrPass)
# These variables should be 'exported' in sharedEnv.sh
source "./sv/sharedEnv.sh"

TCR_REPO="duckrails-api-golang"

# Login with target repo (Assuming login.sh takes the base, user, and pass as arguments)
docker login ${tencentBase}  --username=${tcrUsername}  -p ${tcrPass}

# Change to the build context root (assumes the build context is 3 directories up)
cd ../../..

# --- 3. CAPTURE GIT COMMIT HASH ---

# Capture the current short commit hash into a shell variable
GIT_COMMIT=$(git rev-parse --short HEAD)

# --- 4. Build and Push ---

# Define the base tag (without the version)
BASE_TAG="${tencentBase}/${tcrNamespace}/${TCR_REPO}"

# Construct the final unique tag: [BASE_TAG]:api-[COMMIT_HASH]
TAG="${BASE_TAG}:api-${GIT_COMMIT}"

echo "==================================================================="
echo "Building and Pushing Image: ${TAG}"
echo "==================================================================="

# NOTE: The 'docker images' line was removed as it serves no functional purpose.

# Build the image, passing the Git hash as a build argument
docker build \
    --build-arg GIT_COMMIT="${GIT_COMMIT}" \
    --platform linux/amd64 \
    -t "${TAG}" \
    -f ./Dockerfile .

# Check the exit status of the previous command
if [ $? -ne 0 ]; then
    echo "ERROR: Docker build failed! Exiting."
    # Jumps to the end of the script for cleanup
    exit 1
fi

# Push image to Registry
docker push "${TAG}"

if [ $? -ne 0 ]; then
    echo "ERROR: Docker push failed! Exiting."
    exit 1
fi

# --- 5. Cleanup ---

# forget all variable (Clean up the environment)
#"./shared/sharedEnvClean.sh"

# Return to the original working directory
cd "${P_WD}"

# Note: 'timeout 30' is replaced with 'sleep 5' (or removed)
sleep 5
# @echo on is removed as it's not applicable in Bash