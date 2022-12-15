package gcp

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/Flahmingo-Investments/helpers-go/ferrors"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var versionsRegex = regexp.MustCompile(`/versions/.*$`)

type SecretVersion struct {
	Path  string
	State secretmanagerpb.SecretVersion_State
}

// ListSecretVersions Get all the secret versions and their status
// Expected format is
// - projects/<project id or project name>/secrets/secret-name
// - projects/<project id or project name>/secrets/secret-name/versions/<int>
func ListSecretVersions(name string) ([]SecretVersion, error) {
	if versionsRegex.MatchString(name) {
		name = versionsRegex.ReplaceAllString(name, "")
	}

	var rtnMe []SecretVersion

	// Create the client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	// Build the request.
	req := &secretmanagerpb.ListSecretVersionsRequest{
		Parent: name,
	}

	// Call the API.
	it := client.ListSecretVersions(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list secret versions: %v", err)
		}
		rtnMe = append(rtnMe, SecretVersion{
			Path:  resp.Name,
			State: resp.State,
		})
	}

	return rtnMe, nil
}

// GetSecretByName gets a secret from gcp by its name
// The expected format is [projects/*/secrets/*/versions/*]
// for versions, you can use "latest" to grab the latest version
func GetSecretByName(name string) (string, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()
	// Build the request.

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	// Call the API.
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", err
	}

	return string(result.Payload.Data), nil
}

// SecretClient is a wrapper around GCP Secret Service.
// It provides useful helpers to for getting and creating secrets.
type SecretClient struct {
	client *secretmanager.Client
	ctx    context.Context
}

const minSecretPathLength = 4

// NewSecretClient creates a new gcp secret manager service.
// If no option is provided it will use the Google Cloud ADC to initialize the client.
//
// The returned client must be Closed when it is done being used to clean up
// its underlying connections by calling the `func Close()`
func NewSecretClient(opts ...option.ClientOption) (*SecretClient, error) {
	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx, opts...)
	if err != nil {
		return nil, ferrors.Wrap(err, "unable to create secret manager client")
	}

	return &SecretClient{
		ctx:    ctx,
		client: client,
	}, nil
}

// GetSecret fetches the secret from GCP Secret Manager and return it as
// a string.
//
// The expected formats are:
// - projects/<project>/secrets/<name>/versions/<version>
// - projects/<project>/secrets/<name>/versions/latest
func (c *SecretClient) GetSecret(name string) (string, error) {
	secretPath := strings.Split(name, "/")
	if len(secretPath) < minSecretPathLength {
		return "", ferrors.New("secret name is not in expected format")
	}

	if len(secretPath) == minSecretPathLength {
		secretPath = append(secretPath, "versions", "latest")
	}

	res, err := c.client.AccessSecretVersion(
		c.ctx,
		&secretmanagerpb.AccessSecretVersionRequest{
			Name: strings.Join(secretPath, "/"),
		},
	)
	if err != nil {
		return "", err
	}

	return string(res.Payload.Data), nil
}

// CreateSecret creates a secret in GCP Secret Service.
// The expected name format is:
// - projects/<project>/secrets/<name>
func (c *SecretClient) CreateSecret(name, value string) error {
	secretPath := strings.Split(name, "/")
	if len(secretPath) < minSecretPathLength {
		return ferrors.New("secret name is not in expected format")
	}

	// drop 'secrets' from the path
	parent := strings.Join(secretPath[:len(secretPath)-2], "/")
	secretName := secretPath[len(secretPath)-1]

	createSecretRes, err := c.client.CreateSecret(
		c.ctx,
		&secretmanagerpb.CreateSecretRequest{
			Parent:   parent,
			SecretId: secretName,
			Secret: &secretmanagerpb.Secret{
				Replication: &secretmanagerpb.Replication{
					Replication: &secretmanagerpb.Replication_Automatic_{
						Automatic: &secretmanagerpb.Replication_Automatic{},
					},
				},
			},
		},
	)
	if err != nil {
		return ferrors.Wrapf(err, "unable to create secret: %s", name)
	}

	_, err = c.client.AddSecretVersion(
		c.ctx,
		&secretmanagerpb.AddSecretVersionRequest{
			Parent: createSecretRes.GetName(),
			Payload: &secretmanagerpb.SecretPayload{
				Data: []byte(value),
			},
		},
	)

	return ferrors.Wrapf(err, "unable to attach the value to secret: %s", name)
}

// DeleteSecret deletes a secret.
// - projects/<project>/secrets/<name>
func (c *SecretClient) DeleteSecret(name string) error {
	secretPath := strings.Split(name, "/")
	if len(secretPath) < minSecretPathLength {
		return ferrors.New("secret name is not in expected format")
	}

	err := c.client.DeleteSecret(
		c.ctx,
		&secretmanagerpb.DeleteSecretRequest{
			Name: name,
		},
	)
	return ferrors.Wrapf(err, "unable to delete secret: %s", name)
}

// Close closes the connection to the GCP Secret Service.
// The user should invoke this when the client is no longer required.
func (c *SecretClient) Close() error {
	return c.client.Close()
}
