package github

import (
	"net/http"

	cookiejar "github.com/juju/persistent-cookiejar"
)

const GITHUB_URL string = "https://github.com/"

type GithubClient struct {
	username string
	password string
	client   *http.Client
}

func NewGithubClient(username string, password string) (*GithubClient, bool) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, false
	}

	client := &http.Client{
		Jar: jar,
	}

	return &GithubClient{
		username: username,
		password: password,
		client:   client,
	}, true
}
