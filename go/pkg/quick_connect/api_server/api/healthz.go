// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddHealthzRoutes(router *gin.Engine) error {
	router.GET("/readyz", readyz)
	return nil
}

func readyz(ctx *gin.Context) {
	ctx.String(http.StatusOK, "ok")
}
