package config

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// FetchSecret retrieves a secret string from AWS Secrets Manager in the given
// region by name. The caller is responsible for parsing the returned JSON.
func FetchSecret(region string, name string) (string, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return "", fmt.Errorf("create aws session: %w", err)
	}
	sm := secretsmanager.New(sess)
	out, err := sm.GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: aws.String(name)})
	if err != nil {
		return "", fmt.Errorf("get secret: %w", err)
	}
	if out.SecretString == nil {
		return "", fmt.Errorf("secret has no string value")
	}
	return *out.SecretString, nil
}
