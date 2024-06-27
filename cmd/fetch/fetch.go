package fetch

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type CloneOptions struct {
	Branch  string     `mapstructure:"branch,omitempty"`
	Depth   int        `mapstructure:"depth,omitempty"`
	Recurse bool       `mapstructure:"recurse,omitempty"` // For recursive cloning
	Auth    AuthConfig `mapstructure:"auth,omitempty"`
}

type RepoFilterConfig struct {
	Forks            int      `mapstructure:"forks,omitempty"`             // Minimum forks count
	Stars            int      `mapstructure:"stars,omitempty"`             // Minimum stars count
	UpdatedAfter     string   `mapstructure:"updated_after,omitempty"`     // Only include repos updated after this date
	UpdatedBefore    string   `mapstructure:"updated_before,omitempty"`    // Only include repos updated before this date
	IncludeLanguages []string `mapstructure:"include_languages,omitempty"` // Only include repos with these languages
	ExcludeLanguages []string `mapstructure:"exclude_languages,omitempty"` // Exclude repos with these languages
}

type RepoOrderConfig struct {
	Field     string `mapstructure:"field"`     // Field to sort by. Valid values are "created", "updated", "pushed", "full_name", "size"
	Direction string `mapstructure:"direction"` // Direction to sort. Either "asc" or "desc".
}

type GithubOrgConfig struct {
	Name            string           `mapstructure:"name"`
	ExcludeRepos    []string         `mapstructure:"exclude_repos,omitempty"`
	OrgCloneOptions CloneOptions     `mapstructure:"org_clone_options,omitempty"` // Custom Clone options per Org level
	Auth            AuthConfig       `mapstructure:"auth,omitempty"`
	RepoFilter      RepoFilterConfig `mapstructure:"repo_filter,omitempty"` // Filter out repositories based on certain conditions
	RepoOrder       RepoOrderConfig  `mapstructure:"repo_order,omitempty"`  // Order repositories based on a field
	RepoLimit       int              `mapstructure:"repo_limit,omitempty"`  // Limit the number of repositories to be cloned
}

type RepoConfig struct {
	Url              string        `mapstructure:"url"`
	RepoCloneOptions *CloneOptions `mapstructure:"repo_clone_options,omitempty"` // Custom Clone options per repo level
	Auth             AuthConfig    `mapstructure:"auth,omitempty"`
}

type Path struct {
	Path         string            `mapstructure:"path"`
	Repos        []RepoConfig      `mapstructure:"repos,omitempty"`
	Orgs         []GithubOrgConfig `mapstructure:"orgs,omitempty"`
	CloneOptions *CloneOptions     `mapstructure:"path_clone_options,omitempty"` // Custom Clone options per Path level
	Auth         AuthConfig        `mapstructure:"auth,omitempty"`
}

type AuthConfig struct {
	Username   string `mapstructure:"username,omitempty"`    // for Basic Auth
	Password   string `mapstructure:"password,omitempty"`    // for Basic Auth
	SSHKey     string `mapstructure:"ssh_key,omitempty"`     // for SSH
	OAuthToken string `mapstructure:"oauth_token,omitempty"` // for OAuth
}

type Config struct {
	Paths []Path       `mapstructure:"paths"`
	Clone CloneOptions `mapstructure:"global_clone_options,omitempty"`
	Auth  AuthConfig   `mapstructure:"auth,omitempty"`
}

// BuildFetchCmd clones repos
func BuildFetchCmd() (cmd *cobra.Command) {
	var opts Config

	cmd = &cobra.Command{
		Use:   "fetch",
		Short: "Fetch github repos, clone them to specified directories",
		Long: `Fetch multiple github repositories and clone them to specified directories.
The repositories can be specified by clone URL or by organization.
The organization repos list can have skipped repos.
The configuration is read from the config file specified by the --config flag.
The config file should be in YAML format and have a 'paths' key with a list of paths to clone the repos to.
Each path should have a 'path' key with the directory to clone the repos to, and either a 'repos' key with a list of clone URLs or an 'org' key with the organization name.
The 'org' key can also have an 'org_exceptions' key with a list of repos to skip.

Example config file:
paths:

`,
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			for _, pathConfig := range opts.Paths {
				if len(pathConfig.Repos) == 0 && len(pathConfig.Orgs) == 0 {
					return fmt.Errorf("at least one GitHub URL or an organization is required for each path")
				}
				for _, repo := range pathConfig.Repos {
					err := validateRepo(repo.Url)
					if err != nil {
						return err
					}
				}
				for _, org := range pathConfig.Orgs {
					err := validateOrg(org.Name)
					if err != nil {
						return err
					}
				}

				// validate that the directory exists, that it's a directory and that it's writeable:
				if err := validateTargetPath(pathConfig.Path); err != nil {
					return err
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Fetching repositories...")
			return nil
		},
	}

	viper.UnmarshalKey("paths", &opts)

	return
}

func validateRepo(repoUrl string) error {
	// your validation logic for repo URL goes here
	return nil
}

func validateOrg(orgName string) error {
	// your validation logic for orgName goes here
	return nil
}

// validateTargetPath validates a path as a valid candidate for hosting the cloned git repos
func validateTargetPath(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}

	fi, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("target path '%s' does not exist: %w", path, err)
		}
		return fmt.Errorf("can't stat target path '%s': %w", path, err)
	}

	if !fi.IsDir() {
		return fmt.Errorf("target path '%s' is not a directory", path)
	}

	testFile := filepath.Join(path, ".testfile")
	f, err := os.Create(testFile)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("target path '%s' is not writeable", path)
		}
		return fmt.Errorf("failed to test write to target path '%s'", path)
	}
	f.Close()
	os.Remove(testFile)

	return nil
}
