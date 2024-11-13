# GitHub Credential Service

## Purpose

This is a microservice dedicated to abstracting dangerous GitHub permissions from Jenkins. GitHub credentials for actions such as `approve` and `status` will be hosted in this service. Then when Jenkins wishes to make use of the credentials, it will make a request to this service's API. After identity validation (JWT and IP allow list), the request will be processed with the GitHub API.

This removes the credentials from Jenkins, making them hidden so they can never be retrieved for free-range exploitation.

## Build

The microservice is built and deployed via the provided `Jenkinsfile`. It will compile the Golang binary, and upload both the binary and the OpenAPI Spec to S3. Then the stack will be deployed through CloudFormation using the AWS credentials specified in the Jenkinsfile.

After deployment, the Secret must be updated in AWS to have the value of the token. The pipeline only deploys a placeholder value.

### Prerequisites

While the entire service is built and deployed, a S3 bucket named `git-credential-service-artifacts` must first be deployed **manually** in the AWS account being used. Otherwise the artifacts will fail to upload.

## API Spec

> Version: v1

Each API path hosted by this service contains its own spec of what a valid request can contain. See all the current API specs for API version `v1` below:

### Post Status Service

```json
{
    "commit": {
        "org": "my-org",
        "repo": "my-repo",
        "sha": "1234567890abcdef"
    },
    "status": {
        "state": "success",
        "target_url": "https://example.com/build/status",
        "description": "The build succeeded!",
        "context": "continuous-integration/jenkins"
    }
}
```

## Testing

The service can be tested by manually invoking the API Gateway. An example `curl` command is displayed below. The URL must be replaced with the URL of the deployed API Gateway.

```bash
curl -X POST \
    -H "Content-Type: application/json" \
    -H "Api-Version: v1" \
    -d '{"commit":{"org":"KyleLavorato","repo":"git-credential-service","sha":"ac137c1d5a41f52bf46861e7d6ad03834e4ab576"},"status":{"state":"success","target_url":"https://google.com","description":"The status succeeded!","context":"example-status/celebrate"}}' \
    https://l01lbm90yh.execute-api.us-east-1.amazonaws.com/mdelpd-GithubCredentialAPIDeployment/api/github/status
```

## Future Work
* Add an authorizer to the API Gateway
* Add an IP whitelist to the to the API Gateway
* Add lambdas and APIs for other credential actions
* Add logging for security tracking purposes
