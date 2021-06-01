// Copyright 2021 ghstats Project Authors. Licensed under MIT.

package cmd

import (
	"bytes"

	"github.com/overvenus/ghstats/pkg/config"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newConfigCommand())
}

// newCommand returns PTAL command
func newConfigCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "config",
		Short: "Show configuration template",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &config.Config{
				PTAL: config.PTAL{
					Repos: []config.Repo{{}},
				},
				Review: config.Review{
					Repos: []config.Repo{{}},
				},
			}
			buf := &bytes.Buffer{}
			encoder := toml.NewEncoder(buf)
			encoder.Indentation("")
			if err := encoder.Encode(cfg); err != nil {
				return err
			}
			cmd.Println(buf.String())
			return nil
		},
	}
	return command
}
