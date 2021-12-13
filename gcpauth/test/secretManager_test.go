package test

import (
	"testing"

	"github.com/Flahmingo-Investments/helpers-go/gcpauth"
)

// This test assumes that either gcloud is setup or you are working inside of
func TestGetSecretByName(t *testing.T) {
	secret, err := gcpauth.GetSecretByName("projects/487165749144/secrets/test-secret/versions/latest")

	if err != nil {
		t.Error(err)
	}

	if secret != "test-value" {
		t.Error("wrong value")
	}
}
