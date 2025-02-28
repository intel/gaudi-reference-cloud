// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package gateway

const (
	ServiceName   = "maas-gateway"
	MetricsPrefix = "maas_gateway"
)

var (
	errInvalidArgument = "invalid argument. Please try again"
	errInvalidModel    = "the requested model is not supported. Please try again with a valid model"
	errUnavailable     = "sorry, we are at maximum capacity. Please come back later"
	errGeneral         = "could not generate stream. Please try again"
)
