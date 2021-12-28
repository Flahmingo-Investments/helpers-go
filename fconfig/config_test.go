package fconfig

import (
	"testing"
)

type Config struct {
	EnvTestVar    string `mapstructure:"envTestVar"`
	YamlTestVar   string `mapstructure:"yamlTestVar"`
	SecretTestVar string `mapstructure:"secretTestVar"`
	Nested        Nested `mapstructure:"nested"`
}

type Nested struct {
	val1 string `mapstructure:"val1"`
	val2 int    `mapstructure:"val2"`
	val3 bool   `mapstructure:"val3"`
}

func TestLoadConfig(t *testing.T) {
	var err error

	envTestVal := "EnvVarValue 1235543"
	yamlTestVal := "Yaml Test"
	if err != nil {
		t.Fatal(err)
	}
	files := []string{"./assets/config.yaml", "./assets/test.env"}

	config := &Config{}

	err = LoadConfig(files, config)

	if err != nil {
		t.Fatal(err)
	}

	if config.YamlTestVar != yamlTestVal {
		t.Fatal("Yaml var did not match")
	}
	if config.EnvTestVar != envTestVal {
		t.Fatal("Env var did not match")
	}
	if config.SecretTestVar != "test-value" {
		t.Fatal("Unable to get secret")
	}
	if config.Nested.val1 != "test" && config.Nested.val2 != 2 && config.Nested.val3 != false {
		t.Fatal("Error parsing nested values")
	}
}
