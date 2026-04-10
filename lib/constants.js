// API base URLs for different providers
export const API_URLS = {
  github: 'https://api.github.com',
  gitlab: 'https://gitlab.com/api/v4',
  bitbucket: 'https://api.bitbucket.org/2.0',
};

// Raw content URLs
export const RAW_URLS = {
  github: 'https://raw.githubusercontent.com',
  gitlab: 'https://gitlab.com',
  bitbucket: 'https://bitbucket.org',
};

// Default configuration
export const DEFAULTS = {
  concurrency: 5,
  retries: 3,
  retryDelay: 1000,
  timeout: 30000,
  branch: 'main',
  outputDir: '.',
};

// Config file paths
export const CONFIG_PATHS = [
  '.gitdigg.yaml',
  '.gitdigg.yml',
  '~/.gitdigg.yaml',
  '~/.gitdigg.yml',
  '~/.config/gitdigg/config.yaml',
];

// Provider hostnames
export const PROVIDER_HOSTS = {
  'github.com': 'github',
  'gitlab.com': 'gitlab',
  'bitbucket.org': 'bitbucket',
};

// File type indicators for interactive mode
export const FILE_ICONS = {
  directory: '\u2514',
  file: '\u2500',
};
