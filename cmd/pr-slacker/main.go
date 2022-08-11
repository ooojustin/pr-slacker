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
		exit("Failed to load config.", 0)
	}

	// Initialize database connection.
	db, ok := database.Initialize()
	if !ok {
		exit("Failed to initialize database client.", 0)
	}

	// Initialize slack client used to send messages.
	slackClient, ok := slack.Initialize()
	if !ok {
		exit("Failed to initialize slack client.", 0)
	}

	// Initialize client used to access github.
	ghc, ok := pr_gh.NewGithubClient(
		cfg.GithubUsername,
		cfg.GithubPassword,
		cfg.GithubSaveCookies,
		cfg.GithubManualLogin,
	)
	if !ok {
		exit("Failed to initialize github client.", 0)
	}

	// Login to github via the client
	if login := ghc.Login(); !login {
		exit("Failed to login.", 0)
	}

	prs := &PrSlacker{
		db:    db,
		ghc:   ghc,
		cfg:   cfg,
		slack: slackClient,
	}
	prs.Run()
}

func exit(msg string, code int) {
	fmt.Println(msg)
	os.Exit(code)
}
