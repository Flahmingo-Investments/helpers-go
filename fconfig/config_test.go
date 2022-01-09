//go:build integration
// +build integration

package fconfig

import (
	"reflect"
	"testing"
)

type config struct {
	EnvTestVar    string `mapstructure:"envTestVar"`
	YamlTestVar   string `mapstructure:"yamlTestVar"`
	SecretTestVar string `mapstructure:"secretTestVar"`
	Nested        nested `mapstructure:"nested"`
}

type nested struct {
	Val1 string `mapstructure:"val1"`
	Val2 int    `mapstructure:"val2"`
	Val3 bool   `mapstructure:"val3"`
}

func TestLoadConfig(t *testing.T) {
	// TODO: Find a way to replace project id from config file.
	testCases := []struct {
		name       string
		configName string
		envName    string
		want       *config
		wantErr    bool
	}{
		{
			name:       "should load simple config",
			configName: "testdata/config.yaml",
			envName:    "testdata/test.env",
			want: &config{
				EnvTestVar:    "EnvVarValue 1235543",
				YamlTestVar:   "Yaml Test",
				SecretTestVar: "test-value",
				Nested: nested{
					Val1: "test",
					Val2: 2,
					Val3: true,
				},
			},
		},
		{
			name:       "config without secrets",
			configName: "testdata/configWoSecret.yaml",
			want: &config{
				EnvTestVar:    "",
				YamlTestVar:   "Yaml Test",
				SecretTestVar: "",
				Nested: nested{
					Val1: "test",
					Val2: 2,
					Val3: true,
				},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			got := &config{}

			if tc.envName != "" {
				LoadEnv(tc.envName)
			}

			err := LoadConfig(tc.configName, got)
			if tc.wantErr && err == nil {
				t.Errorf("expected error but got nil")
				return
			}
			if err != nil {
				t.Errorf("expected error to be nil but, got error: %+v", err)
				return
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("expected = %+v but, got = %+v", tc.want, got)
				return
			}
		})
	}
}
