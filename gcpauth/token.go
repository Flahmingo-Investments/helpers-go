package gcpauth

import (
	"context"
	"os"
	"regexp"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"google.golang.org/api/option"
	credentialsPb "google.golang.org/genproto/googleapis/iam/credentials/v1"
)

const (
	EnvServiceAcctFile = "GOOGLE_APPLICATION_CREDENTIALS"
)

func isEnvExist(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

// GetAuthToken returns an access token from GCP
// If the GOOGLE_APPLICATION_CREDENTIALS environment variable is set, it will read an auth.json file from the path
// If it isn't set, it will use the use internal GCP mechanism to authenticate it's self.
func GetAuthToken(saEmail string) (string, error) {

	// getting a token needs this to be appended, so automatically add it if it's not there
	serviceAcctRegex, _ := regexp.Compile("\\.gserviceaccount\\.com$")

	if !serviceAcctRegex.Match([]byte(saEmail)) {
		saEmail = saEmail + ".gserviceaccount.com"
	}

	if isEnvExist(EnvServiceAcctFile) {
		return GetAuthFromFile(os.Getenv(EnvServiceAcctFile), saEmail)
	}
	return GetAuthFromKube(saEmail)
}

func GetAuthFromKube(saEmail string) (string, error) {
	ctx := context.Background()
	credentialsClient, err := credentials.NewIamCredentialsClient(ctx)
	if err != nil {
		return "", err
	}
	defer credentialsClient.Close()

	return getToken(ctx, saEmail, credentialsClient)
}

func GetAuthFromFile(path, saEmail string) (string, error) {
	ctx := context.Background()
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	credentialsClient, err :=
		credentials.NewIamCredentialsClient(ctx, option.WithCredentialsFile(path))
	if err != nil {
		return "", err
	}
	defer credentialsClient.Close()

	return getToken(ctx, saEmail, credentialsClient)
}

func getToken(ctx context.Context,
	saEmail string,
	c *credentials.IamCredentialsClient) (string, error) {
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
