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
	ID         int
	Created    time.Time
	Creator    string
	Repository string
	Title      string
	Href       string
	Labels     []string
	Draft      bool
	Status     string
}

func (pr PullRequest) ToString() string {
	return fmt.Sprintf(
		"%s %s %s %s %s %s %d",
		pr.Title,
		pr.Href,
		pr.Labels,
		pr.Created,
		pr.Creator,
		pr.Status,
		pr.ID,
	)
}

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

func (ghc *GithubClient) loadPullRequests(
	doc *goquery.Document,
	page int,
	org string,
	open bool,
	prs *[]*PullRequest,
) {
	if doc == nil {
		var ok bool
		doc, ok = ghc.loadPullRequestDocument(page, org, open)
		if !ok {
			return
		}
	}

	prNodes := getPullRequestNodes(doc)
	for _, prNode := range prNodes {
		if pr, ok := generatePullRequestObject(prNode); ok {
			*prs = append(*prs, pr)
		}
	}
}

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
	repoName := strings.TrimSpace(aTagRepo.FirstChild.Data)
	aTagPR := aTagRepo.NextSibling.NextSibling
	prName := strings.TrimSpace(aTagPR.FirstChild.Data)
	href, _ := utils.GetAttribute(aTagPR, "href")
	labels := getPullRequestLabels(aTagPR)
	opened := getPullRequestOpenedNode(aTagPR)
	datetimeNode := opened.FirstChild.NextSibling
	datetimeStr, _ := utils.GetAttribute(datetimeNode, "datetime")
	username := strings.TrimSpace(datetimeNode.NextSibling.NextSibling.FirstChild.Data)
	datetime, _ := time.Parse(time.RFC3339, datetimeStr)

	var id int
	if !draft {
		idInput := opened.NextSibling.NextSibling.FirstChild.NextSibling.FirstChild.NextSibling
		if prId, ok := utils.GetAttribute(idInput, "value"); ok {
			id, _ = strconv.Atoi(prId)
		}
	}

	pr := &PullRequest{
		ID:         id,
		Created:    datetime,
		Creator:    username,
		Repository: repoName,
		Title:      prName,
		Href:       href,
		Labels:     labels,
		Draft:      draft,
	}

	return pr, true
}

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

		}
	}
}
