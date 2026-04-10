package cli

import (
	"fmt"
	"os"

	"github.com/rishabnotfound/gitdig/internal/provider"
	"github.com/rishabnotfound/gitdig/internal/tui"
	"github.com/rishabnotfound/gitdig/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "gitdig <repo> [paths...]",
	Short: "Download specific files from git repositories without cloning",
	Long: `GitDig downloads specific files and directories from git repositories
without cloning the entire repo. Supports GitHub, GitLab, and Bitbucket.

Examples:
  gitdig owner/repo README.md
  gitdig owner/repo src/utils
  gitdig owner/repo "src/*.ts"
  gitdig owner/repo --branch dev -o ./output src/
  gitdig owner/repo -i`,
	Version: version.Short(),
	RunE:    run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.gitdig.yaml)")
	rootCmd.Flags().StringP("branch", "b", "", "branch, tag, or commit")
	rootCmd.Flags().StringP("output", "o", ".", "output directory")
	rootCmd.Flags().IntP("concurrency", "c", 4, "parallel downloads")
	rootCmd.Flags().Bool("flat", false, "flatten directory structure")
	rootCmd.Flags().BoolP("interactive", "i", false, "interactive mode")
	rootCmd.Flags().String("token", "", "auth token")

	viper.BindPFlag("branch", rootCmd.Flags().Lookup("branch"))
	viper.BindPFlag("output", rootCmd.Flags().Lookup("output"))
	viper.BindPFlag("concurrency", rootCmd.Flags().Lookup("concurrency"))
	viper.BindPFlag("flat", rootCmd.Flags().Lookup("flat"))
	viper.BindPFlag("token", rootCmd.Flags().Lookup("token"))

	rootCmd.SetVersionTemplate(version.Info() + "\n")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
			viper.SetConfigName(".gitdig")
			viper.SetConfigType("yaml")
		}
	}

	viper.SetEnvPrefix("GITDIG")
	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("repository required\n\nUsage: gitdig <repo> [paths...]\nRun 'gitdig --help' for details")
	}

	interactive, _ := cmd.Flags().GetBool("interactive")
	if interactive {
		return runInteractive(args[0])
	}

	if len(args) < 2 {
		return fmt.Errorf("path required\n\nUsage: gitdig <repo> <path> [paths...]\nOr use: gitdig <repo> -i")
	}

	return downloadFiles(args[0], args[1:])
}

func runInteractive(repo string) error {
	token := viper.GetString("token")
	prov, repoInfo, err := provider.DetectAndGetProvider(repo, token)
	if err != nil {
		return err
	}

	ref := viper.GetString("branch")
	if ref == "" && repoInfo.Ref != "" {
		ref = repoInfo.Ref
	}

	return tui.Run(prov, repoInfo, ref, viper.GetString("output"), viper.GetInt("concurrency"), viper.GetBool("flat"))
}
