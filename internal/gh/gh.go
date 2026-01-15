package gh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func RequireInstalled(bin string) error {
	_, err := exec.LookPath(bin)
	if err != nil {
		return fmt.Errorf("GitHub CLI not found (%s). Install: https://cli.github.com/", bin)
	}
	return nil
}

func RequireAuth() error {
	// gh auth status returns 0 when authenticated
	cmd := exec.Command("gh", "auth", "status")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh not authenticated. Run: gh auth login")
	}
	return nil
}

func DetectRepoFromGitRemote() (string, error) {
	// read origin url
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return "", fmt.Errorf("git remote get-url origin failed: %w", err)
	}
	u := strings.TrimSpace(string(out))

	// Support:
	// - git@github.com:owner/repo.git
	// - https://github.com/owner/repo.git
	// - https://github.com/owner/repo
	u = strings.TrimSuffix(u, ".git")

	var path string
	if strings.HasPrefix(u, "git@github.com:") {
		path = strings.TrimPrefix(u, "git@github.com:")
	} else if strings.HasPrefix(u, "https://github.com/") {
		path = strings.TrimPrefix(u, "https://github.com/")
	} else {
		return "", fmt.Errorf("unsupported origin url: %s", u)
	}

	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("unable to parse owner/repo from: %s", u)
	}
	return parts[0] + "/" + parts[1], nil
}

func EnsureEnvironment(repo, env string) error {
	// PUT /repos/{owner}/{repo}/environments/{environment_name}
	// We'll call: gh api -X PUT repos/OWNER/REPO/environments/ENV
	_, err := run("gh", []string{"api", "-X", "PUT", fmt.Sprintf("repos/%s/environments/%s", repo, env)}, nil, false)
	if err != nil {
		return fmt.Errorf("ensure environment %s: %w", env, err)
	}
	return nil
}

func SetSecretsFromMap(repo, env string, secrets map[string]string, dryRun, verbose bool) error {
	tmp, err := writeEnvTempFile("secrets", secrets)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)

	args := []string{"secret", "set", "-f", tmp, "-R", repo}
	if env != "" {
		args = append(args, "--env", env)
	}
	return runPrint("gh", args, dryRun, verbose)
}

func SetVariablesFromMap(repo, env string, vars map[string]string, dryRun, verbose bool) error {
	tmp, err := writeEnvTempFile("vars", vars)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)

	args := []string{"variable", "set", "-f", tmp, "-R", repo}
	if env != "" {
		args = append(args, "--env", env)
	}
	return runPrint("gh", args, dryRun, verbose)
}

type nameItem struct {
	Name string `json:"name"`
}

func DeleteMissing(repo, env string, desiredSecrets map[string]string, desiredVars map[string]string, dryRun, verbose bool) error {
	existingSecrets, err := listNames("secret", repo, env)
	if err != nil {
		return err
	}
	existingVars, err := listNames("variable", repo, env)
	if err != nil {
		return err
	}

	// Delete secrets not in desiredSecrets
	for _, s := range existingSecrets {
		if _, ok := desiredSecrets[s]; ok {
			continue
		}
		args := []string{"secret", "delete", s, "-R", repo}
		if env != "" {
			args = append(args, "--env", env)
		}
		if err := runPrint("gh", args, dryRun, verbose); err != nil {
			return err
		}
	}

	// Delete vars not in desiredVars
	for _, v := range existingVars {
		if _, ok := desiredVars[v]; ok {
			continue
		}
		args := []string{"variable", "delete", v, "-R", repo}
		if env != "" {
			args = append(args, "--env", env)
		}
		if err := runPrint("gh", args, dryRun, verbose); err != nil {
			return err
		}
	}

	return nil
}

func listNames(kind, repo, env string) ([]string, error) {
	// kind: "secret" or "variable"
	args := []string{kind, "list", "--json", "name", "-R", repo}
	if env != "" {
		args = append(args, "--env", env)
	}
	stdout, err := run("gh", args, nil, false)
	if err != nil {
		return nil, err
	}

	var items []nameItem
	if err := json.Unmarshal(stdout, &items); err != nil {
		return nil, fmt.Errorf("parse gh %s list json: %w", kind, err)
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Name)
	}
	return out, nil
}

func writeEnvTempFile(prefix string, kv map[string]string) (string, error) {
	dir := os.TempDir()
	name := fmt.Sprintf("gh-manage-env_%s_%s.env", prefix, runtime.GOOS)
	path := filepath.Join(dir, name)

	var b strings.Builder
	for k, v := range kv {
		// dotenv style: KEY=VALUE (no quoting by default)
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(v)
		b.WriteString("\n")
	}

	if err := os.WriteFile(path, []byte(b.String()), 0600); err != nil {
		return "", fmt.Errorf("write temp env file: %w", err)
	}
	return path, nil
}

func runPrint(bin string, args []string, dryRun, verbose bool) error {
	if dryRun {
		fmt.Printf("[dry-run] %s %s\n", bin, strings.Join(args, " "))
		return nil
	}
	if verbose {
		fmt.Printf("[run] %s %s\n", bin, strings.Join(args, " "))
	}
	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func run(bin string, args []string, stdin []byte, verbose bool) ([]byte, error) {
	if verbose {
		fmt.Printf("[run] %s %s\n", bin, strings.Join(args, " "))
	}
	cmd := exec.Command(bin, args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("%s %s failed: %w\n%s", bin, strings.Join(args, " "), err, errBuf.String())
	}
	return out.Bytes(), nil
}
