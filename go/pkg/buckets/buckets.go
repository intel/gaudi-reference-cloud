// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package buckets

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/logging"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/units"
)

// Configuration for 'buckets' package.
type Config struct {
	URL             string `koanf:"url"`
	SigningRegion   string `koanf:"region"`
	BucketName      string `koanf:"bucket"`
	DebugMode       bool   `koanf:"debug"`
	CredentialsFile string `koanf:"credentialsFile"`
}

type Bucket struct {
	endpointURL *url.URL
	s3Client    *s3.Client
	bucketName  string
	DebugMode   bool
}

const (
	MinPartSizeBytes = 5 * units.MiB
	MinPartNum       = 1
	MaxPartNum       = 10000
)

func New(ctx context.Context, config *Config) (*Bucket, error) {
	log := log.FromContext(ctx).WithName("Bucket.New")

	var (
		err         error
		endpointURL *url.URL
	)

	if config.URL != "" {
		endpointURL, err = url.Parse(config.URL)
	} else {
		//try local server if no endpoint specified
		endpointURL, err = url.Parse("http://127.0.0.1:9000")
	}
	if err != nil {
		return nil, err
	}

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {

		if service == s3.ServiceID && config != nil {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           endpointURL.String(),
				SigningRegion: config.SigningRegion,
			}, nil
		}

		log.Info("endpoint falling back to default")
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	chosenLogMode := aws.LogRetries

	if config.DebugMode {
		chosenLogMode = aws.LogRetries | aws.LogRequest | aws.LogResponse | aws.LogDeprecatedUsage | aws.LogRequestEventMessage | aws.LogResponseEventMessage
	}

	logger := logging.LoggerFunc(func(classification logging.Classification, format string, v ...interface{}) {
		msg := fmt.Sprintf(format, v...)
		log.WithName("AWS").Info(msg)
	})

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithEndpointResolverWithOptions(customResolver),
		awsconfig.WithClientLogMode(chosenLogMode), awsconfig.WithLogger(logger),
		awsconfig.WithSharedCredentialsFiles([]string{config.CredentialsFile}),
	)
	if err != nil {
		return nil, fmt.Errorf("can't load default config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	log.Info("Bucket configuration", "endpointURL", endpointURL)

	return &Bucket{
		endpointURL: endpointURL,
		s3Client:    client,
		bucketName:  config.BucketName,
		DebugMode:   config.DebugMode,
	}, nil

}

// Upload uses an upload manager to upload an object to a bucket.
// The upload manager uses io.Reader to read out object and puts the data in parts.
// Ensure toUpload Reader satisfies ReadSeekerAt interface to improve
// application resource utilization of the host environment.
func (b *Bucket) Upload(ctx context.Context, toUpload io.Reader, objectKey string, partMiBs int64) error {
	log := log.FromContext(ctx).WithName("Bucket.Upload")
	if !IsReaderAtSeeker(toUpload) && b.DebugMode {
		log.Info("Use ReaderAtSeeker interface for improved resource utilization")
	}

	log.Info("Uploading s3://" + b.bucketName + "/" + objectKey + " via " + b.endpointURL.Host)

	uploader := manager.NewUploader(b.s3Client, func(u *manager.Uploader) {
		u.PartSize = partMiBs * 1024 * 1024

	})

	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(b.bucketName),
		Key:    aws.String(objectKey),
		Body:   toUpload,
	})
	if err != nil {
		return fmt.Errorf("couldn't upload object: %w", err)
	}

	return err
}

func (b *Bucket) ListKeysByPrefix(ctx context.Context, prefix string, maxKeys int32) (keys []string, truncated bool, err error) {
	// Get the first page of results for ListObjectsV2 for a bucket
	output, err := b.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(b.bucketName),
		Prefix:  aws.String(prefix),
		MaxKeys: maxKeys,
	})
	if err != nil {
		return keys, false, fmt.Errorf("can't list objects from bucket: %w", err)
	}

	keys = make([]string, 0, output.KeyCount)
	for _, object := range output.Contents {
		keys = append(keys, aws.ToString(object.Key))

	}
	return keys, output.IsTruncated, err
}

// Download uses a download manager to download an object from a bucket.
// The download manager gets the data in parts and writes them to a buffer until all of
// the data has been downloaded.
func (b *Bucket) Download(ctx context.Context, objectToDownload, targetDirectory string, partMiBs int64, force bool) error {

	log := log.FromContext(ctx).WithName("Bucket.Download")

	downloader := manager.NewDownloader(b.s3Client, func(d *manager.Downloader) {
		d.PartSize = partMiBs * 1024 * 1024
	})

	// Create the directories in the path
	file := filepath.Join(targetDirectory, objectToDownload)
	if err := os.MkdirAll(filepath.Dir(file), 0775); err != nil {
		return err
	}

	var flags int
	if !force {
		flags = os.O_CREATE | os.O_EXCL | os.O_RDWR
	} else {
		flags = os.O_CREATE | os.O_TRUNC | os.O_RDWR

	}

	fd, err := os.OpenFile(file, flags, 0600)
	if err != nil {
		return err
	}
	defer fd.Close()

	log.Info("Downloading s3://" + b.bucketName + "/" + objectToDownload + " to:" + file)
	_, err = downloader.Download(ctx, fd, &s3.GetObjectInput{Bucket: &b.bucketName, Key: &objectToDownload})

	return err
}

func (b *Bucket) CreateMultipartUpload(ctx context.Context, objectKey string, metadata map[string]string) (*s3.CreateMultipartUploadOutput, error) {
	logger := log.FromContext(ctx).WithName("Bucket.Create")
	if objectKey == "" {
		return nil, fmt.Errorf("object key cannot be an empty string")
	}
	input := &s3.CreateMultipartUploadInput{
		Bucket:   aws.String(b.bucketName),
		Key:      aws.String(objectKey),
		Metadata: metadata,
	}
	logger.Info("Creating multipart upload",
		"bucket", b.bucketName,
		"object key", objectKey,
	)
	resp, err := b.s3Client.CreateMultipartUpload(ctx, input)
	logger.Info("Response from API", "response", resp, "error", err)
	return resp, err
}

func (b *Bucket) CompleteMultipartUpload(ctx context.Context, objectKey string, uploadId string, uploadedParts []types.CompletedPart) (*s3.CompleteMultipartUploadOutput, error) {
	logger := log.FromContext(ctx).WithName("Bucket.CompleteMultipartUpload")
	if objectKey == "" {
		return nil, fmt.Errorf("object key cannot be an empty string")
	}
	if uploadId == "" {
		return nil, fmt.Errorf("uploadId cannot be an empty string")
	}
	if len(uploadedParts) == 0 {
		return nil, fmt.Errorf("uploadedParts list cannot be empty")
	}
	for _, part := range uploadedParts {
		if *part.ETag == "" {
			return nil, fmt.Errorf("ETag cannot be an empty string")
		}
		if err := IsValidPartNumber(int32(part.PartNumber)); err != nil {
			return nil, err
		}
	}
	input := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(b.bucketName),
		Key:      aws.String(objectKey),
		UploadId: aws.String(uploadId),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: uploadedParts,
		},
	}
	logger.Info("Completing multipart upload",
		"bucket", b.bucketName,
		"object key", objectKey,
		"uploadId", uploadId,
		"parts", uploadedParts,
	)
	resp, err := b.s3Client.CompleteMultipartUpload(ctx, input)
	logger.Info("Response from API", "response", resp, "error", err)
	return resp, err
}

func (b *Bucket) AbortMultipartUpload(ctx context.Context, objectKey string, uploadId string) error {
	logger := log.FromContext(ctx).WithName("Bucket.AbortMultipartUpload")
	if objectKey == "" {
		return fmt.Errorf("object key cannot be an empty string")
	}
	if uploadId == "" {
		return fmt.Errorf("uploadId cannot be an empty string")
	}
	input := &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(b.bucketName),
		Key:      aws.String(objectKey),
		UploadId: aws.String(uploadId),
	}
	resp, err := b.s3Client.AbortMultipartUpload(ctx, input)
	logger.Info("Response from API", "response", resp, "error", err)
	if err != nil {
		logger.Error(err, "failed to abort multipart upload", "objectKey", objectKey)
		return err
	}
	return nil
}

func (b *Bucket) GetObjectMetadata(ctx context.Context, objectKey string) (map[string]string, error) {
	logger := log.FromContext(ctx).WithName("Bucket.GetObjectMetadata")
	input := &s3.HeadObjectInput{
		Bucket: &b.bucketName,
		Key:    aws.String(objectKey),
	}
	resp, err := b.s3Client.HeadObject(ctx, input)
	if err != nil {
		logger.Error(err, "failed to get object metadata", "objectKey", objectKey)
		return nil, err
	}
	return resp.Metadata, nil
}

func (b *Bucket) PresignGetObject(ctx context.Context, objectKey string, expirationDuration time.Duration) (*url.URL, error) {
	logger := log.FromContext(ctx).WithName("Bucket.PresignGetObject")
	presignClient := s3.NewPresignClient(b.s3Client)
	input := s3.GetObjectInput{
		Bucket: &b.bucketName,
		Key:    aws.String(objectKey),
	}
	resp, err := presignClient.PresignGetObject(context.Background(), &input,
		func(opts *s3.PresignOptions) {
			opts.Expires = expirationDuration
		},
	)
	if err != nil {
		logger.Error(err, "failed to generate presigned URL for artifact download", "objectKey", objectKey)
		return nil, err
	}
	return url.ParseRequestURI(resp.URL)
}

func (b *Bucket) PresignPutObject(ctx context.Context, objectKey string, expirationDuration time.Duration, metadata map[string]string) (*url.URL, error) {
	logger := log.FromContext(ctx).WithName("Bucket.PresignPutObject")
	if objectKey == "" {
		return nil, fmt.Errorf("object key cannot be an empty string")
	}
	presignClient := s3.NewPresignClient(b.s3Client)
	input := s3.PutObjectInput{
		Bucket:   &b.bucketName,
		Key:      aws.String(objectKey),
		Metadata: metadata,
	}
	resp, err := presignClient.PresignPutObject(context.Background(), &input,
		func(opts *s3.PresignOptions) {
			opts.Expires = expirationDuration
		},
	)
	if err != nil {
		logger.Error(err, "failed to generate presigned URL for artifact upload", "objectKey", objectKey)
		return nil, err
	}
	return url.ParseRequestURI(resp.URL)
}

func (b *Bucket) PresignUploadPart(ctx context.Context, objectKey string, expirationDuration time.Duration, uploadId string, partNum uint32) (*url.URL, error) {
	logger := log.FromContext(ctx).WithName("Bucket.PresignUploadPart")
	if objectKey == "" {
		return nil, fmt.Errorf("object key cannot be an empty string")
	}
	if uploadId == "" {
		return nil, fmt.Errorf("uploadId cannot be an empty string")
	}
	if err := IsValidPartNumber(int32(partNum)); err != nil {
		return nil, err
	}
	presignClient := s3.NewPresignClient(b.s3Client)
	input := s3.UploadPartInput{
		Bucket:     &b.bucketName,
		Key:        aws.String(objectKey),
		UploadId:   aws.String(uploadId),
		PartNumber: int32(partNum),
	}
	resp, err := presignClient.PresignUploadPart(context.Background(), &input,
		func(opts *s3.PresignOptions) {
			opts.Expires = expirationDuration
		},
	)
	if err != nil {
		logger.Error(err, "failed to generate presigned URL for artifact part upload", "objectKey", objectKey)
		return nil, err
	}
	return url.ParseRequestURI(resp.URL)
}

func IsReaderAtSeeker(interfaceToTest io.Reader) bool {
	switch interfaceToTest.(type) {
	case readerAtSeeker:
		return true

	default:
		// Reader does not satisfy ReadSeekerAt interface required by Uploader for improved application resource utilization
		return false
	}
}

type readerAtSeeker interface {
	io.ReaderAt
	io.ReadSeeker
}
