package github

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/ooojustin/pr-puller/pkg/utils"
)

func (ghc *GithubClient) Login() bool {
	resp, err := ghc.client.Get(GITHUB_URL + "login")
	if err != nil {
		return false
	}

	if resp.StatusCode == 302 {
		if locationUrl, err := resp.Location(); err == nil {
			if locationUrl.String() == GITHUB_URL {
				// Already logged in
				return true
			}
		}
	}

	body, ok := utils.GetResponseBody(resp)
	if !ok {
		return false
	}

	authenticity_token, ok1 := utils.FindHiddenValue("authenticity_token", body)
	timestamp_secret, ok2 := utils.FindHiddenValue("timestamp_secret", body)
	timestamp, ok3 := utils.FindHiddenValue("timestamp", body)
	if !(ok1 && ok2 && ok3) {
		return false
	}

	exp := regexp.MustCompile("<input type=\"text\" name=\"(required_field_.+?)\"")
	fss := exp.FindStringSubmatch(body)
	if len(fss) != 2 {
		return false
	}
	rf := fss[1]

	fmt.Println("authenticity_token:", authenticity_token)
	fmt.Println("timestamp_secret:", timestamp_secret)
	fmt.Println("timestamp:", timestamp)
	fmt.Println("required_field:", rf)

	data := url.Values{}
	data.Add("login", ghc.username)
	data.Add("password", ghc.password)
	data.Add("authenticity_token", authenticity_token)
	data.Add("timestamp_secret", timestamp_secret)
	data.Add("timestamp", timestamp)
	data.Add("commit", "Sign in")
	data.Add("trusted_device", "")
	data.Add("webauthn-support", "supported")
	data.Add("webauthn-iuvpaa-support", "supported")
	data.Add("return_to", "https%3A%2F%2Fgithub.com%2Flogin")
	data.Add("allow_signup", "")
	data.Add("client_id", "")
	data.Add("integration", "")
	data.Add(rf, "")

	resp, err = ghc.client.PostForm(GITHUB_URL+"session", data)
	if err != nil {
		return false
	}

	body, _ = utils.GetResponseBody(resp)
	fail := strings.Contains(body, "Incorrect username or password.")
	if fail {
		return false
	}

	locationUrlObj, _ := resp.Location()
	locationUrlStr := locationUrlObj.String()
	_2fa := strings.HasSuffix(locationUrlStr, "two-factor")
	if _2fa {
		for {
			resp2fa, _ := ghc.client.Get(locationUrlStr)
			body2fa, _ := utils.GetResponseBody(resp2fa)
			authenticity_token, ok = utils.FindHiddenValue("authenticity_token", body2fa)

			reader := bufio.NewReader(os.Stdin)
			fmt.Print("2FA Code: ")
			otp, _ := reader.ReadString('\n')

			data2fa := url.Values{}
			data2fa.Add("authenticity_token", authenticity_token)
			data2fa.Add("otp", otp)

			resp, err := ghc.client.PostForm(locationUrlStr, data2fa)
			if err != nil {
				fmt.Println("2fa failed")
				return false
			}

			if resp.StatusCode == 200 {
				// We are on the same page, aka it failed
				fmt.Println("you entered the wrong 2fa code")
			} else if resp.StatusCode == 302 {
				// It tried to redirect us, aka login succeeded
				fmt.Println("success! you are now logged in")
				break
			}
		}
	}

	ghc.client.Jar.(*cookiejar.Jar).Save()
	return true
}
