package gcp

import (
	"context"
	"strings"
	"testing"
)

func TestGetTokenStruct(t *testing.T) {
	ctx := context.Background()
	token, err := GetTokenStruct(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if token == nil {
		t.Fatal("no token returned")
	}
}

func TestGetAccessToken(t *testing.T) {
	ctx := context.Background()
	token, err := GetAccessToken(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("no token returned")
	}
}

func TestGenHTTPHeader(t *testing.T) {
	ctx := context.Background()
	tStruct, err := GetTokenStruct(ctx)
	if err != nil {
		t.Fatal(err)
	}
	header, err := GenHTTPHeader(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(header, tStruct.TokenType) {
		t.Fatal("Token type not included")
	}

	if !strings.Contains(header, "ya29") {
		t.Fatal("Access token not included")
	}
}
