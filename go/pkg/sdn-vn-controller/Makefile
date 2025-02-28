MAIN_PACKAGE_PATH = ${base_dir}/cmd
APP_NAME=sdnctl

make_path := ${abspath ${lastword ${MAKEFILE_LIST}}}
base_dir := ${dir ${make_path}}

## build: build the application
.PHONY: build
build: api
	go build -o sdncontroller sdncontroller.go 
	go build -o sdnctl client/sdnctl.go

## controller: build only the controller
.PHONY: controller
controller: api
	go build -o sdncontroller sdncontroller.go 

## client: build only the sdnctl client
.PHONY: client
client: api
	go build -o sdnctl client/sdnctl.go

## all: build everything and run the unit tests
.PHONY: all
all: controller client test

## api: build the grpc
.PHONY: api
api:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./api/sdn/v1/ovnnet.proto

## test: run unit tests
.PHONY: test
test: api
	go test -v -cover -count=1 ./handlers


## secrets: generate some sample certificates and keys to test secure connections
.PHONY: secrets
secrets:
	. ./tests/generate_secrets.sh