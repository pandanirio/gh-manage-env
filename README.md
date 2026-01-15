[![release](https://github.com/pandanirio/gh-manage-env/actions/workflows/release.yml/badge.svg)](https://github.com/pandanirio/gh-manage-env/actions/workflows/release.yml)

# gh-manage-env

A CLI tool to synchronize `.env` files with GitHub Actions repository secrets, variables, and environment-specific secrets/variables.

## Prerequisites

### 1. Install GitHub CLI

The tool requires the GitHub CLI (`gh`) to be installed on your system.

**macOS:**
```bash
brew install gh
```

**Linux:**
```bash
# Debian/Ubuntu
sudo apt install gh

# Fedora
sudo dnf install gh

# Arch Linux
sudo pacman -S github-cli
```

**Windows:**
```powershell
winget install --id GitHub.cli
```

Or download from [GitHub CLI official website](https://cli.github.com/).

### 2. Authenticate with GitHub

After installing GitHub CLI, authenticate with your GitHub account:

```bash
gh auth login
```

Follow the interactive prompts to authenticate. You can choose between:
- GitHub.com or GitHub Enterprise Server
- Authentication method (web browser, token, etc.)

Verify your authentication:
```bash
gh auth status
```

## Installation

### Option 1: Using the install script

```bash
curl -fsSL https://raw.githubusercontent.com/pandanirio/gh-manage-env/main/scripts/install.sh | bash
```

Or install a specific version:
```bash
curl -fsSL https://raw.githubusercontent.com/pandanirio/gh-manage-env/main/scripts/install.sh | bash -s v1.0.0
```

### Option 2: Build from source

```bash
git clone https://github.com/pandanirio/gh-manage-env.git
cd gh-manage-env
go build -o gh-manage-env
sudo mv gh-manage-env /usr/local/bin/
```

### Option 3: Download pre-built binaries

Download the appropriate binary for your platform from the [Releases page](https://github.com/pandanirio/gh-manage-env/releases) and add it to your PATH.

## Usage

### Basic Usage

The tool automatically detects the repository from your git remote. Simply run:

```bash
gh-manage-env
```

This will:
- Read the `.env` file in the current directory
- Identify secrets (variables prefixed with `SECURED_` by default)
- Sync secrets and variables to your GitHub repository

### Example `.env` file

```env
# Regular variables (will be synced as GitHub Actions variables)
API_URL=https://api.example.com
DEBUG=false

# Secrets (prefixed with SECURED_)
SECURED_DATABASE_PASSWORD=mysecretpassword
SECURED_API_KEY=sk-1234567890
```

### Command-line Options

```bash
gh-manage-env [flags]
```

**Flags:**

- `-e, --environment <name>`: GitHub Actions environment name (optional; if empty uses repository scope)
- `-f, --file <path>`: Path to the dotenv file (default: `.env`)
- `-s, --secret-prefix <prefix>`: Prefix used to detect secrets (default: `SECURED_`)
- `--keep-prefix`: Keep secret prefix in GitHub secret name (default: strips the prefix)
- `-d, --delete-missing`: Delete secrets/variables that are not present in the dotenv file
- `--dry-run`: Print actions without executing them
- `--yes`: Skip confirmation prompts
- `-R, --repo <owner/repo>`: Repository in owner/repo format (optional; auto-detected from git remote)
- `-v, --verbose`: Enable verbose output

### Examples

**Sync to repository scope:**
```bash
gh-manage-env
```

**Sync to a specific environment:**
```bash
gh-manage-env -e production
```

**Use a custom .env file:**
```bash
gh-manage-env -f .env.production
```

**Use a custom secret prefix:**
```bash
gh-manage-env -s SECRET_
```

**Keep the prefix in secret names:**
```bash
gh-manage-env --keep-prefix
```

**Dry run to see what would happen:**
```bash
gh-manage-env --dry-run
```

**Delete secrets/variables not in the .env file:**
```bash
gh-manage-env -d
```

**Specify repository explicitly:**
```bash
gh-manage-env -R owner/repo
```

**Verbose output:**
```bash
gh-manage-env -v
```

## How It Works

1. **Reads your `.env` file**: Parses the file and extracts key-value pairs
2. **Identifies secrets**: Variables prefixed with `SECURED_` (or your custom prefix) are treated as secrets
3. **Strips prefix**: By default, the prefix is removed from secret names (e.g., `SECURED_API_KEY` becomes `API_KEY` in GitHub)
4. **Syncs to GitHub**: 
   - Secrets are synced using `gh secret set`
   - Variables are synced using `gh variable set`
5. **Environment support**: If `-e` is specified, secrets/variables are scoped to that environment

## Secret vs Variable

- **Secrets**: Sensitive data (passwords, API keys, tokens) - encrypted at rest in GitHub
- **Variables**: Non-sensitive configuration (URLs, feature flags) - stored as plain text

The tool automatically classifies entries based on the prefix you specify.

## Notes

- The tool requires appropriate GitHub permissions to manage secrets and variables
- Repository secrets/variables are available to all workflows
- Environment secrets/variables are only available to workflows that reference that environment
- The `--delete-missing` flag will permanently delete secrets/variables not in your `.env` file (use with caution)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

