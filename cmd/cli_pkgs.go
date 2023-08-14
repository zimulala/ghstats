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

const (
	DailyKind   = "Daily"
	WeeklyKind  = "Weekly"
	MonthlyKind = "Monthly"
)

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

			return getPRs(cmd, DailyKind, lastDay, today)
		},
	}

	command.AddCommand(&cobra.Command{
		Use:   "weekly",
		Short: "Collect weekly PRs for these pkgs ❤️",
		RunE: func(cmd *cobra.Command, args []string) error {
			today := time.Now().In(timeZone)
			firstDayThisWeek := time.Date(
				today.Year(), today.Month(), today.Day()-7,
				today.Hour(), today.Minute(), today.Second(), 0, today.Location(),
			)
			// [UTC+8 now.hour:now.min on Monday this week, now]
			return getPRs(cmd, WeeklyKind, firstDayThisWeek, today)
		},
	})

	command.AddCommand(&cobra.Command{
		Use:   "monthly",
		Short: "Collect monthly PRs for these pkgs ❤️",
		RunE: func(cmd *cobra.Command, args []string) error {
			today := time.Now().In(timeZone)
			firstDayLastMonth := time.Date(
				today.Year(), today.Month()-1, today.Day(),
				today.Hour(), today.Minute(), today.Second(), 0, today.Location(),
			)
			// [UTC+8 now.hour:now.min on the first day of last month, now]
			return getPRs(cmd, MonthlyKind, firstDayLastMonth, today)
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
	pInfo := ptalInfo{startTimestamp: start, endTimestamp: end}
	fmt.Printf("[repos: %s] PRs %s %s - %s\n", cfg.ReposName(), kind, start.Format(time.RFC3339), end.Format(time.RFC3339))

	maxPages := 5
	switch kind {
	case DailyKind:
		maxPages = 2
	case WeeklyKind:
		maxPages = 5
	case MonthlyKind:
		maxPages = 20
	}

	buf := strings.Builder{}
	for _, proj := range cfg.Repos {
		repoInfs := strings.SplitN(proj.PROwnerRepo, "/", 2)
		if len(repoInfs) != 2 {
			return errors.New(fmt.Sprintf("repo str:%v, split strings:%v", proj.PROwnerRepo, repoInfs))
		}

		results, err := gh.PullRequestsList(ctx, client, repoInfs[0], repoInfs[1], maxPages)
		if err != nil {
			return err
		}
		filterPR(client, pInfo, proj, results, &buf)
	}

	if buf.Len() == 0 {
		// Good! No PR need to be reviewed.
		fmt.Println("No PR need to be reviewed.")
		return nil
	}
	bot := feishu.WebhookBot{Token: cfg.FeishuWebhookToken, IsTest: cfg1.IsOnlyPrintMsg}
	return bot.SendMarkdownMessage(ctx, fmt.Sprintf("%s PTAL Repos:[%v] ❤️ - %s", cfg.ReportName, cfg.ReposName(), kind),
		buf.String(), feishu.TitleColorWathet)
}

type ptalInfo struct {
	startTimestamp time.Time
	endTimestamp   time.Time
}

// Is the ts within [start, end)?
func (c *ptalInfo) withinTimeRange(ts time.Time) bool {
	ts = ts.In(timeZone)
	return (ts.After(c.startTimestamp) || ts.Equal(c.startTimestamp)) && ts.Before(c.endTimestamp)
}

func isInPackages(client *github.Client, packages []string, pr *github.PullRequest) (bool, error) {
	if len(packages) == 0 {
		return false, nil
	}

	owner, repo := gh.GetPRRepository(pr)
	number := pr.GetNumber()
	prFiles, err := gh.PullRequestsListFiles(context.Background(), client, owner, repo, number)
	if err != nil {
		return false, err
	}
	for _, file := range prFiles {
		for _, pkg := range packages {
			if strings.Contains(*file.Filename, pkg) {
				return true, nil
			}
		}
	}
	return false, nil
}

func filterPR(client *github.Client, pInfo ptalInfo, repo config.Repo,
	projectPRs []*github.PullRequest, buf *strings.Builder) error {
	prs := strings.Builder{}
	for _, pr := range projectPRs {
		// filter out PR creation time beyond [start, end) range
		if !pInfo.withinTimeRange(*pr.CreatedAt) {
			fmt.Printf("repo:%s filter PR created time:%s, url:%s, title:%s \n",
				repo.Name, *pr.CreatedAt, pr.GetHTMLURL(), pr.GetTitle())
			continue
		}
		// filter PR created by ti-chi-bot
		if strings.Contains(*pr.User.Login, "ti-chi-bot") {
			fmt.Printf("repo:%s filter creates PR by bot, url:%s, title:%s \n", repo.Name, pr.GetHTMLURL(), pr.GetTitle())
			continue
		}
		// filter out the cfg.packages
		isContainPkg, err := isInPackages(client, repo.Packages, pr)
		if err != nil {
			return err
		}
		if !isContainPkg {
			fmt.Printf("repo:%s filter doesn't contain pkgs:%s, url:%s, title:%s \n",
				repo.Name, repo.Packages, pr.GetHTMLURL(), pr.GetTitle())
			continue
		}

		prs.WriteString(fmt.Sprintf("%s %s\n",
			markdown.Link(fmt.Sprintf("#%d", *pr.Number), *pr.HTMLURL),
			markdown.Escape(*pr.Title),
		))
	}
	if prs.Len() != 0 {
		buf.WriteString(fmt.Sprintf("## %s\n", markdown.Escape(repo.Name)))
		buf.WriteString(prs.String())
	}
	return nil
}
