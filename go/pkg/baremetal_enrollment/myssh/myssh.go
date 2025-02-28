// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package myssh

//go:generate mockgen -destination ../mocks/myssh.go -package mocks github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/myssh SSHManagerAccessor,SSHClientAccessor,SSHSessionAccessor

import (
	"golang.org/x/crypto/ssh"
)

type SSHManager interface {
	SSHManagerAccessor
	SSHClientAccessor
	SSHSessionAccessor
}

type MySSHManager struct {
}

type SSHManagerAccessor interface {
	Dial(network, addr string, config *ssh.ClientConfig) (SSHClientAccessor, error)
}

func (s *MySSHManager) Dial(network, addr string, config *ssh.ClientConfig) (SSHClientAccessor, error) {
	client, err := ssh.Dial(network, addr, config)
	myClient := &MySSHClient{Client: client}
	return myClient, err
}

type SSHClientAccessor interface {
	NewSession() (SSHSessionAccessor, error)
	Close() error
}

type MySSHClient struct {
	Client *ssh.Client
}

type SSHSessionAccessor interface {
	CombinedOutput(cmd string) ([]byte, error)
	Close() error
}

func (s *MySSHClient) NewSession() (SSHSessionAccessor, error) {
	session, err := s.Client.NewSession()
	mySession := &MySSHSession{Session: session}
	return mySession, err
}

func (s *MySSHClient) Close() error {
	return s.Client.Close()
}

type MySSHSession struct {
	Session *ssh.Session
}

func (s *MySSHSession) CombinedOutput(cmd string) ([]byte, error) {
	return s.Session.CombinedOutput(cmd)
}

func (s *MySSHSession) Close() error {
	return s.Session.Close()
}
