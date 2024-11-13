package utils

import (
	"context"
	"fmt"

	"KyleLavorato/git-credential-service/internal/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func GetSecretValue(ctx context.Context, secretName string) (string, error) {
	logger.Log.Debugf("Retrieving token from secret: %s", secretName)
	awsConfig, _ := config.LoadDefaultConfig(ctx)
	secretsClient := secretsmanager.NewFromConfig(awsConfig)
	secretOut, err := secretsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		e := fmt.Errorf("Failed to get secret: %s", err.Error())
		logger.Log.Error(e)
		return "", e
	}
	logger.Log.Debugf("Retrieved token")
	return *secretOut.SecretString, nil
}
