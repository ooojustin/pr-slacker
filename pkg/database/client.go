package database

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/ooojustin/pr-puller/pkg/utils"
)

const region string = "us-east-1"

type Database struct {
	DynamoDB *dynamodb.DynamoDB
}

func Initialize() (*Database, bool) {
	cfg, ok := utils.GetConfig()
	if !ok {
		return nil, false
	}

	creds := credentials.NewStaticCredentials(cfg.AwsAccessKeyID, cfg.AwsAccessKeySecret, "")
	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
	})
	if err != nil {
		return nil, false
	}

	ddb := dynamodb.New(sess)
	db := &Database{
		DynamoDB: ddb,
	}

	return db, true
}
