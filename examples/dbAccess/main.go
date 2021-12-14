package main

import (
	"fmt"

	"github.com/Flahmingo-Investments/helpers-go/flog"
	"github.com/Flahmingo-Investments/helpers-go/gcpauth"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	const saEmail string = "farsos-test@development-flahmingo.iam"
	token, err := gcpauth.GetAuthToken(saEmail)
	if err != nil {
		flog.Errorf("Failed to get auth token", err)
		return
	}

	connString := fmt.Sprintf(
		"user=%s dbname=postgres sslmode=disable password=%s",
		saEmail, token)

	//this automatically runs db.ping()
	_, err = sqlx.Connect("postgres", connString)
	if err != nil {
		flog.Errorf("DB conn failed", err)
	}
	flog.Infof("Successful connection")
}
