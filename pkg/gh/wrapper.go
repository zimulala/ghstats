// Copyright 2021 ghstats Project Authors. Licensed under MIT.

package gh

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v35/github"
	log "github.com/sirupsen/logrus"
)

// GetRepository returns repository owner and name of the issue.
func GetRepository(issue *github.Issue) (owner, repo string) {
	repositoryURL := *issue.RepositoryURL
	parts := strings.Split(repositoryURL, "/")
	return parts[len(parts)-2], parts[len(parts)-1]
}

// GetPRRepository returns a PR's repository owner and name of the issue.
func GetPRRepository(pr *github.PullRequest) (owner, repo string) {
	repositoryURL := *pr.URL
	parts := strings.Split(repositoryURL, "/")
	return parts[len(parts)-4], parts[len(parts)-3]
}

// SearchIssues wraps Search.Issues, supports pagination and rate limit.
func SearchIssues(
	ctx context.Context, client *github.Client, query string,
) ([]*github.IssuesSearchResult, error) {
	results := make([]*github.IssuesSearchResult, 0)
	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{Page: 0},
	}
PAGINATION:
	for {
	RATELIMIT:
		for {
			result, resp, err := client.Search.Issues(ctx, query, opts)
			if rateLimited, err := handleAPIError(err); err != nil {
				return nil, err
			} else if rateLimited {
				continue
			}
			if resp.StatusCode != http.StatusOK {
				body, _ := ioutil.ReadAll(resp.Body)
				return nil, fmt.Errorf("search issue error [%d] %s", resp.StatusCode, string(body))
			}
			results = append(results, result)
			if resp.NextPage == 0 {
				break PAGINATION
			}
			opts.Page = resp.NextPage
			break RATELIMIT
		}
	}
	return results, nil
}

// IssuesListComments wraps Issues.ListComments, supports pagination and rate limit.
func IssuesListComments(
	ctx context.Context, client *github.Client, owner, repo string, number int, since *time.Time,
) ([]*github.IssueComment, error) {
	comments := make([]*github.IssueComment, 0)
	opts := &github.IssueListCommentsOptions{
		Since:       since,
		ListOptions: github.ListOptions{Page: 0},
	}
PAGINATION:
	for {
	RATELIMIT:
		for {
			result, resp, err := client.Issues.ListComments(
				ctx, owner, repo, number, opts)
			if rateLimited, err := handleAPIError(err); err != nil {
				return nil, err
			} else if rateLimited {
				continue
			}
			if resp.StatusCode != http.StatusOK {
				body, _ := ioutil.ReadAll(resp.Body)
				return nil, fmt.Errorf("issue list comments error [%d] %s", resp.StatusCode, string(body))
			}
			comments = append(comments, result...)
			if resp.NextPage == 0 {
				break PAGINATION
			}
			opts.Page = resp.NextPage
			break RATELIMIT
		}
	}
	return comments, nil
}

// PullRequestsListReviews wraps PullRequests.ListReviews,
// supports pagination and rate limit.
func PullRequestsListReviews(
	ctx context.Context, client *github.Client, owner, repo string, number int,
) ([]*github.PullRequestReview, error) {
	reviews := make([]*github.PullRequestReview, 0)
	opts := &github.ListOptions{Page: 0}
PAGINATION:
	for {
	RATELIMIT:
		for {
			result, resp, err := client.PullRequests.ListReviews(
				ctx, owner, repo, number, opts)
			if rateLimited, err := handleAPIError(err); err != nil {
				return nil, err
			} else if rateLimited {
				continue
			}
			if resp.StatusCode != http.StatusOK {
				body, _ := ioutil.ReadAll(resp.Body)
				return nil, fmt.Errorf("issue list comments error [%d] %s", resp.StatusCode, string(body))
			}
			reviews = append(reviews, result...)
			if resp.NextPage == 0 {
				break PAGINATION
			}
			opts.Page = resp.NextPage
			break RATELIMIT
		}
	}
	return reviews, nil
}

// PullRequestsListReviewComments wraps PullRequests.ListReviewComments,
// supports pagination and rate limit.
func PullRequestsListReviewComments(
	ctx context.Context, client *github.Client, owner, repo string, number int, reviewID int64,
) ([]*github.PullRequestComment, error) {
	comments := make([]*github.PullRequestComment, 0)
	opts := &github.ListOptions{Page: 0}
PAGINATION:
	for {
	RATELIMIT:
		for {
			result, resp, err := client.PullRequests.ListReviewComments(
				ctx, owner, repo, number, reviewID, opts)
			if rateLimited, err := handleAPIError(err); err != nil {
				return nil, err
			} else if rateLimited {
				continue
			}
			if resp.StatusCode != http.StatusOK {
				body, _ := ioutil.ReadAll(resp.Body)
				return nil, fmt.Errorf(
					"pull request review comments error [%d] %s",
					resp.StatusCode, string(body))
			}
			comments = append(comments, result...)
			if resp.NextPage == 0 {
				break PAGINATION
			}
			opts.Page = resp.NextPage
			break RATELIMIT
		}
	}
	return comments, nil
}

// IssuesListIssueEvents wraps Issues.ListIssueEvents,
// supports pagination and rate limit.
func IssuesListIssueEvents(
	ctx context.Context, client *github.Client, owner, repo string, number int,
) ([]*github.IssueEvent, error) {
	events := make([]*github.IssueEvent, 0)
	opts := &github.ListOptions{Page: 0}
PAGINATION:
	for {
	RATELIMIT:
		for {
			result, resp, err := client.Issues.ListIssueEvents(
				ctx, owner, repo, number, opts)
			if rateLimited, err := handleAPIError(err); err != nil {
				return nil, err
			} else if rateLimited {
				continue
			}
			if resp.StatusCode != http.StatusOK {
				body, _ := ioutil.ReadAll(resp.Body)
				return nil, fmt.Errorf("issue list comments error [%d] %s", resp.StatusCode, string(body))
			}
			events = append(events, result...)
			if resp.NextPage == 0 {
				break PAGINATION
			}
			opts.Page = resp.NextPage
			break RATELIMIT
		}
	}
	return events, nil
}

// PullRequestsList wraps PullRequests.List,
// supports pagination and rate limit.
func PullRequestsList(
	ctx context.Context, client *github.Client, owner, repo string,
) ([]*github.PullRequest, error) {
	prs := make([]*github.PullRequest, 0)
	opts := github.ListOptions{Page: 0}
	prOpts := &github.PullRequestListOptions{
		State:       "all",
		Sort:        "created",
		Direction:   "asc",
		ListOptions: opts,
	}

PAGINATION:
	for {
	RATELIMIT:
		for {
			result, resp, err := client.PullRequests.List(
				ctx, owner, repo, prOpts)
			if rateLimited, err := handleAPIError(err); err != nil {
				return nil, err
			} else if rateLimited {
				continue
			}
			if resp.StatusCode != http.StatusOK {
				body, _ := ioutil.ReadAll(resp.Body)
				return nil, fmt.Errorf("issue list comments error [%d] %s", resp.StatusCode, string(body))
			}
			prs = append(prs, result...)
			if resp.NextPage == 0 {
				break PAGINATION
			}
			prOpts.ListOptions.Page = resp.NextPage
			break RATELIMIT
		}
	}
	return prs, nil
}

// PullRequestsListFiles wraps PullRequests.ListFiles,
// supports pagination and rate limit.
func PullRequestsListFiles(
	ctx context.Context, client *github.Client, owner, repo string, number int,
) ([]*github.CommitFile, error) {
	files := make([]*github.CommitFile, 0)
	opts := &github.ListOptions{Page: 0}
PAGINATION:
	for {
	RATELIMIT:
		for {
			result, resp, err := client.PullRequests.ListFiles(
				ctx, owner, repo, number, opts)
			if rateLimited, err := handleAPIError(err); err != nil {
				return nil, err
			} else if rateLimited {
				continue
			}
			if resp.StatusCode != http.StatusOK {
				body, _ := ioutil.ReadAll(resp.Body)
				return nil, fmt.Errorf("issue list comments error [%d] %s", resp.StatusCode, string(body))
			}
			files = append(files, result...)
			if resp.NextPage == 0 {
				break PAGINATION
			}
			opts.Page = resp.NextPage
			break RATELIMIT
		}
	}
	return files, nil
}

func handleAPIError(err error) (rateLimited bool, e error) {
	if err == nil {
		return false, nil
	}
	if rateLimit, ok := err.(*github.RateLimitError); ok {
		dur := rateLimit.Rate.Reset.Sub(time.Now()) + 100*time.Millisecond
		log.Warnf("hit rate limit, sleep %s", dur)
		time.Sleep(dur)
		return true, nil
	}
	return false, err
}
