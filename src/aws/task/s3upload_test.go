package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
	"testing"
)

func TestWithLocalStack(t *testing.T) {
	//l, err := localstack.NewInstance()
	//if err != nil {
	//	t.Fatalf("Could not connect to Docker %v", err)
	//}
	//if err := l.Start(); err != nil {
	//	t.Fatalf("Could not start localstack %v", err)
	//}
	//
	//defer func() {
	//	if err := l.Stop(); err != nil {
	//		t.Fatalf("Could not stop localstack %v", err)
	//	}
	//}()

	bucket := "landing"
	//TODO check if we can use same session - revert dbSession
	//sess, err := session.NewSession(&aws.Config{
	//	Credentials:      credentials.NewStaticCredentials("not", "empty", ""),
	//	DisableSSL:       aws.Bool(true),
	//	Region:           aws.String(endpoints.UsWest2RegionID),
	//	Endpoint:         aws.String(l.Endpoint(localstack.S3)),
	//	S3ForcePathStyle: aws.Bool(true),
	//})

	sess, err := session.NewSessionWithOptions(
		session.Options{
			Config: aws.Config{
				Credentials:      credentials.NewStaticCredentials("test", "test", ""),
				Region:           aws.String("us-west-2"),
				Endpoint:         aws.String("http://localhost:4566"),
				S3ForcePathStyle: aws.Bool(true),
			},
			Profile: "localstack",
		})

	// Create S3 service client
	svc := s3.New(sess)
	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		exitErrorf("Unable to create bucket %q, %v", bucket, err)
	}

	// Wait until bucket is created before finishing
	fmt.Printf("Waiting for bucket %q to be created...\n", bucket)

	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	file, err := os.Open("./testdata/FULL_LOAD_BV_CX_CUST_BU_LIFECYCLE_DETAILS_manifesto.csv.gz")
	if err != nil {
		exitErrorf("Unable to open file %q, %v", err)
		t.Fail()
	}

	defer file.Close()

	listObjects, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket)})
	//resp, err = listObjects, err
	if err != nil {
		exitErrorf("Unable to list items in bucket %q, %v", bucket, err)
	}

	for _, item := range listObjects.Contents {
		fmt.Println("Name:         ", *item.Key)
		fmt.Println("Last modified:", *item.LastModified)
		fmt.Println("Size:         ", *item.Size)
		fmt.Println("Storage class:", *item.StorageClass)
		fmt.Println("")
	}

	uploader := s3manager.NewUploader(sess)
	filename := "DataFoundation/SERVICES/LIFECYCLE/LIFECYCLE_DETAILS/20210706/FULL_LOAD_BV_CX_CUST_BU_LIFECYCLE_DETAILS_manifesto.csv.gz"
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		// Print the error and exit.
		exitErrorf("Unable to upload %q to %q, %v", filename, bucket, err)
	}

	fmt.Printf("Successfully uploaded %q to %q\n", filename, bucket)

	listObjects, err = svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket)})
	//resp, err = listObjects, err
	if err != nil {
		exitErrorf("Unable to list items in bucket %q, %v", bucket, err)
	}

	for _, item := range listObjects.Contents {
		fmt.Println("Name:         ", *item.Key)
		fmt.Println("Last modified:", *item.LastModified)
		fmt.Println("Size:         ", *item.Size)
		fmt.Println("Storage class:", *item.StorageClass)
		fmt.Println("")
	}

}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	//os.Exit(1)
}
