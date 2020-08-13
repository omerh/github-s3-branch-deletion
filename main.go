package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type githubHook struct {
	Ref        string `json:"ref"`
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
}

func work(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.Headers["X-GitHub-Event"] == "delete" {
		repoName, branchName := retriveInfoFromHook(req.Body)
		fmt.Printf("Brnach %v was deleted from repo %v\n", branchName, repoName)
		deleteFromS3Key(repoName, branchName)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string("Ok"),
	}, nil
}

func deleteFromS3Key(repoName string, branchName string) {
	bucket, _ := os.LookupEnv("bucket")
	svc := s3.New(session.New())
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(repoName + "/branches/" + branchName),
	}
	result, _ := svc.DeleteObject(input)
	fmt.Println(result)
}

func retriveInfoFromHook(hookBody string) (repoName string, branchName string) {
	var hook githubHook
	body, _ := url.ParseQuery(hookBody)
	for _, v := range body {
		for _, vv := range v {
			err := json.Unmarshal([]byte(vv), &hook)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return hook.Repository.Name, hook.Ref
}

func main() {
	lambda.Start(work)
}
