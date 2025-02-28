// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package etcd

import (
	"context"
	"crypto/tls"
	"crypto/x509"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	etcdclientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/snapshot"
	"go.uber.org/zap"

	"time"

	"github.com/pkg/errors"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/utils"
)

const (
	etcdClientTimeout         = time.Duration(30 * time.Second)
	etcdClientSnapshotTimeout = time.Duration(15 * time.Minute)
	etcdDialTimeout           = time.Duration(30 * time.Second)
)

type Client struct {
	*etcdclientv3.Client
	Config etcdclientv3.Config
}

func NewClient(etcdIP string, etcdPort string, cert, key, ca []byte) (*Client, error) {
	var tlsConfig tls.Config

	clientCertificate, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing cert and key")
	}
	tlsConfig.Certificates = []tls.Certificate{clientCertificate}

	caCertificate, err := utils.ParseCert(ca)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing ca certificate")
	}
	certPool := x509.NewCertPool()
	certPool.AddCert(caCertificate)
	tlsConfig.RootCAs = certPool

	config := etcdclientv3.Config{
		Endpoints:   []string{"https://" + etcdIP + ":" + etcdPort},
		DialTimeout: 5 * time.Second,
		TLS:         &tlsConfig,
	}

	etcdClient, err := etcdclientv3.New(config)
	if err != nil {
		return nil, errors.Wrapf(err, "create etcd client")
	}

	return &Client{etcdClient, config}, nil
}

func (c *Client) EtcdSnapshot(ctx context.Context, etcdLogger *zap.Logger, dbpath string) error {
	ctx, cancel := context.WithTimeout(ctx, etcdClientSnapshotTimeout)
	err := snapshot.Save(ctx, etcdLogger, c.Config, dbpath)
	cancel()

	if err != nil {
		return errors.Wrapf(err, "etcd snapshot")
	}

	return nil
}

func (c *Client) ListEtcdMembers(ctx context.Context) ([]*etcdserverpb.Member, error) {
	ctx, cancel := context.WithTimeout(ctx, etcdClientTimeout)
	resp, err := c.MemberList(ctx)
	cancel()

	if err != nil {
		return nil, errors.Wrapf(err, "list etcd members")
	}

	return resp.Members, nil
}

func (c *Client) RemoveEtcdMember(ctx context.Context, id uint64) error {
	ctx, cancel := context.WithTimeout(ctx, etcdClientTimeout)
	_, err := c.MemberRemove(ctx, id)
	cancel()

	if err != nil {
		return errors.Wrapf(err, "remove etcd member with id %d", id)
	}

	return nil
}

func (c *Client) MemberStatus(ctx context.Context, endpoint string) error {
	ctx, cancel := context.WithTimeout(ctx, etcdClientTimeout)
	_, err := c.Status(ctx, endpoint)
	cancel()

	if err != nil {
		return errors.Wrapf(err, "get endpoint status of %s", endpoint)
	}

	return nil
}
