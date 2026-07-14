package dynamo

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// NewClient builds a DynamoDB client. When DYNAMODB_ENDPOINT is set (local
// development and tests against DynamoDB Local) it targets that endpoint with
// static dummy credentials, so no real AWS credentials are required. Otherwise
// it uses the default AWS config chain (production / Lambda).
func NewClient(ctx context.Context) (*dynamodb.Client, error) {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, err
		}
		return dynamodb.NewFromConfig(cfg), nil
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithBaseEndpoint(endpoint),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		),
	)
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(cfg), nil
}
