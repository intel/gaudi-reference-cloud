ARG http_proxy
ARG https_proxy
ARG no_proxy
ARG GIN_MODE=debug

FROM internal-placeholder.com/cache/library/golang:1.19 as builder
ARG GIN_MODE
ENV GIN_MODE=${GIN_MODE}
COPY . /go/src/bmaas/enrollment/apiservice/
RUN ls -lhRa /go/src/bmaas/
WORKDIR /go/src/bmaas/enrollment/apiservice
RUN --mount=type=cache,mode=0755,target=/go/pkg/mod go mod vendor
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/bmaas-enrollment cmd/main.go

FROM internal-placeholder.com/cache/library/alpine:3.16 as Final
RUN apk update && apk add --no-cache --upgrade \
    busybox \
    curl \
    openssl

ARG GIN_MODE
ENV GIN_MODE=${GIN_MODE}
WORKDIR /
COPY --from=builder /go/src/bmaas/enrollment/apiservice/bin/bmaas-enrollment /

ENTRYPOINT ["/bmaas-enrollment"]
