import { GitHubProvider } from './github.js';
import { GitLabProvider } from './gitlab.js';
import { parseRepoUrl } from '../utils/url.js';

const providers = {
  github: GitHubProvider,
  gitlab: GitLabProvider,
};

/**
 * Create a provider instance from a repository URL or shorthand
 */
export function createProvider(repoUrl, options = {}) {
  const parsed = parseRepoUrl(repoUrl);
  const ProviderClass = providers[parsed.provider];

  if (!ProviderClass) {
    throw new Error(`Unknown provider: ${parsed.provider}`);
  }

  // Use token from URL if present, otherwise from options
  const providerOptions = {
    ...options,
    token: parsed.token || options.token,
  };

  const provider = new ProviderClass(parsed.owner, parsed.repo, providerOptions);

  return {
    provider,
    owner: parsed.owner,
    repo: parsed.repo,
    path: parsed.path,
    branch: parsed.branch,
  };
}

/**
 * Get a provider class by name
 */
export function getProviderClass(name) {
  return providers[name];
}

/**
 * List supported provider names
 */
export function listProviders() {
  return Object.keys(providers);
}

/**
 * Detect provider from URL
 */
export function detectProvider(url) {
  const parsed = parseRepoUrl(url);
  return parsed.provider;
}

export {
  GitHubProvider,
  GitLabProvider,
};

export default {
  createProvider,
  getProviderClass,
  listProviders,
  detectProvider,
};
