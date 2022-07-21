package main

import (
	"fmt"
	"time"

	"github.com/ooojustin/pr-puller/pkg/database"
	pr_gh "github.com/ooojustin/pr-puller/pkg/github"
	"github.com/ooojustin/pr-puller/pkg/slack"
	"github.com/ooojustin/pr-puller/pkg/utils"
)

const LineSeperator string = "-----------------------------------------\n"

type PrSlacker struct {
	cfg   *utils.Config
	db    *database.Database
	ghc   *pr_gh.GithubClient
	slack *slack.Slack
}

func (prs *PrSlacker) Run() {
	prs.processPullRequests(true)
	prs.startPullRequestTicker(3 * time.Minute)
	fmt.Scanln()
}

func (prs *PrSlacker) processPullRequests(all bool) {
	var org string = prs.cfg.GithubOrganization

	var pullRequests []*pr_gh.PullRequest
	if all {
		// Process all pages of pull requests
		prs.ghc.GetAllPullRequests(org, true, &pullRequests)
	} else {
		// Process first page of pull requests (25 most recent)
		prs.ghc.GetPullRequests(nil, 1, org, true, &pullRequests)
	}

	fmt.Printf("Loaded %d PullRequests\n", len(pullRequests))

	pprr := prs.db.PutPullRequests(pullRequests)
	prs.slack.SendPullRequestMessages(pprr.Notify)

	fmt.Printf("Uploaded: %d, Updated: %d, Skipped: %d, Failed: %d, Notified: %d\n",
		len(pprr.Uploaded), len(pprr.Updated), len(pprr.Skipped), len(pprr.Failed), len(pprr.Notify))

	fmt.Println(LineSeperator)
}

func (prs *PrSlacker) startPullRequestTicker(d time.Duration) {
	ticker := time.NewTicker(d)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				prs.processPullRequests(false)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
