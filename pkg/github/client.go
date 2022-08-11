package github

import (
	"fmt"
	"net/http"

	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/ooojustin/pr-puller/pkg/utils"
)

const GITHUB_URL string = "https://github.com/"

type GithubClient struct {
	username string
	password string
	client   *http.Client
}

func NewGithubClient(username string, password string, saveCookies bool, manualLogin bool) (*GithubClient, bool) {
	opts := &cookiejar.Options{}
	if saveCookies && !manualLogin {
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

	if manualLogin {
		var usrErr, pwdErr error

		username, usrErr = utils.ReadString("Username")
		if usrErr != nil {
			fmt.Println("Failed to read Github username.")
			return nil, false
		}

		password, pwdErr = utils.ReadPassword("Password")
		if pwdErr != nil {
			fmt.Println("Failed to read Github password.")
			return nil, false
		}
	}

	return &GithubClient{
		username: username,
		password: password,
		client:   client,
	}, true
}
