package utils

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Org             string `json:"org"`
	SaveCookies     bool   `json:"save_cookies"`
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
