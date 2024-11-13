package main

import (
	"os"
	"strings"

	statuslambda "KyleLavorato/git-credential-service/status-lambda"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	name := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	if name == "" {
		panic("No lambda function detected")
	}
	if strings.Contains(name, "GithubPostStatus") {
		statuslambda.Init()
		lambda.Start(statuslambda.HandleLambdaEvent)
	}
}
