package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/plugins/plugin-sdk"
	"github.com/google/go-github/v57/github"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type GitHubPlugin struct {
	token  string
	org    string
	repos  []string
	client *github.Client
}

// Helper: check rate limit and sleep if needed
func (p *GitHubPlugin) checkRateLimit(ctx context.Context, resp *github.Response) error {
	if resp == nil {
		return nil
	}

	remaining := resp.Rate.Remaining
	resetTime := resp.Rate.Reset.Time

	// If we're getting low on rate limit, wait until reset (with max duration)
	if remaining < 100 {
		sleepDuration := time.Until(resetTime)
		if sleepDuration > 0 {
			// Limit max sleep to prevent indefinite hangs
			// Use 20 seconds since parent timeout is 30s
			const maxSleep = 20 * time.Second
			if sleepDuration > maxSleep {
				return sdk.NewRateLimitError(fmt.Sprintf("rate limit exceeded, would need to sleep for %v (max %v)", sleepDuration, maxSleep))
			}
			fmt.Fprintf(os.Stderr, "Rate limit low (%d remaining), sleeping for %v until %s\n", remaining, sleepDuration, resetTime)

			// Sleep with context cancellation support
			select {
			case <-time.After(sleepDuration):
				// Sleep completed normally
			case <-ctx.Done():
				return fmt.Errorf("rate limit sleep cancelled: %w", ctx.Err())
			}
		}
	}

	return nil
}

// Helper: convert GitHub API errors to appropriate SDK error codes
func (p *GitHubPlugin) handleGitHubError(err error) error {
	if err == nil {
		return nil
	}

	// Check for GitHub error response
	if ghErr, ok := err.(*github.ErrorResponse); ok {
		switch ghErr.Response.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return sdk.NewAuthError(fmt.Sprintf("GitHub authentication failed: %v", err))
		case http.StatusTooManyRequests:
			return sdk.NewRateLimitError(fmt.Sprintf("GitHub rate limit exceeded: %v", err))
		case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return sdk.NewNetworkError(fmt.Sprintf("GitHub service unavailable: %v", err))
		}
	}

	// Network/timeout errors
	if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "connection") {
		return sdk.NewNetworkError(fmt.Sprintf("Network error: %v", err))
	}

	// Default to returning the original error (will be ErrCodePluginInternal)
	return err
}

func (p *GitHubPlugin) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		Name:        "github",
		Version:     "0.1.0",
		Type:        "source",
		Description: "GitHub Issues and Pull Requests",
		Author:      "EngineerDNA",
		ConfigFields: []sdk.ConfigField{
			{
				Name:        "token",
				Type:        "password",
				Required:    true,
				Description: "GitHub Personal Access Token",
				Secret:      true,
			},
			{
				Name:        "org",
				Type:        "string",
				Required:    true,
				Description: "Organization name",
			},
			{
				Name:        "repos",
				Type:        "string",
				Required:    false,
				Description: "Comma-separated repo names (empty = all repos)",
			},
		},
		Anonymization: sdk.AnonymizationSpec{
			Required: false,
			Strategy: "sequential",
			Fields:   []string{"actor", "author", "assignee"},
		},
	}
}

func (p *GitHubPlugin) Configure(config map[string]string) error {
	token, ok := config["token"]
	if !ok || token == "" {
		return fmt.Errorf("token is required")
	}
	p.token = token

	org, ok := config["org"]
	if !ok || org == "" {
		return fmt.Errorf("org is required")
	}
	p.org = org

	if reposStr, ok := config["repos"]; ok && reposStr != "" {
		parts := strings.Split(reposStr, ",")
		var validRepos []string
		for _, repo := range parts {
			trimmed := strings.TrimSpace(repo)
			if trimmed != "" {
				validRepos = append(validRepos, trimmed)
			}
		}
		p.repos = validRepos
	}

	// Initialize GitHub client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	p.client = github.NewClient(tc)

	return nil
}

func (p *GitHubPlugin) Health() sdk.HealthResult {
	if p.client == nil {
		return sdk.HealthResult{
			Healthy: false,
			Message: "Not configured",
		}
	}

	ctx := context.Background()
	_, _, err := p.client.Users.Get(ctx, "")
	if err != nil {
		return sdk.HealthResult{
			Healthy: false,
			Message: fmt.Sprintf("GitHub API error: %v", err),
		}
	}

	return sdk.HealthResult{
		Healthy: true,
		Message: "Connected to GitHub API",
	}
}

func (p *GitHubPlugin) Sync(params sdk.SyncParams) (sdk.SyncResult, error) {
	if p.client == nil {
		return sdk.SyncResult{}, fmt.Errorf("plugin not configured")
	}

	// Create context with timeout to prevent hanging on slow GitHub API
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var allEvents []sdk.Event

	// Get list of repos
	repos, err := p.getRepos(ctx)
	if err != nil {
		return sdk.SyncResult{}, fmt.Errorf("failed to get repos: %w", err)
	}

	// Fetch PRs and issues from each repo
	for _, repo := range repos {
		// Check context timeout/cancellation before processing each repo
		if err := ctx.Err(); err != nil {
			return sdk.SyncResult{}, fmt.Errorf("context cancelled/timed out: %w", err)
		}

		// Fetch Pull Requests
		prEvents, err := p.fetchPullRequests(ctx, repo, params.Since)
		if err != nil {
			return sdk.SyncResult{}, fmt.Errorf("failed to fetch PRs from %s: %w", repo, err)
		}
		allEvents = append(allEvents, prEvents...)

		// Fetch Issues
		issueEvents, err := p.fetchIssues(ctx, repo, params.Since)
		if err != nil {
			return sdk.SyncResult{}, fmt.Errorf("failed to fetch issues from %s: %w", repo, err)
		}
		allEvents = append(allEvents, issueEvents...)
	}

	return sdk.SyncResult{Events: allEvents}, nil
}

func (p *GitHubPlugin) getRepos(ctx context.Context) ([]string, error) {
	if len(p.repos) > 0 {
		return p.repos, nil
	}

	// List all repos in the org (with max limit to prevent unbounded resource usage)
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []string
	const maxRepos = 1000 // Prevent unbounded memory usage
	for {
		repos, resp, err := p.client.Repositories.ListByOrg(ctx, p.org, opt)
		if err != nil {
			return nil, p.handleGitHubError(err)
		}

		// Check rate limit
		if err := p.checkRateLimit(ctx, resp); err != nil {
			return nil, err
		}

		for _, repo := range repos {
			if repo.Name != nil {
				allRepos = append(allRepos, *repo.Name)
				// Enforce max repo limit
				if len(allRepos) >= maxRepos {
					fmt.Fprintf(os.Stderr, "Warning: reached max repo limit of %d, stopping enumeration\n", maxRepos)
					return allRepos, nil
				}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func (p *GitHubPlugin) fetchPullRequests(ctx context.Context, repo string, since time.Time) ([]sdk.Event, error) {
	opt := &github.PullRequestListOptions{
		State:       "all",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var events []sdk.Event
	const maxPRs = 10000 // Prevent unbounded fetching
	for {
		prs, resp, err := p.client.PullRequests.List(ctx, p.org, repo, opt)
		if err != nil {
			return nil, p.handleGitHubError(err)
		}

		// Check rate limit
		if err := p.checkRateLimit(ctx, resp); err != nil {
			return nil, err
		}

		for _, pr := range prs {
			// Skip PRs with missing required fields (prevent nil pointer panics)
			if pr.Number == nil || pr.CreatedAt == nil {
				continue
			}

			// Stop if we've reached PRs older than 'since'
			if pr.UpdatedAt != nil && pr.UpdatedAt.Before(since) {
				return events, nil
			}

			// Determine action
			action := "opened"
			timestamp := pr.CreatedAt.Time
			if pr.MergedAt != nil {
				action = "merged"
				timestamp = pr.MergedAt.Time
			} else if pr.ClosedAt != nil {
				action = "closed"
				timestamp = pr.ClosedAt.Time
			}

			// Only include if timestamp is after 'since'
			if timestamp.Before(since) {
				continue
			}

			actor := ""
			if pr.User != nil && pr.User.Login != nil {
				actor = *pr.User.Login
			}

			// Get reviewers
			reviewers := []string{}
			if pr.RequestedReviewers != nil {
				for _, reviewer := range pr.RequestedReviewers {
					if reviewer.Login != nil {
						reviewers = append(reviewers, *reviewer.Login)
					}
				}
			}

			event := sdk.Event{
				ID:        uuid.New().String(),
				Type:      "pull_request",
				Source:    "github",
				SourceID:  fmt.Sprintf("%s/%s#%d", p.org, repo, *pr.Number),
				Timestamp: timestamp,
				Actor:     actor,
				Data: map[string]interface{}{
					"action":        action,
					"number":        *pr.Number,
					"title":         safeString(pr.Title),
					"url":           safeString(pr.HTMLURL),
					"author":        actor,
					"repo":          fmt.Sprintf("%s/%s", p.org, repo),
					"additions":     safeInt(pr.Additions),
					"deletions":     safeInt(pr.Deletions),
					"changed_files": safeInt(pr.ChangedFiles),
					"reviewers":     reviewers,
				},
			}

			if pr.MergedAt != nil {
				event.Data["merged_at"] = pr.MergedAt.Format(time.RFC3339)
			}
			if pr.CreatedAt != nil {
				event.Data["created_at"] = pr.CreatedAt.Format(time.RFC3339)
			}

			events = append(events, event)

			// Enforce max PR limit
			if len(events) >= maxPRs {
				fmt.Fprintf(os.Stderr, "Warning: reached max PR limit of %d for repo %s, stopping fetch\n", maxPRs, repo)
				return events, nil
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return events, nil
}

func (p *GitHubPlugin) fetchIssues(ctx context.Context, repo string, since time.Time) ([]sdk.Event, error) {
	opt := &github.IssueListByRepoOptions{
		State:       "all",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var events []sdk.Event
	const maxIssues = 10000 // Prevent unbounded fetching
	for {
		issues, resp, err := p.client.Issues.ListByRepo(ctx, p.org, repo, opt)
		if err != nil {
			return nil, p.handleGitHubError(err)
		}

		// Check rate limit
		if err := p.checkRateLimit(ctx, resp); err != nil {
			return nil, err
		}

		for _, issue := range issues {
			// Skip pull requests (they're handled separately)
			if issue.IsPullRequest() {
				continue
			}

			// Skip issues with missing required fields (prevent nil pointer panics)
			if issue.Number == nil || issue.CreatedAt == nil {
				continue
			}

			// Stop if we've reached issues older than 'since'
			if issue.UpdatedAt != nil && issue.UpdatedAt.Before(since) {
				return events, nil
			}

			// Determine action
			action := "opened"
			timestamp := issue.CreatedAt.Time
			if issue.ClosedAt != nil {
				action = "closed"
				timestamp = issue.ClosedAt.Time
			}

			// Only include if timestamp is after 'since'
			if timestamp.Before(since) {
				continue
			}

			actor := ""
			if issue.User != nil && issue.User.Login != nil {
				actor = *issue.User.Login
			}

			assignee := ""
			if issue.Assignee != nil && issue.Assignee.Login != nil {
				assignee = *issue.Assignee.Login
			}

			labels := []string{}
			for _, label := range issue.Labels {
				if label.Name != nil {
					labels = append(labels, *label.Name)
				}
			}

			// Extract story points from labels (e.g., "story-points: 8", "points: 5", "sp: 3")
			storyPoints, hasStoryPoints := extractStoryPoints(labels)

			event := sdk.Event{
				ID:        uuid.New().String(),
				Type:      "issue",
				Source:    "github",
				SourceID:  fmt.Sprintf("%s/%s#%d", p.org, repo, *issue.Number),
				Timestamp: timestamp,
				Actor:     actor,
				Data: map[string]interface{}{
					"action":   action,
					"number":   *issue.Number,
					"title":    safeString(issue.Title),
					"url":      safeString(issue.HTMLURL),
					"author":   actor,
					"assignee": assignee,
					"repo":     fmt.Sprintf("%s/%s", p.org, repo),
					"labels":   labels,
				},
			}

			if issue.CreatedAt != nil {
				event.Data["created_at"] = issue.CreatedAt.Format(time.RFC3339)
			}
			if issue.ClosedAt != nil {
				event.Data["closed_at"] = issue.ClosedAt.Format(time.RFC3339)
			}

			// Add story points if found in labels
			if hasStoryPoints {
				event.Data["story_points"] = storyPoints
			}

			events = append(events, event)

			// Enforce max issue limit
			if len(events) >= maxIssues {
				fmt.Fprintf(os.Stderr, "Warning: reached max issue limit of %d for repo %s, stopping fetch\n", maxIssues, repo)
				return events, nil
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return events, nil
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func safeInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// extractStoryPoints parses story points from GitHub issue labels.
// Supports formats: "story-points: 8", "points: 5", "sp: 3", "8 points", "story points: 13"
// Returns (points, true) if found, or (0, false) if not found.
func extractStoryPoints(labels []string) (int, bool) {
	// Match patterns like:
	// - "story-points: 8", "story points: 13", "storypoints: 5"
	// - "points: 5", "pts: 3"
	// - "sp: 3"
	// - "8 points", "5 story points"
	re := regexp.MustCompile(`(?i)(?:story[-\s]?)?(?:points?|sp|pts?):?\s*(\d+)|(\d+)\s*(?:story[-\s]?)?points?`)

	for _, label := range labels {
		if matches := re.FindStringSubmatch(label); matches != nil {
			// First capture group: pattern like "story-points: 8"
			if matches[1] != "" {
				points, err := strconv.Atoi(matches[1])
				if err == nil && points > 0 {
					return points, true
				}
			}
			// Second capture group: pattern like "8 points"
			if matches[2] != "" {
				points, err := strconv.Atoi(matches[2])
				if err == nil && points > 0 {
					return points, true
				}
			}
		}
	}
	return 0, false
}

func main() {
	sdk.Serve(&GitHubPlugin{})
}
