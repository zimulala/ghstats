// Copyright 2021 ghstats Project Authors. Licensed under MIT.

package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v35/github"
	"github.com/overvenus/ghstats/pkg/config"
	"github.com/overvenus/ghstats/pkg/gh"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

func init() {
	rootCmd.AddCommand(newReviewCommand())
}

// newReviewCommand returns REVIEW command
func newReviewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "review",
		Short: "Collect reviews ❤️",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			cfg1, err := config.ReadConfig(cfgPath)
			if err != nil {
				return err
			}
			cfg := cfg1.Review
			ctx := context.Background()
			client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: cfg.GithubToken},
			)))

			// Only collect 1 day review activity.
			// Date if formated in time.RFC3339.
			// updated:2021-05-23T21:00:00+08:00..2021-05-24T21:00:00+08:00
			today := time.Now().Format(time.RFC3339)
			yesterday := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
			updateRange := fmt.Sprintf(" updated:%s..%s", yesterday, today)
			projects := make(map[string][]*github.IssuesSearchResult)
			for _, proj := range cfg.Repos {
				for _, query := range proj.PRQuery {
					query = strings.TrimSpace(query)
					query += updateRange
					cmd.Println("query", query)
					results, err := gh.SearchIssues(ctx, client, query)
					if err != nil {
						return err
					}
					projects[proj.Name] = append(projects[proj.Name], results...)
				}
			}
			todayTimestamp, _ := time.Parse(time.RFC3339, today)
			yesterdayTimestamp, _ := time.Parse(time.RFC3339, yesterday)
			reviews := make(map[string]review)
			c := &reviewConfig{
				lgtmComments:   cfg.LGTMComments,
				blockComments:  cfg.BlockComments,
				blockUsers:     cfg.BlockUsers,
				startTimestamp: yesterdayTimestamp,
				endTimestamp:   todayTimestamp,
			}
			// println("debug projests issues", debug.PrettyFormat(projects))
			for repo, results := range projects {
				_ = repo
				for _, res := range results {
					if len(res.Issues) == 0 {
						continue
					}
					err := collectReviews(ctx, c, client, res.Issues, reviews)
					if err != nil {
						panic(err)
					}
				}
			}

			for user, review := range reviews {
				cmd.Println(user, "reviews", fmt.Sprintf("%#v", review))
			}
			return nil
			// bot := feishu.WebhookBot(cfg.FeishuWebhookToken)
			// return bot.SendMarkdownMessage(ctx, "PTAL ❤️", buf.String())
		},
	}
	return command
}

type review struct {
	// How many LTGM does one send?
	prLGTMs int
	// How many PR comments does one send?
	prComments int
	// How many issue comments does one send?
	issueComments int
	// How many issues does one create?
	issueCreates int
	// How many labels does one add?
	addLabels int
}

type reviewConfig struct {
	lgtmComments   []string
	blockComments  []string
	blockUsers     []string
	startTimestamp time.Time
	endTimestamp   time.Time
}

// Is the ts within [start, end)?
func (c *reviewConfig) withinTimeRange(ts time.Time) bool {
	return (ts.After(c.startTimestamp) || ts.Equal(c.startTimestamp)) && ts.Before(c.endTimestamp)
}

func (c *reviewConfig) isUserBlocked(userLogin string) bool {
	for i := range c.blockUsers {
		if c.blockUsers[i] == userLogin {
			return true
		}
	}
	return false
}

func (c *reviewConfig) isCommentBlocked(comment string) bool {
	// Unescapes common whitespace in github comments.
	comment = strings.ReplaceAll(comment, "\\n", "\n")
	comment = strings.ReplaceAll(comment, "\\r", "\r")
	comment = strings.ReplaceAll(comment, "\\t", "\t")
	lines := strings.Split(comment, "\n")
	for j := range lines {
		line := strings.TrimSpace(lines[j])
		for i := range c.blockComments {
			if strings.Contains(line, c.blockComments[i]) {
				return true
			}
		}
	}
	return false
}

func (c *reviewConfig) isCommentLGTM(comment string) bool {
	// Unescapes common whitespace in github comments.
	comment = strings.ReplaceAll(comment, "\\n", "\n")
	comment = strings.ReplaceAll(comment, "\\r", "\r")
	comment = strings.ReplaceAll(comment, "\\t", "\t")
	lines := strings.Split(comment, "\n")
	for j := range lines {
		line := strings.TrimSpace(lines[j])
		for i := range c.lgtmComments {
			if c.lgtmComments[i] == line {
				return true
			}
		}
	}
	return false
}

type collector func(
	ctx context.Context, c *reviewConfig, client *github.Client, issues []*github.Issue, reviews map[string]review,
) error

// collect reviews for the given issues and PRs.
func collectReviews(
	ctx context.Context, c *reviewConfig, client *github.Client, issues []*github.Issue, reviews map[string]review,
) error {
	collectors := []collector{
		collectIssueCreates,
		collectPRLGTM,
		collectPRReviewComments,
		collectIssueAndPRComments,
	}
	for _, collect := range collectors {
		err := collect(ctx, c, client, issues, reviews)
		if err != nil {
			return err
		}
	}
	return nil
}

// Collect review.prLGTM.
// LGTM is an APPROVED PR review or a review summary is LGTM.
func collectPRLGTM(
	ctx context.Context, c *reviewConfig, client *github.Client, issues []*github.Issue, reviews map[string]review,
) error {
	for _, issue := range issues {
		if !issue.IsPullRequest() {
			continue
		}
		pr := issue
		owner, repo := gh.GetRepository(pr)
		number := pr.GetNumber()
		prReviews, err := gh.PullRequestsListReviews(ctx, client, owner, repo, number)
		if err != nil {
			return err
		}
		// println("debug reviews", debug.PrettyFormat(prReviews))
		for _, prReview := range prReviews {
			if c.isUserBlocked(*prReview.User.Login) {
				continue
			}
			if !c.withinTimeRange(*prReview.SubmittedAt) {
				continue
			}
			if *prReview.State == "APPROVED" || c.isCommentLGTM(*prReview.Body) {
				review := reviews[*prReview.User.Login]
				review.prLGTMs++
				reviews[*prReview.User.Login] = review
			}
		}
	}
	return nil
}

// Collect review.prComments.
func collectPRReviewComments(
	ctx context.Context, c *reviewConfig, client *github.Client, issues []*github.Issue, reviews map[string]review,
) error {
	for _, issue := range issues {
		if !issue.IsPullRequest() {
			continue
		}
		pr := issue
		owner, repo := gh.GetRepository(pr)
		number := pr.GetNumber()
		prReviews, err := gh.PullRequestsListReviews(ctx, client, owner, repo, number)
		if err != nil {
			return err
		}
		// println("debug reviews", debug.PrettyFormat(prReviews))
		for _, prReview := range prReviews {
			if c.isUserBlocked(*prReview.User.Login) {
				continue
			}
			if !c.withinTimeRange(*prReview.SubmittedAt) {
				continue
			}

			reviewComments, err := gh.PullRequestsListReviewComments(ctx, client, owner, repo, number, *prReview.ID)
			if err != nil {
				return err
			}
			review := reviews[*prReview.User.Login]
			review.prComments += len(reviewComments)
			reviews[*prReview.User.Login] = review
		}
	}

	return nil
}

// Collect review.issueComments and review.prComments.
// Also, collect review.prLGTM if a comment is LGTM.
func collectIssueAndPRComments(
	ctx context.Context, c *reviewConfig, client *github.Client, issues []*github.Issue, reviews map[string]review,
) error {
	for _, issue := range issues {
		owner, repo := gh.GetRepository(issue)
		number := issue.GetNumber()
		comments, err := gh.IssuesListComments(
			ctx, client, owner, repo, number, &c.startTimestamp)
		if err != nil {
			return err
		}
		// println("debug comments", debug.PrettyFormat(comments))
		for _, comment := range comments {
			if c.isUserBlocked(*comment.User.Login) {
				continue
			}
			if c.isCommentBlocked(*comment.Body) {
				continue
			}
			if c.withinTimeRange(*comment.CreatedAt) || c.withinTimeRange(*comment.UpdatedAt) {
				review := reviews[*comment.User.Login]
				if issue.IsPullRequest() {
					if c.isCommentLGTM(*comment.Body) {
						review.prLGTMs++
					} else {
						review.prComments++
					}
				} else {
					review.issueComments++
				}
				reviews[*comment.User.Login] = review
			}
		}
	}
	return nil
}

// Collect review.issueCreates.
func collectIssueCreates(
	ctx context.Context, c *reviewConfig, client *github.Client, issues []*github.Issue, reviews map[string]review,
) error {
	for _, issue := range issues {
		if c.isUserBlocked(*issue.User.Login) {
			continue
		}
		if c.withinTimeRange(*issue.CreatedAt) {
			review := reviews[*issue.User.Login]
			if !issue.IsPullRequest() {
				review.issueCreates++
			}
			reviews[*issue.User.Login] = review
		}
	}
	return nil
}

// Collect review.addLabes.
func collectAddLabels(
	ctx context.Context, c *reviewConfig, client *github.Client, issues []*github.Issue, reviews map[string]review,
) error {
	for _, issue := range issues {
		if c.isUserBlocked(*issue.User.Login) {
			continue
		}
		if c.withinTimeRange(*issue.CreatedAt) {
			review := reviews[*issue.User.Login]
			if !issue.IsPullRequest() {
				review.issueCreates++
			}
			reviews[*issue.User.Login] = review
		}
	}
	return nil
}