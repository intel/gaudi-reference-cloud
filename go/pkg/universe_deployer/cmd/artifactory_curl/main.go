// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Upload or download artifact in Artifactory.

package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/artifactory"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	flag "github.com/spf13/pflag"
)

type Arguments struct {
	Output        string
	Request       string
	RetentionDays int
	UploadFile    string
	Url           string
}

func parseArgs() Arguments {
	var args Arguments

	flag.StringVar(&args.Output, "output", "", "Downloaded artifact will be written to this file")
	flag.StringVar(&args.Request, "request", "GET", "'GET' or 'PUT'")
	flag.IntVar(&args.RetentionDays, "retention-days", 30, "Retention days")
	flag.StringVar(&args.UploadFile, "file", "", "File to upload")
	flag.StringVar(&args.Url, "url", "", "URL")

	flag.Parse()

	args.Output = util.AbsFromWorkspace(args.Output)
	args.UploadFile = util.AbsFromWorkspace(args.UploadFile)

	util.EnsureRequiredStringFlag("request")
	util.EnsureRequiredStringFlag("url")

	return args
}

func main() {
	ctx := context.Background()
	log.SetDefaultLogger()
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("upload_artifacts"))
	log.Info("BEGIN")
	defer log.Info("END")

	err := func() error {
		var args = parseArgs()
		log.Info("args", "args", args)

		artifactory, err := artifactory.New(ctx)
		if err != nil {
			return err
		}
		artifactory.RetentionDays = args.RetentionDays
		artifactUrl, err := url.Parse(args.Url)
		if err != nil {
			return err
		}
		if args.Request == "GET" {
			if err := artifactory.Download(ctx, *artifactUrl, args.Output); err != nil {
				return err
			}
		} else if args.Request == "PUT" {
			if err := artifactory.Upload(ctx, args.UploadFile, *artifactUrl); err != nil {
				return err
			}

		} else {
			return fmt.Errorf("unsupported request '%s'", args.Request)
		}

		return nil
	}()
	if err != nil {
		log.Error(err, "error")
		os.Exit(1)
	}
}
