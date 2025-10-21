package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage interface {
	Save(localTmpPath, destName string) (string, error)
}

type LocalStorage struct {
	basePath string
}

func NewLocal(basePath string) Storage {
	_ = os.MkdirAll(basePath, 0o755)
	return &LocalStorage{basePath: basePath}
}

func (l *LocalStorage) Save(tmpPath, destName string) (string, error) {
	dst := filepath.Join(l.basePath, destName)

	srcF, err := os.Open(tmpPath)
	if err != nil {
		return "", err
	}
	defer srcF.Close()

	dstF, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer dstF.Close()

	if _, err := io.Copy(dstF, srcF); err != nil {
		return "", err
	}

	return dst, nil
}

// S3Storage implements Storage interface for AWS S3
type S3Storage struct {
	client *s3.Client
	bucket string
}

func NewS3(bucket string) (Storage, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	return &S3Storage{
		client: client,
		bucket: bucket,
	}, nil
}

func (s *S3Storage) Save(tmpPath, destName string) (string, error) {
	// Open the file
	file, err := os.Open(tmpPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Upload to S3
	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(destName),
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	// Return S3 key (path in bucket)
	return destName, nil
}
