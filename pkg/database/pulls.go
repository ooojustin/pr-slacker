package database

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	pr_gh "github.com/ooojustin/pr-puller/pkg/github"
)

func (db *Database) PutPullRequest(pr *pr_gh.PullRequest) bool {
	av, err := dynamodbattribute.MarshalMap(pr)
	if err != nil {
		return false
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(prTable),
	}

	_, err = db.DynamoDB.PutItem(input)
	if err != nil {
		return false
	}

	return true
}
