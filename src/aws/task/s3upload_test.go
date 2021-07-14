package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"math/rand"
	"os"
	"testing"
	"time"
)

func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

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
	//bucket := "landing-" + strings.ToLower(RandomString(10))
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
				Endpoint:         aws.String("http://localstack_main:4566"),
				S3ForcePathStyle: aws.Bool(true),
			},
		})

	// Create S3 service client
	svc := s3.New(sess)
	if sess != nil {
		fmt.Println("created")
	}
	//result, err := svc.ListBuckets(nil)
	//if err != nil {
	//	exitErrorf(t, "Unable to list buckets, %v", err)
	//}

	fmt.Println("Buckets:")
	//bktCreated := false
	//for _, b := range result.Buckets {
	//	fmt.Printf("* %s created on %s\n",
	//		aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	//	if strings.EqualFold(*b.Name, bucket) {
	//		continue
	//	} else {
	//		bktCreated = true
	//	}
	//}
	//if !bktCreated {
	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		exitErrorf(t, "Unable to create bucket %q, %v", bucket, err)
	}
	//}

	// Wait until bucket is created before finishing
	fmt.Printf("Waiting for bucket %q to be created...\n", bucket)

	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	fl := "./testdata/FULL_LOAD_BV_CX_CUST_BU_LIFECYCLE_DETAILS_manifesto.csv.gz"
	file, err := os.Open(fl)
	if err != nil {
		exitErrorf(t, "Unable to open file %q, %v", fl, err)
		t.Fail()
	}

	defer file.Close()

	listObjects, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket)})
	//resp, err = listObjects, err
	if err != nil {
		exitErrorf(t, "Unable to list items in bucket %q, %v", bucket, err)
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
		exitErrorf(t, "Unable to upload %q to %q, %v", filename, bucket, err)
	}

	fmt.Printf("Successfully uploaded %q to %q\n", filename, bucket)

	listObjects, err = svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket)})
	//resp, err = listObjects, err
	if err != nil {
		exitErrorf(t, "Unable to list items in bucket %q, %v", bucket, err)
	}

	for _, item := range listObjects.Contents {
		fmt.Println("Name:         ", *item.Key)
		fmt.Println("Last modified:", *item.LastModified)
		fmt.Println("Size:         ", *item.Size)
		fmt.Println("Storage class:", *item.StorageClass)
		fmt.Println("")
	}

	if err := DeleteBucket(t, bucket, svc); err != nil {
		t.Fatalf("unable to delete bucket %v", err)
	}
}

func DeleteBucket(t *testing.T, bucket string, svc *s3.S3) error {
	iter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	})

	if err := s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), iter); err != nil {
		exitErrorf(t, "Unable to delete objects from bucket %q, %v", bucket, err)
	}

	_, err := svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		exitErrorf(t, "Unable to delete bucket %q, %v", bucket, err)
	}

	// Wait until bucket is deleted before finishing
	fmt.Printf("Waiting for bucket %q to be deleted...\n", bucket)

	err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	return err
}

//func DeleteObject(bucket, fileName string) error {
//	if _, err := s.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
//		Bucket: aws.String(bucket),
//		Key:    aws.String(fileName),
//	}); err != nil {
//		return fmt.Errorf("delete: %w", err)
//	}
//
//	if err := s.client.WaitUntilObjectNotExists(&s3.HeadObjectInput{
//		Bucket: aws.String(bucket),
//		Key:    aws.String(fileName),
//	}); err != nil {
//		return fmt.Errorf("wait: %w", err)
//	}
//
//	return nil
//}

func exitErrorf(t *testing.T, msg string, args ...interface{}) {
	t.Fatalf(msg+"\n", args...)
}
