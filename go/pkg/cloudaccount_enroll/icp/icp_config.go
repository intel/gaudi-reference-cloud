// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package icp

import (
	"context"
	"net/url"
	"time"
)

type Config struct {
	//URL of the ICP
	URL string `koanf:"url"`
	// Path to the file containing the username. It should not containing a trailing new-line.
	UsernameFile string `koanf:"usernameFile"`
	// Path to the file containing the password. It should not containing a trailing new-line.
	PasswordFile string `koanf:"passwordFile"`
}

type ICPConfig struct {
	URL      *url.URL
	UserName string
	Password string
	Timeout  time.Duration
}

func New(ctx context.Context, config *Config) (*ICPConfig, error) {
	return &ICPConfig{}, nil
}
