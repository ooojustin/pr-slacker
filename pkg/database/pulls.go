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

type PutPullRequestCounts struct {
	Uploaded int
	Updated  int
	Skipped  int
	Failed   int
}

// Returns: uploaded, skipped, failed
func (db *Database) PutPullRequests(prs []*pr_gh.PullRequest) PutPullRequestCounts {
	var counts PutPullRequestCounts
	for _, pr := range prs {
		existingPR, err := db.GetPullRequest(pr.PK)
		if err != nil && err != ItemNotFoundError {
			counts.Skipped++
			continue
		}

		var update bool
		if existingPR != nil {
			update = (existingPR.Draft && !pr.Draft) ||
				(existingPR.ReviewDecision != pr.ReviewDecision)
			if !update {
				counts.Skipped++
				continue
			}
		}

		if db.PutPullRequest(pr) {
			if update {
				counts.Updated++
			} else {
				counts.Uploaded++
			}
		} else {
			counts.Failed++
		}
	}
	return counts
}

func (db *Database) PutPullRequest(pr *pr_gh.PullRequest) bool {
	av, err := dynamodbattribute.MarshalMap(pr)
	if err != nil {
		fmt.Println("Failed to marshal PullRequest:", err)
		return false
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(pullRequestsTable),
	}

	_, err = db.DynamoDB.PutItem(input)
	if err != nil {
		fmt.Println("Failed to PutItem PullRequest:", err)
		return false
	}

	return true
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
