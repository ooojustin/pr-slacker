package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func main() {
	accBytes, err := ioutil.ReadFile("./account.cfg")
	if err != nil {
		panic("failed to read account file")
	}

	var account map[string]interface{}
	err = json.Unmarshal(accBytes, &account)
	if err != nil {
		panic("failed to parse account credentials")
	}

	username := account["username"].(string)
	password := account["password"].(string)

	fmt.Println(username, password)
}
