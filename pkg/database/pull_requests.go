package database

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	pr_gh "github.com/ooojustin/pr-puller/pkg/github"
)

const pullRequestsTable string = "pull-requests"
const pullRequestPK string = "pr_uid"

var (
	ItemNotFoundError error = errors.New("Item not found.")
)

type PutPullRequestsResponse struct {
	Uploaded []*pr_gh.PullRequest
	Updated  []*pr_gh.PullRequest
	Skipped  []*pr_gh.PullRequest
	Failed   []*pr_gh.PullRequest
	Notify   []*pr_gh.PullRequest
}

func (db *Database) PutPullRequests(prs []*pr_gh.PullRequest) PutPullRequestsResponse {
	var response PutPullRequestsResponse
	for _, pr := range prs {
		existingPR, err := db.GetPullRequest(pr.PK)
		if err != nil && err != ItemNotFoundError {
			response.Skipped = append(response.Skipped, pr)
			continue
		}

		var update bool
		if existingPR != nil {
			update = (existingPR.Draft && !pr.Draft) ||
				((existingPR.ReviewDecision != pr.ReviewDecision) && pr.ReviewDecision == "Review required")
			if !update {
				response.Skipped = append(response.Skipped, pr)
				continue
			}
		}

		var notify bool
		if !pr.Notified && !pr.Draft && pr.ReviewDecision != "Approved" {
			notify = true
			pr.Notified = true
		}

		if err := db.PutPullRequest(pr); err == nil {
			if update {
				response.Updated = append(response.Updated, pr)
			} else {
				response.Uploaded = append(response.Uploaded, pr)
			}

			if notify {
				response.Notify = append(response.Notify, pr)
			}
		} else {
			response.Failed = append(response.Failed, pr)
		}
	}
	return response
}

func (db *Database) PutPullRequest(pr *pr_gh.PullRequest) error {
	av, err := dynamodbattribute.MarshalMap(pr)
	if err != nil {
		fmt.Println("Failed to marshal PullRequest:", err)
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(pullRequestsTable),
	}

	_, err = db.DynamoDB.PutItem(input)
	if err != nil {
		fmt.Println("Failed to PutItem PullRequest:", err)
		return err
	}

	return nil
}

func (db *Database) GetPullRequest(pr_uid string) (*pr_gh.PullRequest, error) {
	key := map[string]interface{}{pullRequestPK: pr_uid}

	av, err := dynamodbattribute.MarshalMap(key)
	if err != nil {
		fmt.Println("Failed to marshal PullRequest key:", err)
		return nil, err
	}

	input := &dynamodb.GetItemInput{
		Key:       av,
		TableName: aws.String(pullRequestsTable),
	}

	var pr pr_gh.PullRequest

	output, err := db.DynamoDB.GetItem(input)
	if err != nil {
		fmt.Println("Failed to GetItem PullRequest:", err)
		return nil, err
	}

	if len(output.Item) == 0 {
		return nil, ItemNotFoundError
	}

	err = dynamodbattribute.UnmarshalMap(output.Item, &pr)
	if err != nil {
		fmt.Println("Failed to convert PullRequest output to object:", err)
		return nil, err
	}

	return &pr, nil
}
