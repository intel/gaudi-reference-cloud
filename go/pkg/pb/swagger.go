// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package pb

import (
	"embed"
)

//go:embed swagger/*.json
var SwaggerFs embed.FS
var SwaggerDir string = "swagger"
