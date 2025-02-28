#!/bin/bash
set -o pipefail

# Go to the directory of this script
cd $(cd "$(dirname "$0")" && pwd)

# Set PATH
export PATH=$PATH:/usr/local/go/bin:/home/$USER/go/bin/:/home/$USER/.local/bin

# Download go dependecies
cd ../tests/goFramework/ginkGo
go mod download
go mod tidy
go install github.com/onsi/ginkgo/v2/ginkgo

# Run Longevity test
cd compute/longevity_test_cases
ginkgo -v --timeout=48h --tags='provision_deprovision_loop' -- --instanceType='bm-spr' --sshPublicKey="$(cat ~/.ssh/id_rsa.pub)"
