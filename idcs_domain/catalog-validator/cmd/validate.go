/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"catalog-validator/pkg/validator"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "validate::catalog-validator",
	Long: `This command validates the correctness of provided custom resource yaml file(s) with their corresponding definitions.
			Example: catalog-validator validate file1.yaml file2.yaml --src=path/to/CRD
			Example: catalog-validator validate file1.yaml file2.yaml -s=path/to/CRD`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			src, _ := cmd.Flags().GetString("src")
			if src == "" {
				fmt.Println("Please provide path to valid CRD yaml files using --src| -s flag!")
				os.Exit(1)
			}
			validator.ValidateResourceYaml(args, src); 
		} else {
			fmt.Println("Please provide valid yaml file path(s) to validate.")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringP("src", "s", "", "CRD file path for validation")
}
