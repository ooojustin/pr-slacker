package main

import (
	"fmt"
	"time"

	"github.com/ooojustin/pr-puller/pkg/database"
	pr_gh "github.com/ooojustin/pr-puller/pkg/github"
	"github.com/ooojustin/pr-puller/pkg/utils"
)

const LineSeperator string = "-----------------------------------------\n"

type DependencyWrapper struct {
	db  *database.Database
	ghc *pr_gh.GithubClient
	cfg *utils.Config
}

func main() {

	// Load config variables from file
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

	// Initialize client used to access github
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
		db:  db,
		ghc: ghc,
		cfg: cfg,
	}

	dw.processPullRequests(true)
	dw.startPullRequestTicker(3 * time.Minute)

	fmt.Scanln()
}

func (dw *DependencyWrapper) processPullRequests(all bool) {
	var pullRequests []*pr_gh.PullRequest
	if all {
		dw.ghc.GetAllPullRequests(dw.cfg.Org, true, &pullRequests)
	} else {
		dw.ghc.GetPullRequests(nil, 1, dw.cfg.Org, true, &pullRequests)
	}
	fmt.Printf("Loaded %d PullRequests\n", len(pullRequests))
	counts := dw.db.PutPullRequests(pullRequests)
	fmt.Printf("Uploaded: %d, Updated: %d, Skipped: %d, Failed: %d\n",
		counts.Uploaded, counts.Updated, counts.Skipped, counts.Failed)
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
