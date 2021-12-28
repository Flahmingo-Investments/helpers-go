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
var envFileRegex = regexp.MustCompile(`\.env$`)

// LoadConfigViper loads the configuration from a given file.
// For environment variables use the following format int the FILE. `${<Env var>}`
// For Google secrets, use the following format.
// gSecret://projects/<projectId>/secrets/<secretName>/versions/<version id or 'latest'>
// In order to parse the config, run v.Unmarshal(<config struct>), and ensure all your properties are exported/public
func LoadConfigViper(files []string) (*viper.Viper, error) {
	var err error

	v := viper.New()
	v.SetEnvPrefix("")
	v.AutomaticEnv()

	firstFile := true
	for _, file := range files {
		if envFileRegex.MatchString(file) {
			err = loadEnv(file)
			if err != nil {
				return nil, err
			}
		} else if firstFile {
			v.SetConfigFile(file)
			firstFile = false
		} else {
			v.AddConfigPath(file)
		}
	}

	err = v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	for _, key := range v.AllKeys() {
		val := v.GetString(key)

		v.Set(key, os.ExpandEnv(val))

		if secretRegex.MatchString(val) {
			secretPath := getSecretPath(val)
			secret, err := gcpauth.GetSecretByName(secretPath)
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
func LoadConfig(files []string, config interface{}) error {
	v, err := LoadConfigViper(files)
	if err != nil {
		return err
	}
	err = v.Unmarshal(config)
	if err != nil {
		return err
	}
	return nil
}

func getSecretPath(val string) string {
	matches := secretRegex.FindStringSubmatch(val)
	pathIndex := secretRegex.SubexpIndex("Path")
	return matches[pathIndex]
}

// loadEnv load environments variables from a file.
// If no file name is given it will try to load .env file.
func loadEnv(filename string) error {
	var err error

	if filename != "" {
		err = godotenv.Load(filename)
	} else {
		err = godotenv.Load()
		// it is fine if .env does not exists
		if os.IsNotExist(err) {
			return nil
		}
	}

	return ferrors.WithStack(err)
}
