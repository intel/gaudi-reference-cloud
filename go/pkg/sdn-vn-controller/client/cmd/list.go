// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all instances of an entity",
	Long: `List all instances of an entity like VPC, Subnet, Router, etc. For example, to list VPCs:

	sdnctl list vpc`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list called")
		//		initConn()
		//		listSwitches()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
