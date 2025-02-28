###############################################################################
# Go Framework Makefile
###############################################################################

export GOPATH = $(shell go env GOPATH)

export PATH := $(PATH):$(GOPATH)/bin

.PHONY: download-go-dependencies
download-go-dependencies: 
	go mod download
	go mod tidy
	go install github.com/onsi/ginkgo/v2/ginkgo
