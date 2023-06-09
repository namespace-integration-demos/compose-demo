package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type LocalstackConfig struct {
	Region   string
	Endpoint string
}

func createLocalStackConfig(ctx context.Context, c LocalstackConfig) (aws.Config, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           c.Endpoint,
			SigningRegion: region,
		}, nil
	})

	var opts []func(*config.LoadOptions) error
	// Specify a custom resolver to be able to point to localstack's endpoint.
	opts = append(opts, config.WithEndpointResolverWithOptions(customResolver))

	if c.Region != "" {
		opts = append(opts, config.WithRegion(c.Region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config with error: %w, for endpoint %s", err, c.Endpoint)
	}

	return cfg, nil
}

func CreateLocalstackS3Client(ctx context.Context, config LocalstackConfig) (*s3.Client, error) {
	cfg, err := createLocalStackConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	s3client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return s3client, nil
}
