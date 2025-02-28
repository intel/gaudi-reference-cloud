# Build the manager binary
FROM golang:1.19 as builder
ARG TARGETOS
ARG TARGETARCH

ENV http_proxy=http://internal-placeholder.com:912
ENV https_proxy=http://internal-placeholder.com:912
ENV no_proxy=localhost,127.0.0.0/8,10.0.0.0/8,.intel.com
ENV GOPRIVATE=github.com/intel-innersource/*

# COPY .netrc /root/.netrc

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY pkg/kubernetes_operator/cmd/main.go pkg/kubernetes_operator/cmd/main.go
COPY pkg/kubernetes_operator/api/ pkg/kubernetes_operator/api/
COPY pkg/kubernetes_operator/internal/controller/ pkg/kubernetes_operator/internal/controller/
COPY pkg/kubernetes_operator/kubernetes_provider/ pkg/kubernetes_operator/kubernetes_provider/
COPY pkg/kubernetes_operator/node_provider/ pkg/kubernetes_operator/node_provider/
COPY pkg/kubernetes_operator/addon_provider/ pkg/kubernetes_operator/addon_provider/
COPY pkg/kubernetes_operator/etcd/ pkg/kubernetes_operator/etcd/
COPY pkg/kubernetes_operator/utils/ pkg/kubernetes_operator/utils/

COPY pkg/ilb_operator/cmd/main.go pkg/ilb_operator/cmd/main.go
COPY pkg/ilb_operator/api/ pkg/ilb_operator/api/
COPY pkg/ilb_operator/internal/ pkg/ilb_operator/internal/

COPY pkg/instance_operator/vm/builder pkg/instance_operator/vm/builder

COPY pkg/grpcutil pkg/grpcutil
COPY pkg/pb pkg/pb
COPY pkg/conf pkg/conf
COPY pkg/log pkg/log
COPY pkg/tlsutil pkg/tlsutil

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager pkg/kubernetes_operator/cmd/main.go

FROM debian AS certs

ENV http_proxy=http://internal-placeholder.com:912
ENV https_proxy=http://internal-placeholder.com:912
ENV no_proxy=localhost,127.0.0.0/8,10.0.0.0/8,.intel.com

RUN apt update -qqq && \
  apt install -yqqq ca-certificates \
  unzip \
  wget

RUN wget http://certificates.intel.com/repository/certificates/IntelSHA2RootChain-Base64.zip \
  -O /usr/local/share/ca-certificates/IntelSHA2RootChain-Base64.zip && \
  unzip -u  /usr/local/share/ca-certificates/IntelSHA2RootChain-Base64.zip -d /usr/local/share/ca-certificates && \  
  rm -f /usr/local/share/ca-certificates/IntelSHA2RootChain-Base64.zip && \ 
  chmod 644 /usr/local/share/ca-certificates/*.crt && \
  /usr/sbin/update-ca-certificates

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=certs /etc/ssl/certs /etc/ssl/certs
USER 65532:65532

ENTRYPOINT ["/manager"]
