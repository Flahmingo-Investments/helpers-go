//go:build integration
// +build integration

package fconfig

import (
	"io"
	"os"
	"testing"
)

type Config struct {
	EnvTestVar    string `mapstructure:"envTestVar"`
	YamlTestVar   string `mapstructure:"yamlTestVar"`
	SecretTestVar string `mapstructure:"secretTestVar"`
	Nested        Nested `mapstructure:"nested"`
}

type Nested struct {
	Val1 string `mapstructure:"val1"`
	Val2 int    `mapstructure:"val2"`
	Val3 bool   `mapstructure:"val3"`
}

func TestLoadConfig(t *testing.T) {
	var err error

	envTestVal := "EnvVarValue 1235543"
	yamlTestVal := "Yaml Test"
	if err != nil {
		t.Fatal(err)
	}

	copyEnvFile(t)
	defer deleteEnvFile()

	config := &Config{}

	err = LoadConfig("./testdata/config.yaml", config)

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
	if config.Nested.Val1 != "test" && config.Nested.Val2 != 2 && config.Nested.Val3 != true {
		t.Fatal("Error parsing nested values")
	}
}

var envFilePath = "./testdata/test.env"
var envFileDest = "./.env"

func copyEnvFile(t *testing.T) {
	_, err := os.Stat(envFilePath)
	if err != nil {
		t.Fatal(err)
	}

	source, err := os.Open(envFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()

	destination, err := os.Create(envFileDest)
	if err != nil {
		t.Fatal(err)
	}
	defer destination.Close()
	io.Copy(destination, source)
}

func deleteEnvFile() {
	os.Remove(envFileDest)
}
