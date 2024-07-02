package model

type CloneOptions struct {
	Branch  string      `mapstructure:"branch,omitempty"`
	Depth   int         `mapstructure:"depth,omitempty"`
	Recurse bool        `mapstructure:"recurse,omitempty"` // For recursive cloning
	Auth    *AuthConfig `mapstructure:"auth,omitempty"`
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
	Name            string            `mapstructure:"name"`
	ExcludeRepos    []string          `mapstructure:"exclude_repos,omitempty"`
	OrgCloneOptions *CloneOptions     `mapstructure:"org_clone_options,omitempty"` // Custom Clone options per Org level
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

type Config struct {
	// tmp: org name (to be removed)
	OrgName string
	Paths   []Path        `mapstructure:"paths"`
	Clone   *CloneOptions `mapstructure:"global_clone_options,omitempty"`
	Auth    *AuthConfig   `mapstructure:"auth,omitempty"`
}
