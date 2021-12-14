package gcpauth

import (
	"context"

	secretManager "cloud.google.com/go/secretmanager/apiv1"
	secretManagerProto "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// GetSecretByName gets a secret from gcp by its name
// The expected format is [projects/*/secrets/*/versions/*]
// for versions, you can use "latest" to grab the latest version
func GetSecretByName(name string) (string, error) {
	ctx := context.Background()
	client, err := secretManager.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	// Build the request.
	req := &secretManagerProto.AccessSecretVersionRequest{
		Name: name,
	}

	// Call the API.
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", err
	}

	return string(result.Payload.Data), nil
}
