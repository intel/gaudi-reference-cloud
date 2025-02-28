// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/build/tools/util"
)

var includeDir string

const validate = "/pkg/mod/github.com/envoyproxy/protoc-gen-validate@v0.10.1"

func main() {

	outputDir := ""
	flag.StringVar(&outputDir, "output-dir", "", "output dirctory")
	flag.StringVar(&includeDir, "I", "", "include dir")
	flag.Parse()
	if outputDir == "" {
		log.Fatal("--output-dir is required")
	}

	deployFiles := map[string][]string{}
	args := flag.Args()
	if len(args) < 1 {
		return
	}

	// Sort arguments to ensure reproducible output.
	sort.Strings(args)

	for _, fileName := range args {
		fd, err := util.ReadProtoFd(fileName)
		if err != nil {
			log.Fatal(err)
		}
		deploy := util.GetDeploy(fd)
		if deploy == "" {
			continue
		}
		deployFiles[deploy] = append(deployFiles[deploy], fileName)
		deployFiles["all"] = append(deployFiles["all"], fileName)
	}
	tmpDir, err := os.MkdirTemp("", "protoc_swagger")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	protoPath := filepath.Dir(args[0])
	for deploy, files := range deployFiles {
		cmdArgs := []string{
			"--proto_path", protoPath,
			"-I", (os.Getenv("GOPATH") + validate),
			"--openapiv2_out=" + tmpDir,
			"--openapiv2_opt=generate_unbound_methods=true",
			"--openapiv2_opt=allow_merge=true,merge_file_name=protoc",
		}
		cmdArgs = append(cmdArgs, files...)
		cmd := exec.Command("protoc", cmdArgs...)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("proto error: %v", err)
		}

		outFileName := fmt.Sprintf("%v/%v.swagger.json", outputDir, deploy)
		if util.WriteFileAtomically(outFileName,
			func(outf io.Writer) error {
				fd, err := os.Open(tmpDir + "/protoc.swagger.json")
				if err != nil {
					log.Fatal(err)
				}
				content, err := io.ReadAll(fd)
				if err != nil {
					log.Fatal(err)
				}
				swagger := map[string]any{}
				if err = json.Unmarshal(content, &swagger); err != nil {
					log.Fatal(err)
				}
				info := swagger["info"]
				if info == nil {
					log.Fatalf("output for %v is missing info object", deploy)
				}
				infoMap, ok := info.(map[string]any)
				if !ok {
					log.Fatalf("output for %v has wrong type for info object", deploy)
				}
				label := ""
				if deploy == "all" {
					label = "Complete"
				} else {
					label = strings.ToUpper(deploy[:1]) + deploy[1:]
				}
				infoMap["title"] = fmt.Sprintf("Intel Developer Cloud: %v API", label)
				if content, err = json.MarshalIndent(swagger, "", "  "); err != nil {
					log.Fatal(err)
				}
				if _, err = outf.Write(content); err != nil {
					log.Fatal(err)
				}
				return nil
			}); err != nil {
			log.Fatal(err)
		}
	}
}
