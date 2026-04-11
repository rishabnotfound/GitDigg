<p align="center">
      <img
        src="./assets/logo.png"
        width="200"
        height="200"
      />
    </p>

<p align="center">
  Download specific files from git repositories without cloning.
  <br />
  <br><a href="https://gitdigg.vercel.app/">SITE Link</a></br><a href="https://www.npmjs.com/package/gitdigg">NPM Link</a>
</p>


## Installation

```bash
npm install -g gitdigg
```

Requires Node.js 18+.

## Preview

<p align="center">
    <img src="./assets/preview.gif" width="900" height="400" />
    <img width="1203" height="679" alt="Screenshot 2026-04-12 at 1 05 55 AM" src="https://github.com/user-attachments/assets/e74091a8-1c4b-4ff6-b8a5-92220d86456b" />
</p>

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

### Options

```
-b, --branch <branch>  Branch, tag, or commit to download from
-o, --output <dir>     Output directory (default: ".")
-i, --interactive      Interactive mode - browse and select files
-c, --concurrency <n>  Number of concurrent downloads (default: 5)
-t, --token <token>    Authentication token for private repositories
--flat                 Download all files without preserving structure
-v, --verbose          Verbose output
-q, --quiet            Suppress output
--retries <n>          Number of retries for failed downloads (default: 3)
```

### Interactive Mode Controls

| Key | Action |
|-----|--------|
| `↑` `↓` | Navigate |
| `→` | Expand directory |
| `←` | Collapse directory |
| `Space` | Toggle selection |
| `a` | Toggle select all |
| `e` | Toggle expand all |
| `/` | Search |
| `Enter` | Download selected |
| `q` / `Esc` | Quit |

### Full URLs

```bash
gitdigg https://github.com/owner/repo src/
gitdigg https://gitlab.com/owner/repo lib/
```

## Authentication

For private repositories, use one of these methods:

### 1. Token in URL (easiest)

```bash
gitdigg https://user:TOKEN@github.com/owner/repo README.md
gitdigg https://user:TOKEN@gitlab.com/owner/repo README.md
```

### 2. Command line flag

```bash
gitdigg owner/repo README.md --token YOUR_TOKEN
```

### 3. Environment variables

```bash
# GitHub
export GITHUB_TOKEN=your_token
# or
export GH_TOKEN=your_token

# GitLab
export GITLAB_TOKEN=your_token
```

## Configuration

Create `~/.gitdigg.yaml` for persistent settings:

```yaml
concurrency: 8
retries: 3
outputDir: ./downloads

tokens:
  github: your_github_token
  gitlab: your_gitlab_token
```

## License

Built with ❤️ by Rishab

[MIT](LICENSE)
