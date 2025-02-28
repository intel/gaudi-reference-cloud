// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package modules

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/getkin/kin-openapi/openapi3filter"
	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/ci/k6/modules/weka/api/v4"
	auth "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/ci/k6/modules/weka/auth"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	middleware "github.com/oapi-codegen/echo-middleware"
)

type WekaRequest struct {
	Path    string              `json:"path,omitempty"`
	Method  string              `json:"method,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    interface{}         `json:"body,omitempty"`
}

type WekaAPI struct {
	Requests []*WekaRequest
	mutex    sync.RWMutex
}

func WekaV4() WekaAPI {
	return WekaAPI{
		Requests: make([]*WekaRequest, 0),
		mutex:    sync.RWMutex{},
	}
}

func (w *WekaAPI) StartAPI(port int, mock bool) {
	swagger, err := v4.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	fa, err := auth.NewFakeAuthenticator()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating authenticator:\n: %s", err)
		os.Exit(1)
	}

	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.BodyDump(w.recordRequest))

	validator := middleware.OapiRequestValidatorWithOptions(swagger,
		&middleware.Options{
			Options: openapi3filter.Options{
				AuthenticationFunc: auth.NewAuthenticator(fa),
			},
		})

	e.Use(validator)

	if mock {
		v4.RegisterHandlersWithBaseURL(e, v4.NewMockWekaAPI(), "/api/v2")
	} else {
		v4.RegisterHandlersWithBaseURL(e, v4.NewSimulatedWekaAPI(fa), "/api/v2")
	}

	go func() {
		e.Logger.Fatal(e.Start(net.JoinHostPort("0.0.0.0", strconv.Itoa(port))))
	}()
}

func (w *WekaAPI) recordRequest(c echo.Context, reqBody, _ []byte) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	request := c.Request()

	var headers = make(map[string][]string)
	for key, val := range request.Header {
		headers[key] = val
	}

	var body interface{}

	// Fails for login, but struct is correct
	_ = json.Unmarshal(reqBody, &body)
	wekaRequest := WekaRequest{
		Path:    request.RequestURI,
		Method:  request.Method,
		Headers: headers,
		Body:    body,
	}
	w.Requests = append(w.Requests, &wekaRequest)
}
