// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package support

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

func GetSupportMD(ctx context.Context, releaseId string) (common.ReleaseSupportMD, error) {
	logger := log.FromContext(ctx).WithName("common.MakePUTAPICall")
	supportMD := common.ReleaseSupportMD{}

	relNum := strings.Split(releaseId, "v")
	if len(relNum) != 2 {
		logger.Info(strings.Join(relNum, ""), "invalid version", releaseId)
		return supportMD, fmt.Errorf("invalid input version")
	}

	mm := strings.Split(relNum[1], ".")
	if len(mm) != 3 {
		logger.Info(strings.Join(mm, ""), "invalid version", len(releaseId))
		return supportMD, fmt.Errorf("invalid input version")
	}

	queryRelUri := fmt.Sprintf("/kubernetes/%s.%s.json", mm[0], mm[1])
	retCode, resp, err := common.MakeGetAPICall(ctx, "https://endoflife.date/api", queryRelUri, nil)
	if err != nil {
		logger.Error(err, "error querying eos url")
		return supportMD, fmt.Errorf("error making eos query")
	}
	if retCode != http.StatusOK {
		logger.Info("eos-query", "unexpected error code ", retCode)
		return supportMD, fmt.Errorf("error making eos query")
	}

	logger.Info("return result", "output", string(resp))
	if err := json.Unmarshal(resp, &supportMD); err != nil {
		logger.Error(err, "error parsing response")
	}

	supportMD.EOLTime, err = time.Parse("2006-01-02", supportMD.Eol)
	if err != nil {
		logger.Error(err, "error parsing supportMD.Eol")
	}
	supportMD.EOSTime, err = time.Parse("2006-01-02", supportMD.Eos)
	if err != nil {
		logger.Error(err, "error parsing supportMD.Eos")
	}
	logger.Info("", "supportMD", supportMD)
	return supportMD, nil
}
