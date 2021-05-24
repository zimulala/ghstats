// Copyright 2021 ghstats Project Authors. Licensed under MIT.

package main

import (
	_ "github.com/google/go-github/v35/github"
	"github.com/overvenus/ghstats/cmd"
	_ "github.com/spf13/cobra"
)

func main() {
	cmd.Execute()
}
