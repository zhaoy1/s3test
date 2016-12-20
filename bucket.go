package main

import "github.com/aws/aws-sdk-go/service/s3"
import "github.com/aws/aws-sdk-go/aws/awserr"
import "github.com/aws/aws-sdk-go/aws"

func BucketExist(svc *s3.S3, bktname *string) (bool, error) {

	req, _ := svc.GetBucketLocationRequest(&s3.GetBucketLocationInput{
		Bucket: bktname,
	})

	err := req.Send()
	if err != nil {
		if e, ok := err.(awserr.RequestFailure); ok && e.StatusCode() == 404 {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func CreateBucket(svc *s3.S3, bktname *string) error {
	param := &s3.CreateBucketInput{
		Bucket: aws.String(*bktname),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String("us-central-1"),
		},
	}

	_, err := svc.CreateBucket(param)
	if e, ok := err.(awserr.RequestFailure); ok && e.StatusCode() == 409 {
		return nil
	}

	return err
}
