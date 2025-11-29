package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fpgschiba/volleygoals/models"
)

func GeneratePresignedPutURL(ctx context.Context, key, contentType string, expires int) (string, error) {
	presignClient = GetPresignClient()
	response, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(options *s3.PresignOptions) {
		options.Expires = time.Duration(expires * 60)
	})
	if err != nil {
		return "", err
	}
	return response.URL, nil
}

func GeneratePresignedUploadURLForUserPicture(ctx context.Context, userID, filename, contentType string, expires int) (string, string, error) {
	presignClient = GetPresignClient()
	fileExtension := filepath.Ext(filename)
	newFilename := fmt.Sprintf("%s%s", models.GenerateID(), fileExtension)
	key := fmt.Sprintf("users/%s/%s", userID, newFilename)
	url, err := GeneratePresignedPutURL(ctx, key, contentType, expires)
	if err != nil {
		return "", key, err
	}
	return url, key, nil
}
