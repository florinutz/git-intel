package fetch

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Paths []Path        `mapstructure:"paths"`
	Clone *CloneOptions `mapstructure:"global_clone_options,omitempty"`
	Auth  *AuthConfig   `mapstructure:"auth,omitempty"`
}

type CloneOptions struct {
	Branch  string      `mapstructure:"branch,omitempty"`
	Depth   int         `mapstructure:"depth,omitempty"`
	Recurse bool        `mapstructure:"recurse,omitempty"` // For recursive cloning
	Auth    *AuthConfig `mapstructure:"auth,omitempty"`
}

type RepoFilterConfig struct {
	Forks            *int     `mapstructure:"forks,omitempty"`             // Minimum forks count
	Stars            *int     `mapstructure:"stars,omitempty"`             // Minimum stars count
	UpdatedAfter     *string  `mapstructure:"updated_after,omitempty"`     // Only include repos updated after this date
	UpdatedBefore    *string  `mapstructure:"updated_before,omitempty"`    // Only include repos updated before this date
	IncludeLanguages []string `mapstructure:"include_languages,omitempty"` // Only include repos with these languages
	ExcludeLanguages []string `mapstructure:"exclude_languages,omitempty"` // Exclude repos with these languages
}

type RepoOrderConfig struct {
	Field     *string `mapstructure:"field"`     // Field to sort by. Valid values are "created", "updated", "pushed", "full_name", "size"
	Direction *string `mapstructure:"direction"` // Direction to sort. Either "asc" or "desc".
}

type GithubOrgConfig struct {
	Name            string            `mapstructure:"name"`
	ExcludeRepos    []string          `mapstructure:"exclude_repos,omitempty"`
	OrgCloneOptions *CloneOptions     `mapstructure:"clone,omitempty"` // Custom Clone options per Org level
	Auth            *AuthConfig       `mapstructure:"auth,omitempty"`
	RepoFilter      *RepoFilterConfig `mapstructure:"repo_filter,omitempty"` // Filter out repositories based on certain conditions
	RepoOrder       *RepoOrderConfig  `mapstructure:"repo_order,omitempty"`  // Order repositories based on a field
	RepoLimit       int               `mapstructure:"repo_limit,omitempty"`  // Limit the number of repositories to be cloned
}

type RepoConfig struct {
	Url              string        `mapstructure:"url"`
	RepoCloneOptions *CloneOptions `mapstructure:"repo_clone_options,omitempty"` // Custom Clone options per repo level
	Auth             *AuthConfig   `mapstructure:"auth,omitempty"`
}

type Path struct {
	Path         string            `mapstructure:"path"`
	Repos        []RepoConfig      `mapstructure:"repos,omitempty"`
	Orgs         []GithubOrgConfig `mapstructure:"orgs,omitempty"`
	CloneOptions *CloneOptions     `mapstructure:"path_clone_options,omitempty"` // Custom Clone options per Path level
	Auth         *AuthConfig       `mapstructure:"auth,omitempty"`
}

type AuthConfig struct {
	Username   string `mapstructure:"username,omitempty"`    // for Basic Auth
	Password   string `mapstructure:"password,omitempty"`    // for Basic Auth
	SSHKey     string `mapstructure:"ssh_key,omitempty"`     // for SSH
	OAuthToken string `mapstructure:"oauth_token,omitempty"` // for OAuth
}

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

			// Marshal the config to JSON
			jsonConfig, err := json.Marshal(config)
			if err != nil {
				return err
			}

			// Unmarshal the JSON back to a map
			var configMap map[string]interface{}
			err = json.Unmarshal(jsonConfig, &configMap)
			if err != nil {
				return err
			}

			// Clean up the map
			cleanMap(configMap)

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
