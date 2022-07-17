package main

import (
	"fmt"

	pr_gh "github.com/ooojustin/pr-puller/pkg/github"
	"github.com/ooojustin/pr-puller/pkg/utils"
)

func main() {
	// Load config variables from file
	cfg, ok := utils.GetConfig()
	if !ok {
		panic("failed to load config")
	}

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
	print("login: ", login)

	// Load pull requests
	var pullRequests []*pr_gh.PullRequest
	ghc.GetPullRequests(cfg.Org, true, &pullRequests)

	fmt.Println("LOADED PRS:", len(pullRequests))
	for _, pr := range pullRequests {
		fmt.Println(pr.Title, pr.Href, pr.Labels)
	}
}
