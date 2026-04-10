# GitDigg

Download specific files from git repositories without cloning.

## Installation

```bash
npm install -g gitdigg
```

## Usage

```bash
# Download a single file
gitdigg owner/repo README.md

# Download a directory
gitdigg owner/repo src/utils

# Download with glob patterns
gitdigg owner/repo "src/*.ts"
gitdigg owner/repo "**/*.md"

# Specify branch and output directory
gitdigg owner/repo --branch dev -o ./output src/

# Increase download concurrency
gitdigg owner/repo -c 10 src/

# Flatten directory structure
gitdigg owner/repo --flat src/utils/

# Interactive file browser
gitdigg owner/repo -i
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
gitdigg https://github.com/owner/repo src/
gitdigg https://gitlab.com/owner/repo lib/
gitdigg https://bitbucket.org/owner/repo src/
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
gitdigg owner/repo --token your_token README.md
```

## Configuration

Create `~/.gitdigg.yaml`:

```yaml
output: ./downloads
concurrency: 8
flat: false
```

## License

MIT
