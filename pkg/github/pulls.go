package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ooojustin/pr-puller/pkg/utils"
	"golang.org/x/net/html"
)

type PullRequest struct {
	PK             string    `json:"-" dynamodbav:"pr_uid"`
	ID             int       `json:"id" dynamodbav:"id"`
	Created        time.Time `json:"created" dynamodbav:"created"`
	Creator        string    `json:"creator" dynamodbav:"creator"`
	Repository     string    `json:"repository" dynamodbav:"repository"`
	Organization   string    `json:"organization" dynamodbav:"organization"`
	Title          string    `json:"title" dynamodbav:"title"`
	URL            string    `json:"url" dynamodbav:"url"`
	Labels         []string  `json:"labels" dynamodbav:"labels"`
	Draft          bool      `json:"draft" dynamodbav:"draft"`
	ReviewDecision string    `json:"review_decision" dynamodbav:"review_decision"`
	Number         int       `json:"number" dynamodbav:"number"`
}

func (pr PullRequest) ToString() (string, error) {
	prBytes, err := json.MarshalIndent(pr, "", "    ")
	if err != nil {
		return "", err
	}
	return string(prBytes), nil
}

// Generate all pull request objects for a given org.
func (ghc *GithubClient) GetPullRequests(
	org string,
	open bool,
	prs *[]*PullRequest,
) {
	// Load the first page
	doc, ok := ghc.loadPullRequestDocument(1, org, open)
	if !ok {
		return
	}

	// Determine the total number of pages
	maxPage, ok := getPageCount(doc)
	if !ok {
		return
	}

	// Extract pull request data from each page
	for i := 1; i <= maxPage; i++ {
		var document *goquery.Document
		if i == 1 {
			document = doc
		}

		ghc.loadPullRequests(document, i, org, open, prs)
	}
}

// Generate pull request objects from a given page.
func (ghc *GithubClient) loadPullRequests(
	doc *goquery.Document,
	page int,
	org string,
	open bool,
	prs *[]*PullRequest,
) {
	if doc == nil {
		// Download page HTML and process it as a goquery Document for parsing.
		var ok bool
		doc, ok = ghc.loadPullRequestDocument(page, org, open)
		if !ok {
			return
		}
	}

	// Parse document and extract data from nodes to generate PR objects for this page.
	var prsNew []*PullRequest
	prNodes := getPullRequestNodes(doc)
	for _, prNode := range prNodes {
		if pr, ok := generatePullRequestObject(prNode); ok {
			prsNew = append(prsNew, pr)
		}
	}

	// Use Github's hidden 'pull_request_review_decisions' endpoint to attach
	// review decisions to each of these PullRequest objects, if available.
	ghc.loadPullRequestReviewDecisions(&prsNew)

	// Add PRs from this page to the ongoing list.
	*prs = append(*prs, prsNew...)
}

// Download Github pull requests page HTML and process it as a goquery Document for parsing.
func (ghc *GithubClient) loadPullRequestDocument(page int, org string, open bool) (*goquery.Document, bool) {
	items := []string{}
	items = append(items, "org:"+org)
	if open {
		items = append(items, "is:open")
	}

	encodedItems := []string{}
	for _, item := range items {
		encodedItems = append(encodedItems, url.QueryEscape(item))
	}

	query := strings.Join(encodedItems, "+")

	pullsUrl := fmt.Sprintf("%spulls?page=%d&q=%s", GITHUB_URL, page, query)

	resp, err := ghc.client.Get(pullsUrl)
	if err != nil {
		return nil, false
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, false
	}

	return doc, true
}

// Generate a PullRequest object to represent a pull request given the primary node.
func generatePullRequestObject(prNode *html.Node) (*PullRequest, bool) {
	iconBoxNode := prNode.FirstChild.NextSibling.FirstChild.NextSibling
	iconNode := iconBoxNode.FirstChild.NextSibling

	lbl, _ := utils.GetAttribute(iconNode, "aria-label")
	if !strings.HasSuffix(lbl, "pull request") {
		// It's not actually a pull request
		return nil, false
	}

	draft := strings.Contains(lbl, "draft")
	aTagRepo := iconBoxNode.NextSibling.NextSibling.NextSibling.NextSibling.FirstChild.NextSibling
	repositoryPath := strings.TrimSpace(aTagRepo.FirstChild.Data)
	repositoryPathSplit := strings.Split(repositoryPath, "/")
	organization := repositoryPathSplit[0]
	repositoryName := repositoryPathSplit[1]
	aTagPR := aTagRepo.NextSibling.NextSibling
	prName := strings.TrimSpace(aTagPR.FirstChild.Data)
	href, _ := utils.GetAttribute(aTagPR, "href")
	hrefSplit := strings.Split(href, "/")
	number, _ := strconv.Atoi(hrefSplit[len(hrefSplit)-1])
	labels := getPullRequestLabels(aTagPR)
	opened := getPullRequestOpenedNode(aTagPR)
	datetimeNode := opened.FirstChild.NextSibling
	datetimeStr, _ := utils.GetAttribute(datetimeNode, "datetime")
	username := strings.TrimSpace(datetimeNode.NextSibling.NextSibling.FirstChild.Data)
	datetime, _ := time.Parse(time.RFC3339, datetimeStr)
	pk := fmt.Sprintf("%s#%s#%d", organization, repositoryName, number)
	url := GITHUB_URL + href

	var id int
	if !draft {
		idInput := opened.NextSibling.NextSibling.FirstChild.NextSibling.FirstChild.NextSibling
		if prId, ok := utils.GetAttribute(idInput, "value"); ok {
			id, _ = strconv.Atoi(prId)
		}
	}

	pr := &PullRequest{
		PK:           pk,
		ID:           id,
		Created:      datetime,
		Creator:      username,
		Repository:   repositoryName,
		Organization: organization,
		Title:        prName,
		URL:          url,
		Labels:       labels,
		Draft:        draft,
		Number:       number,
	}

	return pr, true
}

// Get the 'opened by' node, used as a starting point to find the created timestamp
// of the pull request and the username of the person who created it.
func getPullRequestOpenedNode(aTagPR *html.Node) *html.Node {
	for c := aTagPR.NextSibling; c != nil; c = c.NextSibling {
		class, _ := utils.GetAttribute(c, "class")
		if c.Data == "div" && strings.Contains(class, "color-fg-muted") {
			openedBy := c.FirstChild.NextSibling
			return openedBy
		}
	}
	return nil
}

// Get the labels that have been attached to a pull request.
func getPullRequestLabels(aTagPR *html.Node) []string {
	var labels []string
	for c := aTagPR.NextSibling; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "span" {
			class, _ := utils.GetAttribute(c, "class")
			if strings.Contains(class, "lh-default") {
				for labelNode := c.FirstChild; labelNode != nil; labelNode = labelNode.NextSibling {
					if labelNode.Type == html.ElementNode && labelNode.Data == "a" {
						label := strings.TrimSpace(labelNode.FirstChild.Data)
						labels = append(labels, label)
					}
				}
			}
		}
	}
	return labels
}

// Get a list of the nodes that each represent a pull request.
func getPullRequestNodes(doc *goquery.Document) []*html.Node {
	var nodes []*html.Node
	selection := doc.Find("div.js-navigation-container")
	for c := selection.Nodes[0].FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			nodes = append(nodes, c)
		}
	}
	return nodes
}

// Parse pagination nodes to determine total number of pages.
func getPageCount(doc *goquery.Document) (int, bool) {
	selection := doc.Find("div.pagination")
	pagination := selection.Nodes[0]
	lastPageNode := pagination.LastChild.PrevSibling.PrevSibling
	lastPageText, ok := utils.GetAttribute(lastPageNode, "aria-label")
	if !ok {
		return 0, false
	}
	pageNum := strings.Split(lastPageText, " ")[1]
	count, err := strconv.Atoi(pageNum)
	if err != nil {
		return 0, false
	}
	return count, true
}

// Loads pull request review decision for a list of PRs.
// This will POST multipart/form-data to a hidden endpoint, allowing us to
// efficiently determine the review decision for multiple PRs at the same time.
// Sample request body: https://pastebin.com/xvieweYs
func (ghc *GithubClient) loadPullRequestReviewDecisions(prs *[]*PullRequest) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// Default field added to each of these requests.
	utils.AddFormField(writer, "_method", "GET")

	// Add form field with each one of these IDs to request review decision.
	for idx, pr := range *prs {
		name := fmt.Sprintf("items[item-%d][pull_request_id]", idx)
		idStr := strconv.Itoa(pr.ID)
		utils.AddFormField(writer, name, idStr)
	}

	// Close multipart writer since all fields have been written.
	writer.Close()

	// Prepare request with buffer containing form data.
	url := GITHUB_URL + "pull_request_review_decisions"
	req, err := http.NewRequest("POST", url, &buffer)
	if err != nil {
		fmt.Println("failed to create request:", err)
		return
	}

	// Set important headers and execute request.
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	resp, err := ghc.client.Do(req)
	if err != nil {
		fmt.Println("failed to execute request:", err)
		return
	}

	defer resp.Body.Close()

	// Decode json response body into a map.
	// Sample response: https://pastebin.com/5WGaG0yz
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		fmt.Println("failed to decode response:", err)
		return
	}

	for key, value := range data {
		if len(value.(string)) == 0 {
			continue
		}

		// Determine index of the PR that this key corresponds with.
		idx, err := strconv.Atoi(strings.Split(key, "-")[1])
		if err != nil {
			fmt.Println("failed to determine index from key:", key)
			continue
		}

		// Access the PullRequest instance that we're updating review decision of.
		pr := (*prs)[idx]

		// Parse the HTML string from value we got in return.
		// Sample of this node: https://pastebin.com/55gQ1VbU
		node, err := html.Parse(strings.NewReader(value.(string)))
		if err != nil {
			fmt.Println("failed to parse item:", key)
			continue
		}

		// Extract review decision text from HTML.
		span := node.FirstChild.FirstChild.NextSibling.FirstChild // html->head->body->span
		a := span.FirstChild.NextSibling                          // span->text->a

		// Update review decision in original PullRequest object.
		pr.ReviewDecision = strings.TrimSpace(a.FirstChild.Data)
	}
}
