// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloud_init

import (
	"bytes"
	"fmt"

	"github.com/flosch/pongo2"
)

func InitTemplateEngineOptions() error {
	if !pongo2.FilterExists("indent") {
		if err := pongo2.RegisterFilter("indent", indentFilter); err != nil {
			return fmt.Errorf("Error registering custom indent filter: %s", err)
		}
	}

	pongo2.SetAutoescape(false)

	return nil
}

// func Indent(input string, width int) string {
// 	// Create a buffer to store the indented string
// 	var indented bytes.Buffer

// 	// Iterate over the input string and add user requested indentation to the string
// 	for i := 0; i < len(input); i++ {
// 		indented.WriteByte(input[i])

// 		if input[i] == '\n' {
// 			for indent := 0; indent < width; indent++ {
// 				indented.WriteByte(' ')
// 			}
// 		}
// 	}

// 	return indented.String()
// }

func Indent(input string, width int) string {
	// Create a buffer to store the indented string
	var indented bytes.Buffer

	// Initialize an indentation flag
	needIndent := true

	// Iterate over the input string and add user-requested indentation to the string
	for i := 0; i < len(input); i++ {
		// Add the indentation spaces if needed
		if needIndent {
			for indent := 0; indent < width; indent++ {
				if err := indented.WriteByte(' '); err != nil {
					// Handle the error from WriteByte
					fmt.Println("Error writing indentation space:", err)
					return ""
				}
			}
			needIndent = false
		}

		// Append the current character to the indented buffer
		if err := indented.WriteByte(input[i]); err != nil {
			// Handle the error from WriteByte
			fmt.Println("Error writing character:", err)
			return ""
		}

		// Set the indentation flag if a newline character is encountered
		if input[i] == '\n' {
			needIndent = true
		}
	}

	return indented.String()
}

func indentFilter(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
	indented := Indent(in.String(), param.Integer())
	return pongo2.AsValue(indented), nil
}
