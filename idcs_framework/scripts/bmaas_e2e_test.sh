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

# Run BMaaS e2e tests
cd compute/e2e-test-cases/bmaas
ginkgo -v --tags='BM' -- --instanceType='bm-spr' --sshPublicKey="$(cat ~/.ssh/id_rsa.pub)"
