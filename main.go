package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	pr_gh "github.com/ooojustin/pr-puller/pkg/github"
)

func main() {
	cfgBytes, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic("failed to read account file")
	}

	var config map[string]interface{}
	err = json.Unmarshal(cfgBytes, &config)
	if err != nil {
		panic("failed to parse account credentials")
	}

	username := config["username"].(string)
	password := config["password"].(string)
	org := config["org"].(string)
	saveCookies := config["save_cookies"].(bool)

	ghc, ok := pr_gh.NewGithubClient(username, password, saveCookies)
	if !ok {
		panic("failed to initialize github client")
	}

	var login bool = ghc.Login()
	if !login {
		print("login: ", login)
	}

	var pullRequests []*pr_gh.PullRequest
	ghc.GetPullRequests(org, true, &pullRequests)

	fmt.Println("LOADED PRS:", len(pullRequests))
	for _, pr := range pullRequests {
		fmt.Println(pr.Title, pr.Href, pr.Labels)
	}
}
