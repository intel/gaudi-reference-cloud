// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

// Needed to force specific version in go.mod
import (
	_ "cloud.google.com/go"
	_ "cloud.google.com/go/gaming/apiv1/gamingpb"
	_ "github.com/chzyer/logex"
	_ "github.com/chzyer/test"
	_ "github.com/decred/dcrd/dcrec/secp256k1/v4"
	_ "github.com/edsrzf/mmap-go"
	_ "github.com/elazarl/goproxy"
	_ "github.com/go-kit/kit/log"
	_ "github.com/gobuffalo/logger"
	_ "github.com/gobuffalo/packd"
	_ "github.com/gobuffalo/packr/v2"
	_ "github.com/gofrs/uuid"
	_ "github.com/google/renameio"
	_ "github.com/google/s2a-go"
	_ "github.com/googleapis/gax-go/v2"
	_ "github.com/karrick/godirwalk"
	_ "github.com/lestrrat-go/blackmagic"
	_ "github.com/lestrrat-go/iter/arrayiter"
	_ "github.com/lestrrat-go/option"
	_ "github.com/markbates/oncer"
	_ "github.com/miekg/dns"
	_ "github.com/minio/sha256-simd"
	_ "github.com/pkg/sftp"
	_ "github.com/sony/gobreaker"
	_ "github.com/vishvananda/netns"
	_ "github.com/xdg-go/stringprep"
	_ "google.golang.org/api"
	_ "gopkg.in/gcfg.v1"
	_ "gopkg.in/warnings.v0"
	_ "oras.land/oras-go/pkg/auth"
)

func main() {
}
