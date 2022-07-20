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

type DependencyWrapper struct {
	cfg   *utils.Config
	db    *database.Database
	ghc   *pr_gh.GithubClient
	slack *slack.Slack
}

func main() {

	// Load config variables from file.
	cfg, ok := utils.GetConfig()
	if !ok {
		panic("failed to load config")
	}

	// Initialize database connection.
	db, ok := database.Initialize()
	if !ok {
		panic("failed to initialize database client")
	}
	fmt.Println("database:", db)

	// Initialize slack client used to send messages.
	slackClient, ok := slack.Initialize()
	if !ok {
		panic("failed to initialize slack client")
	}

	// Initialize client used to access github.
	ghc, ok := pr_gh.NewGithubClient(
		cfg.Username,
		cfg.Password,
		cfg.SaveCookies,
	)
	if !ok {
		panic("failed to initialize github client")
	}

	// Login to github via the client
	var login bool = ghc.Login()
	fmt.Println("Login: ", login, "\n", LineSeperator)
	if !login {
		panic("failed to login")
	}

	dw := &DependencyWrapper{
		db:    db,
		ghc:   ghc,
		cfg:   cfg,
		slack: slackClient,
	}

	dw.processPullRequests(true)
	dw.startPullRequestTicker(3 * time.Minute)

	fmt.Scanln()
}

func (dw *DependencyWrapper) processPullRequests(all bool) {
	var pullRequests []*pr_gh.PullRequest
	if all {
		// Process all pages of pull requests
		dw.ghc.GetAllPullRequests(dw.cfg.Org, true, &pullRequests)
	} else {
		// Process first page of pull requests (25 most recent)
		dw.ghc.GetPullRequests(nil, 1, dw.cfg.Org, true, &pullRequests)
	}
	fmt.Printf("Loaded %d PullRequests\n", len(pullRequests))
	pprr := dw.db.PutPullRequests(pullRequests)
	dw.slack.SendPullRequestMessages(pprr.Notify)
	fmt.Printf("Uploaded: %d, Updated: %d, Skipped: %d, Failed: %d, Notified: %d\n",
		pprr.Uploaded, pprr.Updated, pprr.Skipped, pprr.Failed, len(pprr.Notify))
	fmt.Println(LineSeperator)
}

func (dw *DependencyWrapper) startPullRequestTicker(d time.Duration) {
	ticker := time.NewTicker(d)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				dw.processPullRequests(false)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
