package test

import (
	"os"
	"testing"

	"github.com/Flahmingo-Investments/helpers-go/gcpauth"
)

// This test assumes that either gcloud is setup or you are working inside of
func TestGetSecretByName(t *testing.T) {
	secretName := os.Getenv("SecretPath")
	secret, err := gcpauth.GetSecretByName(secretName)

	if err != nil {
		t.Error(err)
	}

	if secret != "test-value" {
		t.Error("wrong value")
	}
}
