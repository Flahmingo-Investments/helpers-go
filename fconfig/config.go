// Package fconfig provides support to load config files and expand secrets.
package fconfig

import (
	"os"
	"reflect"
	"regexp"

	"github.com/Flahmingo-Investments/helpers-go/ferrors"
	"github.com/Flahmingo-Investments/helpers-go/gcp"
	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var secretRegex = regexp.MustCompile(`^gSecret://(?P<Path>.+)`)

// secretClient is helper to expand gcp.SecretClient to support gSecret in path.
type secretClient struct {
	*gcp.SecretClient
}

// getSecret parses a `gSecret://` string into a GCP secret path, and retrieve
// it from GCP Secret Service.
func (c *secretClient) getSecret(val string) (string, error) {
	matches := secretRegex.FindStringSubmatch(val)
	pathIndex := secretRegex.SubexpIndex("Path")
	path := matches[pathIndex]
	return c.GetSecret(path)
}

func decodeEnvVars() mapstructure.DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		// Check that the data is string
		if f.Kind() != reflect.String {
			return data, nil
		}

		// expand environment variables inside config file
		// e.g. ${ENV_NAME}
		eval := os.Expand(data.(string), func(str string) string {
			if str == "$" {
				return "$"
			}

			return os.Getenv(str)
		})

		return eval, nil
	}
}

func decodeGSecret(sc *secretClient) mapstructure.DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		// Check that the data is string
		if f.Kind() != reflect.String {
			return data, nil
		}

		if secretRegex.MatchString(data.(string)) {
			if sc == nil {
				gsc, err := gcp.NewSecretClient()
				if err != nil {
					return "", err
				}
				// wrap gsc into secretClient to support `gSecret://` expansion.
				sc = &secretClient{SecretClient: gsc}
			}

			secret, err := sc.getSecret(data.(string))
			if err != nil {
				return data, err
			}

			return secret, nil
		}

		return data, nil
	}
}

// loadConfig loads the configuration from a given file.
//
// It expands the environment variables if the value matches `${ENV_NAME}`.
// It fetches the secret if the value matches gSecret://uri.
func loadConfig(file string, config interface{}) error {
	// initialize viper and set the config file to read from.
	v := viper.New()
	v.SetConfigFile(file)

	// Read the configuration.
	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	// Declaring a client early not initializing it.
	// So, we can initailize it only when we find a 'gSecret'.
	var sc *secretClient

	defer func() {
		if sc != nil {
			_ = sc.Close()
		}
	}()

	return v.Unmarshal(config,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
				decodeEnvVars(),
				decodeGSecret(sc),
			),
		))
}

// LoadConfig loads the configuration from a given file and unmarshal it into
// the provided config.
// It maps the fields using `mapstructure` tag.
//
// Example:
//
//	type Config struct {
//		 Env   string `mapstructure:env`
//		 DBUri string `mapstructure:db_uri`
//	}
//
// It expands the environment variables if the value matches `${ENV_NAME}`.
// It fetches the secret if the value matches gSecret://uri.
func LoadConfig(file string, config interface{}) error {
	// load .env file inside the current working directory.
	err := LoadEnv("")
	if err != nil {
		return ferrors.Wrap(err, "unable to read environment variables")
	}

	// load configuration into the provided config.
	err = loadConfig(file, config)
	if err != nil {
		return err
	}

	return nil
}

// LoadEnv load environments variables from a file.
// If no file name is given it will try to load .env file.
func LoadEnv(filename string) error {
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
