package gcp

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func GetAccessToken(ctx context.Context, scopes ...string) (string, error) {
	token, err := GetTokenStruct(ctx, scopes...)
	if err != nil {
		return "", nil
	}
	return token.AccessToken, nil
}

func GetTokenStruct(ctx context.Context, scopes ...string) (*oauth2.Token, error) {
	credentials, err := google.FindDefaultCredentials(ctx, scopes...)
	if err != nil {
		return nil, err
	}

	token, err := credentials.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	return token, nil
}

func GenHTTPHeader(ctx context.Context, scopes ...string) (string, error) {
	token, err := GetTokenStruct(ctx, scopes...)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s", token.TokenType, token.AccessToken), err
}
