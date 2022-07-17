package github

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ooojustin/pr-puller/pkg/utils"
	"golang.org/x/net/html"
)

type PullRequest struct {
	Repository string
	Title      string
	Href       string
	Labels     []string
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
		pr := generatePullRequestObject(prNode)
		*prs = append(*prs, pr)
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

func generatePullRequestObject(prNode *html.Node) *PullRequest {
	aTagRepo := prNode.FirstChild.NextSibling.FirstChild.NextSibling.NextSibling.
		NextSibling.NextSibling.NextSibling.FirstChild.NextSibling
	repoName := strings.TrimSpace(aTagRepo.FirstChild.Data)
	aTagPR := aTagRepo.NextSibling.NextSibling
	prName := strings.TrimSpace(aTagPR.FirstChild.Data)
	href, _ := utils.GetAttribute(aTagPR, "href")
	labels := getPullRequestLabels(aTagPR)
	return &PullRequest{
		Repository: repoName,
		Title:      prName,
		Href:       href,
		Labels:     labels,
	}
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
