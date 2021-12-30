package fconfig

import (
	"os"
	"regexp"

	"github.com/Flahmingo-Investments/helpers-go/ferrors"
	"github.com/Flahmingo-Investments/helpers-go/gcpauth"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var secretRegex = regexp.MustCompile(`^gSecret://(?P<Path>.+)`)

// readConfig loads the configuration from a given file.
// For environment variables use the following format int the FILE. `${<Env var>}`
// For Google secrets, use the following format.
// gSecret://projects/<projectId>/secrets/<secretName>/versions/<version id or 'latest'>
// In order to parse the config, run v.Unmarshal(<config struct>), and ensure all your properties are exported/public
func readConfig(file string) (*viper.Viper, error) {
	var err error

	v := viper.New()
	v.SetConfigFile(file)
	err = loadEnv()
	if err != nil {
		return nil, err
	}

	if err != nil {

	}

	err = v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	for _, key := range v.AllKeys() {
		val := v.GetString(key)

		v.Set(key, os.ExpandEnv(val))

		if secretRegex.MatchString(val) {
			secret, err := getSecret(val)
			if err != nil {
				return nil, err
			}
			v.Set(key, secret)
		}
	}

	return v, nil
}

// LoadConfig Read the config files, and modify the conifg object
// For environment variables use the following format int the FILE. `${<Env var>}`
// For Google secrets, use the following format.
// gSecret://projects/<projectId>/secrets/<secretName>/versions/<version id or 'latest'>
// This function automatically modifies the config file. Ensure all your properties are exported/public
func LoadConfig(file string, config interface{}) error {
	v, err := readConfig(file)
	if err != nil {
		return err
	}
	err = v.Unmarshal(config)
	if err != nil {
		return err
	}
	return nil
}

// getSecret parse a `gSecret://` string into a gcp secret path, and retrieve it from storage
func getSecret(val string) (string, error) {
	matches := secretRegex.FindStringSubmatch(val)
	pathIndex := secretRegex.SubexpIndex("Path")
	path := matches[pathIndex]
	return gcpauth.GetSecretByName(path)
}

// loadEnv load environments variables from a file.
// If no file name is given it will try to load .env file.
func loadEnv() error {
	var err error
	err = godotenv.Load()
	return ferrors.WithStack(err)
}
