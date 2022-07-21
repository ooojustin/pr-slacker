package github

import (
	"fmt"
	"net/http"

	cookiejar "github.com/juju/persistent-cookiejar"
)

const GITHUB_URL string = "https://github.com/"

type GithubClient struct {
	username string
	password string
	client   *http.Client
}

func NewGithubClient(username string, password string, saveCookies bool) (*GithubClient, bool) {
	opts := &cookiejar.Options{}
	if saveCookies {
		opts.Filename = fmt.Sprintf("./%s.json", username)
	} else {
		opts.NoPersist = true
	}

	jar, err := cookiejar.New(opts)
	if err != nil {
		return nil, false
	}

	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &GithubClient{
		username: username,
		password: password,
		client:   client,
	}, true
}
