package s3

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
)

func UploadFileToS3(filePath string, resource model.FileResource) (err error) {
	dat, errFile := os.Open(filePath)
	if errFile != nil {
		return errFile
	}

	sess, errSession := session.NewSession(&aws.Config{
		Region: aws.String(resource.Region),
		Credentials: credentials.NewStaticCredentials(
			resource.UploadParams.Credentials.AccessKeyId,
			resource.UploadParams.Credentials.SecretAccessKey,
			resource.UploadParams.Credentials.SessionToken),
	},
	)
	if errSession != nil {
		return errSession
	}

	// set the fixed timeout
	ctx := context.Background()
	var cancelFn func()
	ctx, cancelFn = context.WithTimeout(ctx, time.Millisecond*1000*30)
	if cancelFn != nil {
		defer cancelFn()
	}

	svc := s3.New(sess)

	_, err = svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(resource.UploadParams.Bucket),
		Key:    aws.String(resource.UploadParams.Key),
		Body:   dat,
	})

	return err
}
