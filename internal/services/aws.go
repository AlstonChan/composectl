/*
Copyright © 2025 Chan Alston git@chanalston.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/AlstonChan/composectl/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	awsS3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

// Load AWS credentials and region from environment or ~/.aws/config
func GetAwsAccount(ctx context.Context) (*awsS3.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	client := awsS3.NewFromConfig(cfg)
	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("not logged in to AWS: %v", err)
	}

	fmt.Printf("Using AWS credential %s\n", *identity.Account)

	return client, nil
}

// Get the s3 bucket to that stores the backups
func GetS3BackupStoreBucket(ctx context.Context, s3Client *awsS3.Client,
	defaultBucketEnv string) (string, error) {
	var s3Bucket string = ""
	var err error = nil
	if s3Bucket == "" {
		CreateLocalCacheDir(os.Getenv(config.ConfigDirEnv))
		if val := viper.GetString(defaultBucketEnv); val != "" {
			s3Bucket = val
		}
	}

	if s3Bucket == "" {
		s3Bucket, err = PromptSelectS3Bucket(s3Client, ctx)
		if err != nil {
			return "", err
		}
	}

	return s3Bucket, nil
}

// Get all the directory from the bucket, check if the directory
// name contains the service name.
func ValidateS3BucketExists(ctx context.Context, s3Client *awsS3.Client, targetBucket string,
	serviceName string) (bool, error) {
	result, err := s3Client.ListObjectsV2(ctx, &awsS3.ListObjectsV2Input{
		Bucket:    &targetBucket,
		Delimiter: aws.String("/"), // treat "/" as directory separator
	})
	if err != nil {
		return false, fmt.Errorf("unable to retrieve service from S3: %v", err)
	}

	var serviceDirectoryFound bool = false
	for _, prefix := range result.CommonPrefixes {
		if strings.TrimSuffix(*prefix.Prefix, "/") == serviceName {
			serviceDirectoryFound = true
			fmt.Println("Service directory found:", serviceName)
		}
	}

	if !serviceDirectoryFound {
		return false, nil
	}

	return true, nil
}

func GetFileFromBucket(ctx context.Context, s3Client *awsS3.Client, targetBucket string,
	serviceName string, dateToRestoreAfter time.Time) (string, error) {
	// Get the file by filename
	backups, err := s3Client.ListObjectsV2(ctx, &awsS3.ListObjectsV2Input{
		Bucket: &targetBucket,
		Prefix: aws.String(serviceName + "/"),
	})
	if err != nil {
		return "", fmt.Errorf("unable to retrieve backup files from S3: %v", err)

	}

	if len(backups.Contents) == 0 {
		return "", fmt.Errorf("no file for service %s retrieved, recheck backup setting or consider manual backup: %v",
			serviceName, err)
	}

	var closestKey *awsS3Types.Object
	var closestDate time.Time

	// Truncate the target date to ignore time component (start of day)
	targetDate := dateToRestoreAfter.Truncate(24 * time.Hour)
	for _, content := range backups.Contents {
		t, err := ParseBackupTime(*content.Key, serviceName)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// Truncate backup date to ignore time component (start of day)
		backupDate := t.Truncate(24 * time.Hour)

		// Only consider backups from the target date or after
		if backupDate.Before(targetDate) {
			continue
		}

		// If this is our first valid backup, or if it's closer to the target date
		if closestKey == nil || backupDate.Before(closestDate) {
			closestKey = &content
			closestDate = backupDate
		}
	}

	// Check if we found any valid backup
	if closestKey == nil {
		return "", fmt.Errorf("no backup found for service %s on or after %s", serviceName,
			targetDate.Format("2006-01-02"))
	}

	fmt.Printf("Found closest backup: %s (date: %s)\n", *closestKey.Key, closestDate.Format("2006-01-02"))

	return *closestKey.Key, nil
}

// Download the object directly into memory
func S3DownloadToMemory(ctx context.Context, s3Client *awsS3.Client, targetBucket string,
	backupFilename string) ([]byte, error) {
	fmt.Printf("Downloading backup file: %s\n", backupFilename)
	downloadedBackup, err := s3Client.GetObject(ctx, &awsS3.GetObjectInput{
		Bucket: &targetBucket,
		Key:    &backupFilename,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to download backup file from S3: %v", err)
	}
	defer downloadedBackup.Body.Close()

	// Read the entire file into memory
	fileData, err := io.ReadAll(downloadedBackup.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading backup file data: %v", err)
	}

	fmt.Printf("Successfully downloaded %d bytes\n", len(fileData))

	return fileData, nil
}

func PromptSelectS3Bucket(s3Client *awsS3.Client, ctx context.Context) (string, error) {
	bucketList, err := s3Client.ListBuckets(ctx, &awsS3.ListBucketsInput{})
	if err != nil {
		return "", fmt.Errorf("failed to list buckets, %v", err)
	}

	bucketSlice := make([]string, len(bucketList.Buckets))
	for i, bucket := range bucketList.Buckets {
		bucketSlice[i] = *bucket.Name
	}

	prompt := promptui.Select{
		Label: "Select S3 Bucket",
		Items: bucketSlice,
	}

	_, s3Bucket, err := prompt.Run()

	if err != nil {
		return "", fmt.Errorf("prompt cancelled %v", err)
	}

	fmt.Printf("S3 bucket chosen: %q\n", s3Bucket)
	return s3Bucket, nil
}
