// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package config

import (
	"os"
)

type Configuration struct {
	usernameFile string
	passwordFile string
}

func NewConfiguration(usernameFile string, passwordFile string) (*Configuration, error) {

	_, err := readFile(usernameFile)
	if err != nil {
		return nil, err
	}

	_, err = readFile(passwordFile)
	if err != nil {
		return nil, err
	}

	return &Configuration{
		usernameFile: usernameFile,
		passwordFile: passwordFile,
	}, nil
}

// Read the username file from disk
func (c *Configuration) GetAPIUsername() (string, error) {
	return readFile(c.usernameFile)
}

// Read the password file from disk
func (c *Configuration) GetAPIPassword() (string, error) {
	return readFile(c.passwordFile)
}

// Read a file from disk returning it's contents
func readFile(configFileName string) (string, error) {
	val, err := os.ReadFile(configFileName)
	if err != nil {
		return "", err
	}
	return string(val), nil
}
