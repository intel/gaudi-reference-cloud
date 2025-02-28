// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

func main() {
	ctx := context.Background()
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("main")

	var task tasks.Runnable
	var err error
	const supportedArgs = "enroll|disenroll"

	if len(os.Args) != 2 {
		log.Error(fmt.Errorf("needs one argument"), "exiting", "supported", supportedArgs)
		os.Exit(1)
	}

	switch arg := os.Args[1]; arg {
	case "enroll":
		task, err = tasks.NewEnrollmentTask(ctx)
		if err != nil {
			log.Error(err, "task failed to initialize")
			os.Exit(1)
		}
	case "disenroll":
		task, err = tasks.NewDisenrollmentTask(ctx)
		if err != nil {
			log.Error(err, "task failed to initialize")
			os.Exit(1)
		}
	default:
		log.Error(fmt.Errorf("invalid argument %s", arg), "exiting", "supported", supportedArgs)
		os.Exit(1)
	}

	if err := task.Run(ctx); err != nil {
		log.Error(err, "task failed to run")
		os.Exit(1)
	}
}
