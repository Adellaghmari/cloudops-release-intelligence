package dynamo

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type ClientOptions struct {
	Region   string
	Endpoint string
	Table    string
}

// usesLocalStaticCredentials is true only when an explicit DynamoDB Local
// endpoint is configured. Production AWS / Lambda must leave Endpoint empty so
// the SDK default credential chain (including the Lambda execution role) is used.
func usesLocalStaticCredentials(endpoint string) bool {
	return strings.TrimSpace(endpoint) != ""
}

func loadAWSConfig(ctx context.Context, opt ClientOptions) (aws.Config, error) {
	if opt.Region == "" {
		opt.Region = "eu-west-1"
	}
	loadOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(opt.Region),
	}
	// DynamoDB Local rejects SigV4 against real IAM credentials. Inject dummy
	// local/local credentials only for that explicit emulator path.
	if usesLocalStaticCredentials(opt.Endpoint) {
		loadOpts = append(loadOpts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("local", "local", ""),
		))
	}
	return awsconfig.LoadDefaultConfig(ctx, loadOpts...)
}

func NewClient(ctx context.Context, opt ClientOptions) (*dynamodb.Client, error) {
	cfg, err := loadAWSConfig(ctx, opt)
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		if opt.Endpoint != "" {
			o.BaseEndpoint = aws.String(opt.Endpoint)
		}
	}), nil
}
