package github

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/ooojustin/pr-puller/pkg/utils"
)

var (
	MissingHiddenValueError = errors.New("Missing hidden value.")
	InvalidCredentialsError = errors.New("Invalid credentials.")
	Failed2FAError          = errors.New("Failed two factor authenticationship.")
)

func (ghc *GithubClient) Login() error {
	resp, err := ghc.client.Get(GITHUB_URL + "login")
	if err != nil {
		return err
	}

	if resp.StatusCode == 302 {
		if locationUrl, err := resp.Location(); err == nil {
			if locationUrl.String() == GITHUB_URL {
				// Already logged in with this session
				return nil
			}
		}
	}

	body, err := utils.GetResponseBody(resp)
	if err != nil {
		return err
	}

	authenticity_token, ok1 := utils.FindHiddenValue("authenticity_token", body)
	timestamp_secret, ok2 := utils.FindHiddenValue("timestamp_secret", body)
	timestamp, ok3 := utils.FindHiddenValue("timestamp", body)
	if !(ok1 && ok2 && ok3) {
		return MissingHiddenValueError
	}

	exp := regexp.MustCompile("<input type=\"text\" name=\"(required_field_.+?)\"")
	fss := exp.FindStringSubmatch(body)
	if len(fss) != 2 {
		return MissingHiddenValueError
	}
	required_field := fss[1]

	fmt.Println("authenticity_token:", authenticity_token)
	fmt.Println("timestamp_secret:", timestamp_secret)
	fmt.Println("timestamp:", timestamp)
	fmt.Println("required_field:", required_field)

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
	data.Add(required_field, "")

	resp, err = ghc.client.PostForm(GITHUB_URL+"session", data)
	if err != nil {
		return err
	}

	body, err = utils.GetResponseBody(resp)
	if err != nil {
		return err
	} else if strings.Contains(body, "Incorrect username or password.") {
		return InvalidCredentialsError
	}

	locationUrlObj, _ := resp.Location()
	locationUrlStr := locationUrlObj.String()

	_2fa := strings.HasSuffix(locationUrlStr, "two-factor")
	if _2fa && !ghc.handle2FA(locationUrlStr) {
		return Failed2FAError
	}

	ghc.client.Jar.(*cookiejar.Jar).Save()
	return nil
}

func (ghc *GithubClient) handle2FA(locationUrl string) bool {
	for {
		resp2fa, err := ghc.client.Get(locationUrl)
		if err != nil {
			fmt.Printf("Error loading %s:\n%s\n", locationUrl, err)
			return false
		}

		body2fa, err := utils.GetResponseBody(resp2fa)
		if err != nil {
			fmt.Printf("Failed to read response body from %s:\n%s\n", locationUrl, err)
			return false
		}

		authenticity_token, ok := utils.FindHiddenValue("authenticity_token", body2fa)
		if !ok {
			fmt.Println("Failed to find hidden value 'authenticity_token' (GithubClient.handle2FA)")
			return false
		}

		otp, err := utils.ReadPassword("2FA Code")
		if err != nil {
			fmt.Println("Failed to read 2FA code input.")
			return false
		}

		data2fa := url.Values{}
		data2fa.Add("authenticity_token", authenticity_token)
		data2fa.Add("otp", otp)

		resp, err := ghc.client.PostForm(locationUrl, data2fa)
		if err != nil {
			fmt.Println("2FA Failed for an unknown. Please try again.")
			return false
		}

		if resp.StatusCode == 302 {
			// It tried to redirect us, aka login succeeded
			fmt.Println("Success! you are now logged into Github.")
			return true
		} else if resp.StatusCode == 200 {
			// We are on the same page, aka it failed
			fmt.Println("You've entered the incorrect 2FA code. Please try again.")
		}
	}
}
