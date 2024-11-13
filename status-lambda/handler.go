package statuslambda

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"KyleLavorato/git-credential-service/internal/logger"
	"KyleLavorato/git-credential-service/internal/utils"

	"github.com/aws/aws-lambda-go/events"
)

var (
	GIT_STATUS_CREDENTIAL_NAME string
)

func Init() {
	GIT_STATUS_CREDENTIAL_NAME = os.Getenv("GIT_STATUS_CREDENTIAL_NAME")
}

func HandleLambdaEvent(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger.SetRequestId(ctx, true)

	response := events.APIGatewayProxyResponse{
		StatusCode: http.StatusBadRequest,
		Headers: map[string]string{
			"Strict-Transport-Security": "max-age=63072000; includeSubDomains; preload", // HSTS header
		},
	}
	response.Body = fmt.Sprintf("{\"TraceId\": \"%s\"}", request.RequestContext.RequestID)

	// Validate headers
	err := validateHeaders(request.Headers)
	if err != nil {
		logger.Log.Errorf("Invalid headers: %s", err)
		return response, nil
	}
	response.Headers["Api-Version"] = request.Headers["Api-Version"] // Already validated

	// Validate body
	statusRequest, err := validateBody(request.Body)
	if err != nil {
		logger.Log.Errorf("Invalid request body: %s", err)
		return response, nil
	}

	// Get secret
	gitToken, err := utils.GetSecretValue(ctx, GIT_STATUS_CREDENTIAL_NAME)
	if err != nil {
		e := fmt.Errorf("Failed to get git token: %s", err.Error())
		logger.Log.Error(e)
		return response, e
	}
	logger.Log.Infof("Got token: %s", gitToken) // TODO: Remove this line

	// Trigger Git API
	err = statusRequest.PostCommitStatus(gitToken)
	if err != nil {
		return response, err
	}

	logger.Log.Infof("Successfully triggered Git API")
	response.StatusCode = http.StatusAccepted
	return response, nil
}
