//go:build integration
// +build integration

package gcp

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// This test assumes that either gcloud is setup or you are working inside of
func TestGetSecretByName(t *testing.T) {
	secretName := os.Getenv("SECRET_PATH")
	secret, err := GetSecretByName(secretName)
	if err != nil {
		t.Error(err)
	}

	if secret != "test-value" {
		t.Error("wrong value")
	}
}

// TestSecretClient is a quick integration test to validate create, get and
// delete secret.
func TestSecretClient(t *testing.T) {
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		t.Fatalf("env GCP_PROJECT_ID is required to run the integration test")
	}

	var testClient *SecretClient

	t.Cleanup(func() {
		if testClient == nil {
			return
		}
		if closingErr := testClient.Close(); closingErr != nil {
			// closing connection is an error but it is not an error that we care
			// about in the tests. So, a normal log is sufficient for our usecase.
			t.Logf("failed to close connection: %+v", closingErr)
		}
	})

	// Creating a unique key so, we don't go and delete the key if somebody else
	// is running the test.
	key := fmt.Sprintf("projects/%s/secrets/test-secret-client-%d", projectID, time.Now().Unix())

	t.Run("should create a client", func(t *testing.T) {
		client, err := NewSecretClient()
		if err != nil {
			t.Errorf("expected to create a client but, got an error: %v", err)
			return
		}
		testClient = client
	})

	t.Run("should create a secret", func(t *testing.T) {
		err := testClient.CreateSecret(key, "test-secret-client-value")
		if err != nil {
			t.Errorf("expected to create a secret but, got an error: %+v", err)
			return
		}
	})

	t.Run("should get the created secret", func(t *testing.T) {
		want, err := testClient.GetSecret(key)
		if err != nil {
			t.Errorf("expected to get secret, but got an error: %+v", err)
			return
		}

		if want != "test-secret-client-value" {
			t.Errorf("expected to get secret, but got an error: %+v", err)
			return
		}
	})

	t.Run("should delete the created secret", func(t *testing.T) {
		err := testClient.DeleteSecret(key)
		if err != nil {
			t.Errorf("expected to delete the secret, but got an error: %+v", err)
			return
		}
	})
}
