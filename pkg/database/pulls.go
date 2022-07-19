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

func (db *Database) GetPullRequestExists(pr_uid string) (bool, error) {
	key := map[string]interface{}{pullRequestPK: pr_uid}

	av, err := dynamodbattribute.MarshalMap(key)
	if err != nil {
		return false, err
	}

	input := &dynamodb.GetItemInput{
		Key:                  av,
		TableName:            aws.String(pullRequestsTable),
		ProjectionExpression: aws.String(pullRequestPK),
	}

	output, err := db.DynamoDB.GetItem(input)
	if err != nil {
		return false, err
	}

	return len(output.Item) > 0, nil
}
