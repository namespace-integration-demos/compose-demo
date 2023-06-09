package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/cenkalti/backoff/v4"
)

const connBackoff = time.Second

func EnsureBucketExistsByName(ctx context.Context, client *s3.Client, name string) error {
	log.Printf("%s: creating bucket...\n", name)
	if err := backoff.Retry(func() error {
		input := &s3.CreateBucketInput{
			Bucket: &name,
		}

		// Speed up bucket creation through faster retries.
		ctx, cancel := context.WithTimeout(ctx, connBackoff)
		defer cancel()

		if _, err := client.CreateBucket(ctx, input); err != nil {
			var alreadyExists *types.BucketAlreadyExists
			var alreadyOwned *types.BucketAlreadyOwnedByYou
			if errors.As(err, &alreadyExists) || errors.As(err, &alreadyOwned) {
				log.Printf("%s: bucket already exists.\n", name)
				return nil
			}

			err = fmt.Errorf("failed to create bucket: %w", err)
			log.Println(err)
			return err
		}

		log.Printf("%s: bucket created.\n", name)

		return nil
	}, backoff.WithContext(backoff.NewConstantBackOff(connBackoff), ctx)); err != nil {
		return fmt.Errorf("failed to create S3 bucket: %w", err)
	}

	return nil
}
