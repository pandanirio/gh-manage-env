package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pandanirio/gh-manage-env/internal/dotenv"
	"github.com/pandanirio/gh-manage-env/internal/gh"
)

type Options struct {
	Env             string
	File            string
	SecretPrefix    string
	DeleteMissing   bool
	DryRun          bool
	Yes             bool
	Repo            string // owner/repo
	KeepPrefix      bool
	StripExport     bool
	Verbose         bool
}

var opts Options

var rootCmd = &cobra.Command{
	Use:   "gh-manage-env",
	Short: "Sync .env to GitHub Actions variables and secrets (repo or environment scope)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if opts.File == "" {
			opts.File = ".env"
		}
		opts.File = filepath.Clean(opts.File)

		if opts.SecretPrefix == "" {
			opts.SecretPrefix = "SECURED_"
		}

		if err := gh.RequireInstalled("gh"); err != nil {
			return err
		}
		if err := gh.RequireAuth(); err != nil {
			return err
		}

		repo := opts.Repo
		if repo == "" {
			r, err := gh.DetectRepoFromGitRemote()
			if err != nil {
				return fmt.Errorf("unable to detect repo (use -R owner/repo): %w", err)
			}
			repo = r
		}

		entries, err := dotenv.ParseFile(opts.File)
		if err != nil {
			return err
		}

		// Split entries into secrets + variables
		secrets := map[string]string{}
		vars := map[string]string{}

		for k, v := range entries {
			isSecret := strings.HasPrefix(k, opts.SecretPrefix)
			name := k
			if isSecret && !opts.KeepPrefix {
				name = strings.TrimPrefix(k, opts.SecretPrefix)
			}
			if isSecret {
				secrets[name] = v
			} else {
				vars[name] = v
			}
		}

		// Ensure environment exists (idempotent) if -e
		if opts.Env != "" {
			if opts.DryRun {
				fmt.Printf("[dry-run] would ensure environment exists: %s (repo %s)\n", opts.Env, repo)
			} else {
				if err := gh.EnsureEnvironment(repo, opts.Env); err != nil {
					return err
				}
			}
		}

		// Upsert (batch via temp files)
		if len(secrets) > 0 {
			if err := gh.SetSecretsFromMap(repo, opts.Env, secrets, opts.DryRun, opts.Verbose); err != nil {
				return err
			}
		}
		if len(vars) > 0 {
			if err := gh.SetVariablesFromMap(repo, opts.Env, vars, opts.DryRun, opts.Verbose); err != nil {
				return err
			}
		}

		// Delete missing
		if opts.DeleteMissing {
			if !opts.Yes {
				// simple guard (you can remove if you want)
				fmt.Println("⚠️  -d enabled: this will delete secrets/variables not present in the file.")
				fmt.Println("Use --yes to skip confirmation.")
				fmt.Print("Continue? [y/N] ")
				var answer string
				fmt.Scanln(&answer)
				answer = strings.ToLower(strings.TrimSpace(answer))
				if answer != "y" && answer != "yes" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			if err := gh.DeleteMissing(repo, opts.Env, secrets, vars, opts.DryRun, opts.Verbose); err != nil {
				return err
			}
		}

		fmt.Println("✅ Done.")
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&opts.Env, "environment", "e", "", "GitHub Actions environment name (optional; if empty uses repository scope)")
	rootCmd.Flags().StringVarP(&opts.File, "file", "f", ".env", "dotenv file path")
	rootCmd.Flags().StringVarP(&opts.SecretPrefix, "secret-prefix", "s", "SECURED_", "prefix used to detect secrets")
	rootCmd.Flags().BoolVarP(&opts.DeleteMissing, "delete-missing", "d", false, "delete secrets/variables that are not present in the dotenv file")
	rootCmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print actions without executing")
	rootCmd.Flags().BoolVar(&opts.Yes, "yes", false, "skip confirmation prompts")
	rootCmd.Flags().StringVarP(&opts.Repo, "repo", "R", "", "repository in owner/repo format (optional; auto-detected from git remote)")
	rootCmd.Flags().BoolVar(&opts.KeepPrefix, "keep-prefix", false, "keep secret prefix in GitHub secret name (default strips it)")
	rootCmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "verbose output")
}
