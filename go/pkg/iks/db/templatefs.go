// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package db

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"text/template"
)

type TemplateFile struct {
	fs.File
	name string
	data interface{}
}

type TemplateFs struct {
	embed.FS
	data interface{}
}

func NewTemplateFs(fs embed.FS, data interface{}) *TemplateFs {
	return &TemplateFs{FS: fs, data: data}
}

func (fs *TemplateFs) Open(name string) (fs.File, error) {
	file, err := fs.FS.Open(name)
	if err != nil {
		return nil, err
	}
	return TemplateFile{File: file, data: fs.data, name: name}, nil
}

func (f TemplateFile) Read(b []byte) (int, error) {
	_, err := f.File.Read(b)
	if err != nil {
		return 0, err
	}
	buf := new(bytes.Buffer)
	// load and parse the template
	if tmpl, err := template.New(f.name).
		Option("missingkey=error").
		Parse(string(b)); err == nil {
		if err := tmpl.Execute(buf, f.data); err != nil {
			return 0, err
		}
	} else {
		return 0, err
	}
	nullIndex := strings.Index(string(buf.Bytes()), "\x00")
	if nullIndex < 0 {
		return 0, fmt.Errorf("negative null index, something wrong with %s template", f.name)
	}
	b = b[:nullIndex] // resize byte slice to actual template size after parsing
	return copy(b, buf.Bytes()), nil
}

func (f TemplateFile) Stat() (fs.FileInfo, error) {
	// nothing to implement here yet ...
	return nil, nil
}

func (f TemplateFile) Close() error {
	// nothing to implement here yet ...
	return nil
}
