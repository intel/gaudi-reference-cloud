// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tradecheckintel

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/apigee"
)

type config struct {
	tokenURL                 string
	createProductURL         string
	createOrderURL           string
	screenBusinessPartnerURL string
	username                 string
	password                 string
}

type GTSclient struct {
	cfg       config
	mu        sync.RWMutex
	client    *resty.Client
	apgclient *apigee.ApigeeClient
	token     atomic.Pointer[Token]
}

func CreateConfig(usernameFile string, passwordFile string, tokenURL string, createProductURL string, createOrderURL string, screenBusinessPartnerURL string) (config, error) {
	cfg := config{}
	username, err := os.ReadFile(usernameFile)
	if err != nil {
		return cfg, fmt.Errorf("error creating GTS client config, unable to read username file %s: %v", usernameFile, err)
	}
	password, err := os.ReadFile(passwordFile)
	if err != nil {
		return cfg, fmt.Errorf("error creating GTS client config, unable to read password file %s: %v", passwordFile, err)
	}
	return config{
		username:                 string(username),
		password:                 string(password),
		tokenURL:                 tokenURL,
		createProductURL:         createProductURL,
		createOrderURL:           createOrderURL,
		screenBusinessPartnerURL: screenBusinessPartnerURL,
	}, nil
}

func NewClient(cfg config) (*GTSclient, error) {
	cli := resty.New().SetTimeout(1 * time.Minute)

	apgcli, err := apigee.NewClient(cfg.tokenURL, cfg.username, cfg.password)
	if err != nil {
		return nil, fmt.Errorf("error encountered in creating Apigee client: %v", err)
	}

	return &GTSclient{
		cfg:       cfg,
		client:    cli,
		apgclient: apgcli,
	}, nil
}
