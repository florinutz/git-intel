package fetch

import (
	"encoding/json"
	"fmt"
	. "github.com/florinutz/git-intel/cmd/fetch/model"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func BuildConfigGenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config-gen",
		Short: "Generates a skeleton TOML configuration file",
		Long:  `Generates a skeleton TOML configuration file using the configuration structure defined in fetch.go.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := Config{
				Paths: []Path{
					{
						Path: "/path/to/clone/repos",
						Repos: []RepoConfig{
							{
								Url: "https://github.com/example/repo1.git",
							},
							{
								Url: "https://github.com/example/repo2.git",
							},
						},
						Orgs: []GithubOrgConfig{
							{
								Name: "exampleOrg",
							},
						},
						Auth: &AuthConfig{
							Username: "username",
							Password: "password",
						},
					},
				},
				Clone: &CloneOptions{
					Branch: "main",
					Depth:  1,
				},
				Auth: &AuthConfig{
					Username: "username",
					Password: "password",
				},
			}

			configMap, err := getCleanConfigMap2(config)
			if err != nil {
				return fmt.Errorf("failed to generate config: %w", err)
			}

			// Use the map for viper
			viper.Set("fetch", configMap)

			viper.SetConfigType("yaml")
			viper.SetConfigFile("config.yml")

			if err := viper.WriteConfig(); err != nil {
				return err
			}

			return nil
		},
	}
}

func ptr[T any](t T) *T {
	return &t
}

func getCleanConfigMap(config Config) (map[string]interface{}, error) {
	jsonConfig, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	var configMap map[string]interface{}
	err = json.Unmarshal(jsonConfig, &configMap)
	if err != nil {
		return nil, err
	}
	cleanMap(configMap)
	return configMap, nil
}

func getCleanConfigMap2(config Config) (map[string]interface{}, error) {
	var configMap map[string]interface{}
	decoderConfig := &mapstructure.DecoderConfig{
		Result:  &configMap,
		TagName: "mapstructure",
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}
	cleanMap(configMap)

	return configMap, nil
}

func cleanMap(m map[string]interface{}) {
	for k, v := range m {
		switch v := v.(type) {
		case map[string]interface{}:
			cleanMap(v)
			if len(v) == 0 {
				delete(m, k)
			}
		case []interface{}:
			if len(v) == 0 {
				delete(m, k)
			} else {
				for i, item := range v {
					if itemMap, ok := item.(map[string]interface{}); ok {
						cleanMap(itemMap)
						if len(itemMap) == 0 {
							v = append(v[:i], v[i+1:]...)
						}
					}
				}
			}
		case string:
			if v == "" {
				delete(m, k)
			}
		case nil:
			delete(m, k)
		}
	}
}
