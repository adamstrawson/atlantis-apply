package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v38/github"
	"golang.org/x/oauth2"
)

func main() {

	repo := flag.String("repo", "", "GitHub repository name")
	token := flag.String("token", "", "GitHub PAT")

	// Parse command-line flags
	flag.Parse()

	if *repo == "" {
		// If not provided, check the GITHUB_REPO environment variable
		envRepo := os.Getenv("GITHUB_REPO")
		if envRepo != "" {
			*repo = envRepo
		} else {
			log.Fatal("Please provide -repo flag or set GITHUB_REPO environment variable")
		}
	}

	if *token == "" {
		// If not provided, check the GITHUB_TOKEN environment variable
		envToken := os.Getenv("GITHUB_TOKEN")
		if envToken != "" {
			*token = envToken
		} else {
			log.Fatal("Please provide -token flag or set GITHUB_TOKEN environment variable")
		}
	}

	parts := strings.Split(*repo, "/")
	if len(parts) != 2 {
		log.Fatal("Please provide repo in the format of \"Organisation/Repository\"")
		os.Exit(1)
	}

	owner := parts[0]
	*repo = parts[1]

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	label := "renovate"

	// List all pull requests in the repository
	pullRequests, _, err := client.PullRequests.List(ctx, owner, *repo, nil)
	if err != nil {
		log.Fatalf("Error listing pull requests: %v", err)
	}

	if len(pullRequests) == 0 {
		log.Fatal("No PRs found with that label")
	}

	// Iterate through pull requests
	for _, pr := range pullRequests {
		labels, _, err := client.Issues.ListLabelsByIssue(ctx, owner, *repo, *pr.Number, nil)
		if err != nil {
			log.Fatalf("Error getting labels for PR #%d: %v", *pr.Number, err)
		}

		// Check if the PR has the specified label
		hasLabel := false
		for _, l := range labels {
			if *l.Name == label {
				hasLabel = true
				break
			}
		}

		if hasLabel {
			fmt.Printf("Checking rebase status for PR #%d: %s\n", *pr.Number, *pr.Title)

			// Get the base branch of the PR
			baseBranch := *pr.Base.Ref

			// Check if the PR is up-to-date with the base branch
			opts := &github.ListOptions{Page: 1, PerPage: 10}
			compare, _, err := client.Repositories.CompareCommits(ctx, owner, *repo, baseBranch, *pr.Head.Ref, opts)
			if err != nil {
				log.Fatalf("Error comparing commits for PR #%d: %v", *pr.Number, err)
			}

			if *compare.Status != "ahead" {

				fmt.Printf("PR #%d is not up-to-date. Rebasing...\n", *pr.Number)

				prOpts := &github.PullRequestBranchUpdateOptions{}
				_, _, _ = client.PullRequests.UpdateBranch(ctx, owner, *repo, *pr.Number, prOpts)
				fmt.Printf("Waiting for PR #%d to rebase...\n", *pr.Number)
				time.Sleep(10 * time.Second)
			}

			fmt.Printf("PR #%d is up-to-date. Approving review.\n", *pr.Number)

			// Create a review approving the PR
			review := github.PullRequestReviewRequest{
				Event: github.String("APPROVE"),
			}
			_, _, err = client.PullRequests.CreateReview(ctx, owner, *repo, *pr.Number, &review)
			if err != nil {
				log.Fatalf("Error approving review for PR #%d: %v", *pr.Number, err)
			}

			// Comment on the PR
			comment := &github.IssueComment{
				Body: github.String("atlantis apply"),
			}
			_, _, err = client.Issues.CreateComment(ctx, owner, *repo, *pr.Number, comment)
			if err != nil {
				log.Fatalf("Error commenting on PR #%d: %v", *pr.Number, err)
			}

			// Wait for the PR to be merged
			fmt.Printf("Waiting for PR #%d to be merged...\n", *pr.Number)
			for {
				status, _, err := client.PullRequests.Get(ctx, owner, *repo, *pr.Number)
				if err != nil {
					log.Fatalf("Error getting PR status for PR #%d: %v", *pr.Number, err)
				}

				if status.GetMerged() {
					fmt.Printf("PR #%d has been merged.\n", *pr.Number)
					break
				}

				time.Sleep(10 * time.Second)
			}

		}
	}
}
