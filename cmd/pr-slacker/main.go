package main

import (
	"fmt"
	"os"

	"github.com/ooojustin/pr-puller/pkg/database"
	pr_gh "github.com/ooojustin/pr-puller/pkg/github"
	"github.com/ooojustin/pr-puller/pkg/slack"
	"github.com/ooojustin/pr-puller/pkg/utils"
)

func main() {
	// Load config variables from file.
	cfg, ok := utils.GetConfig()
	if !ok {
		exitf(0, "Failed to load config.")
	}

	// Initialize database connection.
	db, ok := database.Initialize()
	if !ok {
		exitf(0, "Failed to initialize database client.")
	}

	// Initialize slack client used to send messages.
	slackClient, ok := slack.Initialize()
	if !ok {
		exitf(0, "Failed to initialize slack client.")
	}

	// Initialize client used to access github.
	ghc, ok := pr_gh.NewGithubClient(
		cfg.GithubUsername,
		cfg.GithubPassword,
		cfg.GithubSaveCookies,
		cfg.GithubManualLogin,
	)
	if !ok {
		exitf(0, "Failed to initialize github client.")
	}

	// Login to github via the client
	if err := ghc.Login(); err != nil {
		exitf(0, "Failed to login to Github: %s", err)
	}

	prs := &PrSlacker{
		db:    db,
		ghc:   ghc,
		cfg:   cfg,
		slack: slackClient,
	}
	prs.Run()
}

func exitf(code int, format string, a ...interface{}) {
	fmt.Printf(format, a...)
	os.Exit(code)
}
