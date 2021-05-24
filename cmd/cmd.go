// Copyright 2021 ghstats Project Authors. Licensed under MIT.

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Cmd of the ghstats projects.
var rootCmd = &cobra.Command{
	Use:   "gh",
	Short: "Github Tools",
}

// Execute runs the root command
func Execute() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "Configuration file")
	// Ouputs cmd.Print to stdout.
	rootCmd.SetOut(os.Stdout)
	if err := rootCmd.Execute(); err != nil {
		rootCmd.Println(err)
		os.Exit(1)
	}
}
