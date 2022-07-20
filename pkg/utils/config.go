package utils

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	GithubUsername     string `json:"github_username"`
	GithubPassword     string `json:"github_password"`
	GithubOrganization string `json:"github_organization"`
	GithubSaveCookies  bool   `json:"github_save_cookies"`
	AwsAccessKeyID     string `json:"aws_access_key_id"`
	AwsAccessKeySecret string `json:"aws_access_key_secret"`
	SlackOauthToken    string `json:"slack_oauth_token"`
	SlackChannelID     string `json:"slack_channel_id"`
}

func GetConfig() (*Config, bool) {
	cfgBytes, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return nil, false
	}

	var cfg Config
	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		return nil, false
	}

	return &cfg, true
}
