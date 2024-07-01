package fetch

import (
	"context"
	"errors"
	"fmt"
	. "github.com/florinutz/git-intel/cmd/fetch/model"
	"github.com/go-git/go-git/v5"
	git_ssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/google/go-github/v62/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

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
		// todo orgs will be in the config file along with everything else
		Args: cobra.ExactArgs(1), // the org name, which should ve removed once it works
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
			// todo orgName as an arg once it works (it will be in the config file)
			orgName := args[0]

			token := os.Getenv("GITHUB_TOKEN")
			if token == "" {
				return fmt.Errorf("GITHUB_TOKEN environment variable not set")
			}

			client := github.NewClient(nil).WithAuthToken(token)

			ghListOpts := &github.RepositoryListByOrgOptions{
				Type:      RepoTypePrivate.String(),
				Sort:      RepoListSortUpdated.String(),
				Direction: "desc",
				ListOptions: github.ListOptions{
					PerPage: 100,
				},
			}

			ctx := context.Background()
			var orgRepos []*github.Repository
			for {
				pageRepos, resp, err := client.Repositories.ListByOrg(ctx, orgName, ghListOpts)
				if err != nil {
					return fmt.Errorf("error occurred while fetching org repositories: %w", err)
				}
				orgRepos = append(orgRepos, pageRepos...)
				if resp.NextPage == 0 {
					break
				}
				ghListOpts.Page = resp.NextPage
			}

			// Iterate through and print all the names of the organization's repositories
			fmt.Printf("%-70s %-30s\n", "Repo Name", "Last Updated")
			for _, repo := range orgRepos {
				fmt.Printf("%-70s %-30s\n", *repo.Name, repo.UpdatedAt.Format("2006-01-02"))
			}

			// clone the first 1 repo:
			if len(orgRepos) > 0 {
				repo := orgRepos[0]

				conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
				if err != nil {
					return fmt.Errorf("dialing the ssh agent failed: %w", err)
				}
				defer conn.Close()

				sshAgent := agent.NewClient(conn)

				auth := &git_ssh.PublicKeysCallback{
					User: "git",
					Callback: func() ([]ssh.Signer, error) {
						signers, err := sshAgent.Signers()
						if err != nil {
							return nil, fmt.Errorf("error occurred while getting ssh agent signers: %w", err)
						}

						// todo this single key signer also works, but let's have it disabled till the configs work and we can use it as a configurable method alongside the agent

						//keySigner, err := getSSHKeySigner(os.Getenv("HOME") + "/.ssh/id_ed25519")
						//if err != nil {
						//	return nil, fmt.Errorf("error occurred while getting ssh key signer: %w", err)
						//}
						//signers := []ssh.Signer{keySigner}

						return signers, nil
					},
				}

				clonePath := filepath.Join("repos", *repo.Name)
				if _, err := git.PlainCloneContext(ctx, clonePath, false, &git.CloneOptions{
					URL:   *repo.SSHURL,
					Depth: 1,
					Auth:  auth,
				}); err != nil {
					return fmt.Errorf("error occurred while cloning repo: %w", err)
				}
			}

			return nil
		},
	}

	viper.UnmarshalKey("paths", &opts)

	return
}

// getSSHKeySigner reads an SSH key from the given path and returns a signer
// todo use this when the configs work and we can use it as a configurable method alongside the agent
func getSSHKeySigner(sshKeyPath string) (ssh.Signer, error) {
	sshKey, err := os.ReadFile(sshKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error occurred while reading ssh key at %s: %w", sshKeyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(sshKey)
	if err != nil {
		var passphraseMissingError *ssh.PassphraseMissingError
		needsPassphrase := errors.As(err, &passphraseMissingError)
		if !needsPassphrase {
			return nil, fmt.Errorf("parsing private key failed: %w", err)
		}

		fmt.Print("Enter passphrase for encrypted key: ")
		passphraseBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, fmt.Errorf("read passphrase error: %w", err)
		}
		passphrase := strings.TrimSpace(string(passphraseBytes))
		fmt.Println()

		signer, err = ssh.ParsePrivateKeyWithPassphrase(sshKey, []byte(passphrase))
		if err != nil {
			return nil, fmt.Errorf("parsing private key with passphrase failed: %w", err)
		}
	}
	return signer, nil
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
