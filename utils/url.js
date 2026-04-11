import { PROVIDER_HOSTS } from '../lib/constants.js';

/**
 * Parse a repository URL or shorthand into components
 * Supports:
 * - github.com/owner/repo
 * - https://github.com/owner/repo
 * - https://user:token@github.com/owner/repo (with credentials)
 * - owner/repo (defaults to GitHub)
 * - gitlab.com/owner/repo
 */
export function parseRepoUrl(input) {
  // Normalize input
  let url = input.trim();

  // Handle shorthand format: owner/repo
  if (/^[a-zA-Z0-9_.-]+\/[a-zA-Z0-9_.-]+$/.test(url)) {
    return {
      provider: 'github',
      owner: url.split('/')[0],
      repo: url.split('/')[1],
      path: '',
      branch: null,
      token: null,
    };
  }

  // Add protocol if missing
  if (!url.startsWith('http://') && !url.startsWith('https://')) {
    url = 'https://' + url;
  }

  try {
    const parsed = new URL(url);
    const hostname = parsed.hostname.toLowerCase();
    const provider = PROVIDER_HOSTS[hostname];

    if (!provider) {
      throw new Error(`Unsupported provider: ${hostname}`);
    }

    // Extract credentials if present (user:token@host or just token@host)
    let token = null;
    if (parsed.password) {
      token = parsed.password;
    } else if (parsed.username && !parsed.password) {
      // Some formats use username as token
      token = parsed.username;
    }

    // Parse path: /owner/repo[/tree/branch/path] or /owner/repo[/blob/branch/path]
    const pathParts = parsed.pathname.split('/').filter(Boolean);

    if (pathParts.length < 2) {
      throw new Error('Invalid repository URL: missing owner or repo');
    }

    const result = {
      provider,
      owner: pathParts[0],
      repo: pathParts[1].replace(/\.git$/, ''),
      path: '',
      branch: null,
      token,
    };

    // Check for tree/blob path format
    if (pathParts.length > 3 && (pathParts[2] === 'tree' || pathParts[2] === 'blob')) {
      result.branch = pathParts[3];
      result.path = pathParts.slice(4).join('/');
    } else if (pathParts.length > 2) {
      // Direct path without tree/blob
      result.path = pathParts.slice(2).join('/');
    }

    return result;
  } catch (err) {
    if (err.message.includes('Unsupported') || err.message.includes('Invalid')) {
      throw err;
    }
    throw new Error(`Invalid URL format: ${input}`);
  }
}

/**
 * Build a repository identifier string
 */
export function repoId(owner, repo) {
  return `${owner}/${repo}`;
}

/**
 * Validate and normalize a path
 */
export function normalizePath(path) {
  if (!path) return '';

  return path
    .replace(/^\/+/, '')  // Remove leading slashes
    .replace(/\/+$/, '')  // Remove trailing slashes
    .replace(/\/+/g, '/'); // Collapse multiple slashes
}

/**
 * Join path segments
 */
export function joinPath(...segments) {
  return segments
    .filter(Boolean)
    .map(s => s.replace(/^\/+|\/+$/g, ''))
    .filter(Boolean)
    .join('/');
}

/**
 * Get the base name of a path
 */
export function basename(path) {
  if (!path) return '';
  const parts = path.split('/').filter(Boolean);
  return parts[parts.length - 1] || '';
}

/**
 * Get the directory name of a path
 */
export function dirname(path) {
  if (!path) return '';
  const parts = path.split('/').filter(Boolean);
  parts.pop();
  return parts.join('/');
}

export default {
  parseRepoUrl,
  repoId,
  normalizePath,
  joinPath,
  basename,
  dirname,
};
