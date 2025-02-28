// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type tString struct {
	xml.Name   `xml:"string"`
	StringName string `xml:"name,attr"`
	Value      string `xml:"value,attr"`
}

type tRule struct {
	xml.Name `xml:"rule"`
	Class    string    `xml:"class,attr"`
	Location string    `xml:"location,attr"`
	RuleName string    `xml:"name,attr"`
	Strings  []tString `xml:"string"`
}

type tQuery struct {
	xml.Name `xml:"query"`
	Rule     []tRule `xml:"rule"`
}

type tExternal struct {
	importpath string
	version    string
	sum        string
}

var extMap map[string]*tExternal

func main() {
	extMap = getExternal()
	for _, arg := range os.Args[1:] {
		genSbom(arg)
	}
}

func getExternal() map[string]*tExternal {
	cmd := exec.Command("bazel", "query", "--noshow_progress", "--output=xml", "//external:*")
	buf := bytes.Buffer{}
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	query := tQuery{}

	// lop off xml header, it has version 1.1 and go xml only
	// supports 1.0
	bb := buf.Bytes()
	ii := bytes.IndexByte(bb, '\n')
	if ii == -1 {
		log.Fatal("no newline in xml output")
	}

	if err := xml.Unmarshal(bb[ii:], &query); err != nil {
		log.Fatal(err)
	}

	extMap := map[string]*tExternal{}
	for _, rule := range query.Rule {
		colon := strings.IndexByte(rule.RuleName, ':')
		if colon == -1 {
			log.Fatalf("missing colon in %v", rule.RuleName)
		}
		name := rule.RuleName[colon+1:]
		importpath := ""
		version := ""
		sum := ""
		for _, str := range rule.Strings {
			switch str.StringName {
			case "importpath":
				importpath = str.Value
			case "version":
				version = str.Value
			case "sum":
				sum = str.Value
			}
		}
		if importpath == "" || version == "" || sum == "" {
			// Skip anything that doesn't look like a go module
			continue
		}

		extMap[name] = &tExternal{
			importpath: importpath,
			version:    version,
			sum:        sum,
		}
	}
	return extMap
}

func genSbom(target string) {
	buf := strings.Builder{}
	cmd := exec.Command("bazel", "query", "--noshow_progress", fmt.Sprintf("deps(%s)", target))
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	deps := map[string]bool{}
	for _, line := range strings.Split(buf.String(), "\n") {
		if len(line) < 2 {
			continue
		}
		if line[:1] != "@" {
			// Skip dependencies within our repo
			continue
		}
		ii := strings.Index(line, "//")
		if ii == -1 {
			continue
		}
		dep := line[1:ii]
		deps[dep] = true

	}

	fmt.Printf("%v\n", target)
	var extDeps []*tExternal
	for dep := range deps {
		ext, ok := extMap[dep]
		if !ok {
			// Skip dependencies on things other than go modules.
			// We can include these additional dependencies on tools and
			// libraries. We just need to define how these additional
			// dependencies should appear in the SBOM.
			//
			// For now what we show is intended to replace the data from
			// "go version -m" on go binaries produced by "go build" (as opposed to
			// bazel's rules_go)
			continue
		}
		extDeps = append(extDeps, ext)
	}

	sort.Slice(extDeps, func(left int, right int) bool {
		return extDeps[left].importpath < extDeps[right].importpath
	})

	for _, ext := range extDeps {
		fmt.Printf("\tdep\t%v\t%v\t%v\n", ext.importpath, ext.version, ext.sum)
	}
}
