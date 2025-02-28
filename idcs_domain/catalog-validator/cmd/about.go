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
	"fmt"

	"github.com/spf13/cobra"
)

// aboutCmd represents the about command
var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "about::catalog-validator",
	Long: `Catalog Validator is a CLI tool for Product Catalog Developers to quickly validate new custom resources
	like vendors, products, etc before adding them to product catalog.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("About :: Catalog Validator")
		fmt.Println("Catalog Validator is a CLI tool that validates CR yamls with given CRD yamls.")
	},
}

func init() {
	rootCmd.AddCommand(aboutCmd)
}
