//go:build integration
// +build integration

package gcpauth

import (
	"os"
	"testing"
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
