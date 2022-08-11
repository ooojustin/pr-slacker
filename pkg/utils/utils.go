package utils

import (
	"bufio"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/html"
)

func ReadString(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", prompt)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(response), nil
}

func ReadPassword(prompt string) (string, error) {
	var stdin int = int(syscall.Stdin)
	fmt.Printf("%s: ", prompt)
	pwdBytes, err := terminal.ReadPassword(stdin)
	if err != nil {
		return "", err
	}
	fmt.Println()
	pwd := string(pwdBytes)
	return strings.TrimSpace(pwd), nil
}

func GetResponseBody(r *http.Response) (string, error) {
	defer r.Body.Close()

	rbytes, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	return string(rbytes), nil
}

func FindHiddenValue(key string, pageSource string) (string, bool) {
	pattern := fmt.Sprintf("<input type=\"hidden\" name=\"%s\" value=\"(.+?)\"", key)
	exp := regexp.MustCompile(pattern)

	fss := exp.FindStringSubmatch(pageSource)
	if len(fss) != 2 {
		// Expect the whole match (1) and the grouped value (2)
		return "", false
	}

	return fss[1], true
}

func GetAttribute(node *html.Node, key string) (string, bool) {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}

func AddFormField(w *multipart.Writer, key string, value string) error {
	writer, err := w.CreateFormField(key)
	if err != nil {
		return err
	}
	reader := strings.NewReader(value)
	_, err = io.Copy(writer, reader)
	return err
}
