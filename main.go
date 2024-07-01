package main

import (
	"fmt"
	"github.com/florinutz/git-intel/cmd/fetch"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	rootCmd := buildRootCommand()
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(rootCmd.ErrOrStderr(), "%v\n", err)
	}
}

func buildRootCommand() *cobra.Command {
	var opts struct {
		cfgFile string
	}

	cmd := &cobra.Command{
		Use:     "git-intel",
		Short:   "extracts information from a github org's git repositories",
		Long:    "extracts information from a github org's private and public git repositories",
		Version: "0.0.1",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if opts.cfgFile != "" {
				viper.SetConfigFile(opts.cfgFile)
				if err := viper.ReadInConfig(); err != nil {
					fmt.Printf("Warning: Can't read config file: %s\n", err)
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Welcome to git-org-manager!")
		},
	}

	cmd.AddCommand(
		fetch.BuildFetchCmd(),
		fetch.BuildConfigGenCmd(),
	)

	cmd.PersistentFlags().StringVarP(&opts.cfgFile, "config", "c", "", "config file")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		cobra.CheckErr(fmt.Sprintf("Can't use config file: %s", viper.ConfigFileUsed()))
	}

	return cmd
}
