# GitDig

Download specific files from git repositories without cloning.

## Installation

```bash
npm install -g gitdig
```

## Usage

```bash
# Download a single file
gitdig owner/repo README.md

# Download a directory
gitdig owner/repo src/utils

# Download with glob patterns
gitdig owner/repo "src/*.ts"
gitdig owner/repo "**/*.md"

# Specify branch and output directory
gitdig owner/repo --branch dev -o ./output src/

# Increase download concurrency
gitdig owner/repo -c 10 src/

# Flatten directory structure
gitdig owner/repo --flat src/utils/

# Interactive file browser
gitdig owner/repo -i
```

### Interactive Mode Controls

| Key | Action |
|-----|--------|
| `↑/k` `↓/j` | Navigate |
| `←/h` `→/l` | Collapse/Expand |
| `Space` | Toggle selection |
| `a` | Select all |
| `A` | Deselect all |
| `/` | Search |
| `d` | Download |
| `q` | Quit |

### Full URLs

```bash
gitdig https://github.com/owner/repo src/
gitdig https://gitlab.com/owner/repo lib/
gitdig https://bitbucket.org/owner/repo src/
```

## Authentication

For private repositories:

```bash
# GitHub
export GITHUB_TOKEN=your_token

# GitLab
export GITLAB_TOKEN=your_token

# Bitbucket
export BITBUCKET_TOKEN=your_token
```

Or use the `--token` flag:

```bash
gitdig owner/repo --token your_token README.md
```

## Configuration

Create `~/.gitdig.yaml`:

```yaml
output: ./downloads
concurrency: 8
flat: false
```

## License

MIT
