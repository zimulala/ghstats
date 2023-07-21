package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v35/github"
	"github.com/overvenus/ghstats/pkg/config"
	"github.com/overvenus/ghstats/pkg/feishu"
	"github.com/overvenus/ghstats/pkg/gh"
	"github.com/overvenus/ghstats/pkg/markdown"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

func init() {
	rootCmd.AddCommand(newPkgsCommand())
}

// newCommand returns pkgs command
func newPkgsCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "pkgs",
		Short: "Collect daily PRs for these pkgs ❤️",
		RunE: func(cmd *cobra.Command, args []string) error {
			today := time.Now().In(timeZone)
			lastDay := today
			switch today.Weekday() {
			// Monday, collect past 3 days review activity.
			case time.Monday:
				lastDay = lastDay.Add(-3 * 24 * time.Hour)
			// Others, collect past 1 day review activity.
			default:
				lastDay = lastDay.Add(-24 * time.Hour)
			}

			return getPRs(cmd, "Daily", lastDay, today)
		},
	}

	command.AddCommand(&cobra.Command{
		Use:   "weekly",
		Short: "Collect weekly PRs for these pkgs ❤️",
		RunE: func(cmd *cobra.Command, args []string) error {
			today := time.Now().In(timeZone)
			firstDayThisWeek := time.Date(
				today.Year(), today.Month(), today.Day()-int(today.Weekday())+1,
				10, 0, 0, 0, today.Location(),
			)
			// [UTC+8 10:00 on Monday this week, now]
			return getPRs(cmd, "Weekly", firstDayThisWeek, today)
		},
	})

	command.AddCommand(&cobra.Command{
		Use:   "monthly",
		Short: "Collect monthly PRs for these pkgs ❤️",
		RunE: func(cmd *cobra.Command, args []string) error {
			today := time.Now().In(timeZone)
			firstDayLastMonth := time.Date(
				today.Year(), today.Month()-1, today.Day(),
				10, 0, 0, 0, today.Location(),
			)
			// [UTC+8 10:00 on the first day of last month, now]
			return getPRs(cmd, "Monthly", firstDayLastMonth, today)
		},
	})

	return command
}

func getPRs(cmd *cobra.Command, kind string, start, end time.Time) error {
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
	pInfo := ptalInfo{packages: cfg1.Packages, startTimestamp: start, endTimestamp: end}
	fmt.Printf("[pkgs: %s] PRs %s %s-%s\n", pInfo.packages, kind, start.Format(time.RFC3339), end.Format(time.RFC3339))

	ProjectPRs := make(map[string][]*github.PullRequest)
	for _, proj := range cfg.Repos {
		repoInfs := strings.SplitN(proj.PROwnerRepo, "/", 2)
		if len(repoInfs) != 2 {
			return errors.New(fmt.Sprintf("repo str:%v, split strings:%v", proj.PROwnerRepo, repoInfs))
		}

		results, err := gh.PullRequestsList(ctx, client, repoInfs[0], repoInfs[1])
		if err != nil {
			return err
		}
		ProjectPRs[proj.Name] = append(ProjectPRs[proj.Name], results...)
	}
	buf := strings.Builder{}
	for repo, results := range ProjectPRs {
		prs := strings.Builder{}
		for _, pr := range results {
			// filter out PR creation time beyond [start, end) range
			if !pInfo.withinTimeRange(*pr.CreatedAt) {
				continue
			}
			// filter PR created by ti-chi-bot
			if strings.Contains(*pr.User.Login, "ti-chi-bot") {
				fmt.Printf("filter creates PR by bot, url:%s, title:%s \n", pr.GetHTMLURL(), pr.GetTitle())
				continue
			}
			// filter out the cfg.packages
			isContainPkg, err := pInfo.isInPackages(client, pr)
			if err != nil {
				return err
			}
			if !isContainPkg {
				fmt.Printf("filter doesn't contain pkgs:%s, url:%s, title:%s \n", pInfo.packages, pr.GetHTMLURL(), pr.GetTitle())
				continue
			}

			prs.WriteString(fmt.Sprintf("%s %s\n",
				markdown.Link(fmt.Sprintf("#%d", *pr.Number), *pr.HTMLURL),
				markdown.Escape(*pr.Title),
			))
		}
		if prs.Len() != 0 {
			buf.WriteString(fmt.Sprintf("## %s\n", markdown.Escape(repo)))
			buf.WriteString(prs.String())
		}
	}
	if buf.Len() == 0 {
		// Good! No PR need to be reviewed.
		fmt.Println("No PR need to be reviewed.")
		return nil
	}
	bot := feishu.WebhookBot(cfg.FeishuWebhookToken)
	return bot.SendMarkdownMessage(ctx, fmt.Sprintf("PTAL Pkgs:%v ❤️ - %s", pInfo.packages, kind), buf.String(), feishu.TitleColorWathet)
}

type ptalInfo struct {
	packages       []string
	startTimestamp time.Time
	endTimestamp   time.Time
}

// Is the ts within [start, end)?
func (c *ptalInfo) withinTimeRange(ts time.Time) bool {
	ts = ts.In(timeZone)
	return (ts.After(c.startTimestamp) || ts.Equal(c.startTimestamp)) && ts.Before(c.endTimestamp)
}

func (c *ptalInfo) isInPackages(client *github.Client, pr *github.PullRequest) (bool, error) {
	if len(c.packages) == 0 {
		return false, nil
	}

	owner, repo := gh.GetPRRepository(pr)
	number := pr.GetNumber()
	prFiles, err := gh.PullRequestsListFiles(context.Background(), client, owner, repo, number)
	if err != nil {
		return false, err
	}
	for _, file := range prFiles {
		for _, pkg := range c.packages {
			if strings.Contains(*file.Filename, pkg) {
				return true, nil
			}
		}
	}
	return false, nil
}
