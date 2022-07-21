package main

import (
	"fmt"

	"github.com/ooojustin/pr-puller/pkg/database"
	pr_gh "github.com/ooojustin/pr-puller/pkg/github"
	"github.com/ooojustin/pr-puller/pkg/slack"
	"github.com/ooojustin/pr-puller/pkg/utils"
)

func main() {

	// Load config variables from file.
	cfg, ok := utils.GetConfig()
	if !ok {
		panic("Failed to load config.")
	}

	// Initialize database connection.
	db, ok := database.Initialize()
	if !ok {
		panic("Failed to initialize database client.")
	}
	fmt.Println("database:", db)

	// Initialize slack client used to send messages.
	slackClient, ok := slack.Initialize()
	if !ok {
		panic("Failed to initialize slack client.")
	}

	// Initialize client used to access github.
	ghc, ok := pr_gh.NewGithubClient(
		cfg.GithubUsername,
		cfg.GithubPassword,
		cfg.GithubSaveCookies,
	)
	if !ok {
		panic("Failed to initialize github client.")
	}

	// Login to github via the client
	var login bool = ghc.Login()
	fmt.Println("Login: ", login, "\n", LineSeperator)
	if !login {
		panic("Failed to login.")
	}

	prs := &PrSlacker{
		db:    db,
		ghc:   ghc,
		cfg:   cfg,
		slack: slackClient,
	}
	prs.Run()
}
