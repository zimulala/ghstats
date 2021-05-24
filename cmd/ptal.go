// Copyright 2021 ghstats Project Authors. Licensed under MIT.

package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v35/github"
	"github.com/overvenus/ghstats/pkg/config"
	"github.com/overvenus/ghstats/pkg/feishu"
	"github.com/overvenus/ghstats/pkg/markdown"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

func init() {
	cliCmd := newCommand()
	rootCmd.AddCommand(cliCmd)
}

// Execute runs the root command
func newCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "ptal",
		Short: "Please take a look Pull Requests ❤️",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			cfg1, err := config.ReadConfig(cfgPath)
			if err != nil {
				return err
			}
			cfg := cfg1.PTAL
			ctx := context.Background()
			client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: cfg.GithubToken},
			)))

			for {
				var results []*github.IssuesSearchResult
				for _, query := range cfg.PRQuery {
					result, resp, err := client.Search.Issues(
						ctx, query, nil)
					if err != nil {
						if _, ok := err.(*github.RateLimitError); ok {
							cmd.PrintErrln("hit rate limit, sleep 1s")
							time.Sleep(time.Second)
							continue
						}
						return err
					}
					if resp.StatusCode != http.StatusOK {
						body, _ := ioutil.ReadAll(resp.Body)
						return fmt.Errorf("search issue error [%d] %s", resp.StatusCode, string(body))
					}
					results = append(results, result)
				}
				buf := strings.Builder{}
				// To keep message short, we only keep the most recent 8 PRs.
				max := 8
				for i, res := range results {
					for j, issue := range res.Issues {
						if j > max {
							break
						}
						buf.WriteString(fmt.Sprintf("%s %s\n",
							markdown.Link(fmt.Sprintf("#%d", *issue.Number), *issue.HTMLURL),
							markdown.Escape(*issue.Title),
						))
					}
					if i+1 != len(results) {
						buf.WriteString(fmt.Sprintf("%s\n", markdown.Separate))
					}
				}
				if buf.Len() == 0 {
					// Good! No PR need to be reviewed.
					return nil
				}
				bot := feishu.WebhookBot(cfg.FeishuWebhookToken)
				return bot.SendMarkdownMessage(ctx, "PTAL ❤️", buf.String())
			}
		},
	}
	return command
}
