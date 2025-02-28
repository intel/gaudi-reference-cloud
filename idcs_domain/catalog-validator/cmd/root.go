package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "catalog-validator",
		Version: getVersion(),
		Short: "Catalog Validator CLI",
		Long: `Catalog Validator is a CLI tool for Product Catalog Developers to quickly validate new custom resources
		like vendors, products, etc before adding them to product catalog.`,
	}
)

func getVersion() string {

	app := "git"

    arg0 := "rev-parse"
	arg1 := "--short"
    arg2 := "HEAD"

	var commit = exec.Command(app, arg0, arg1, arg2)
	var version, err = commit.Output()

	if err != nil {
        fmt.Println(err.Error())
		os.Exit(1)
    }

	return string(version)
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}