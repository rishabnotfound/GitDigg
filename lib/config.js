import { readFileSync, existsSync } from 'node:fs';
import { homedir } from 'node:os';
import { join } from 'node:path';
import YAML from 'yaml';
import { DEFAULTS, CONFIG_PATHS } from './constants.js';

let loadedConfig = null;

/**
 * Find and load configuration file
 */
function findConfigFile() {
  for (const configPath of CONFIG_PATHS) {
    const resolved = configPath.startsWith('~')
      ? join(homedir(), configPath.slice(1))
      : configPath;

    if (existsSync(resolved)) {
      return resolved;
    }
  }
  return null;
}

/**
 * Load configuration from file
 */
export function loadConfig() {
  if (loadedConfig) return loadedConfig;

  const configFile = findConfigFile();
  let fileConfig = {};

  if (configFile) {
    try {
      const content = readFileSync(configFile, 'utf-8');
      fileConfig = YAML.parse(content) || {};
    } catch (err) {
      // Ignore config parsing errors, use defaults
    }
  }

  loadedConfig = {
    ...DEFAULTS,
    ...fileConfig,
    tokens: {
      github: process.env.GITHUB_TOKEN || process.env.GH_TOKEN || fileConfig.tokens?.github,
      gitlab: process.env.GITLAB_TOKEN || fileConfig.tokens?.gitlab,
      bitbucket: process.env.BITBUCKET_TOKEN || fileConfig.tokens?.bitbucket,
    },
  };

  return loadedConfig;
}

/**
 * Get a specific config value
 */
export function getConfig(key) {
  const config = loadConfig();
  return key ? config[key] : config;
}

/**
 * Get auth token for a provider
 */
export function getToken(provider) {
  const config = loadConfig();
  return config.tokens?.[provider];
}

/**
 * Merge CLI options with config
 */
export function mergeOptions(cliOptions) {
  const config = loadConfig();

  return {
    concurrency: cliOptions.concurrency ?? config.concurrency,
    retries: cliOptions.retries ?? config.retries,
    retryDelay: cliOptions.retryDelay ?? config.retryDelay,
    timeout: cliOptions.timeout ?? config.timeout,
    branch: cliOptions.branch ?? config.branch,
    outputDir: cliOptions.output ?? config.outputDir,
    interactive: cliOptions.interactive ?? false,
    verbose: cliOptions.verbose ?? false,
    quiet: cliOptions.quiet ?? false,
    flat: cliOptions.flat ?? false,
  };
}

export default {
  loadConfig,
  getConfig,
  getToken,
  mergeOptions,
};
