package pkg

import (
	"context"
	"os"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"google.golang.org/api/option"
	credentialsPb "google.golang.org/genproto/googleapis/iam/credentials/v1"
)

const (
	EnvServiceAcctFile = "GOOGLE_APPLICATION_CREDENTIALS"
)

func isEnvExist(key string) bool {
	if _, ok := os.LookupEnv(key); ok {
		return true
	}
	return false
}

func GetAuthToken(saEmail string) (string, error) {
	if isEnvExist(EnvServiceAcctFile) {
		return GetAuthFromFile(os.Getenv(EnvServiceAcctFile), saEmail)
	}
	return GetAuthFromKube(saEmail)

}

func GetAuthFromKube(saEmail string) (string, error) {
	ctx := context.Background()
	c, err := credentials.NewIamCredentialsClient(ctx)
	if err != nil {
		return "", err
	}
	defer c.Close()

	requestOpts := &credentialsPb.GenerateAccessTokenRequest{
		Name:  saEmail,
		Scope: []string{"https://www.googleapis.com/auth/cloud-platform"},
	}

	token, err := c.GenerateAccessToken(ctx, requestOpts)

	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func GetAuthFromFile(path, saEmail string) (string, error) {
	ctx := context.Background()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", err
	} else if err != nil {
		return "", err
	}

	c, err := credentials.NewIamCredentialsClient(ctx, option.WithCredentialsFile(path))
	if err != nil {
		return "", err
	}

	defer c.Close()

	requestOpts := &credentialsPb.GenerateAccessTokenRequest{
		Name:  saEmail,
		Scope: []string{"https://www.googleapis.com/auth/cloud-platform"},
	}

	token, err := c.GenerateAccessToken(ctx, requestOpts)

	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}
