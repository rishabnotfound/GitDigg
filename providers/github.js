import { BaseProvider } from './base.js';
import { API_URLS, RAW_URLS } from '../lib/constants.js';

/**
 * GitHub API provider
 */
export class GitHubProvider extends BaseProvider {
  get name() {
    return 'github';
  }

  get apiBase() {
    return API_URLS.github;
  }

  get rawBase() {
    return RAW_URLS.github;
  }

  getHeaders() {
    const headers = {
      'Accept': 'application/vnd.github+json',
      'User-Agent': 'gitdigg-cli',
      'X-GitHub-Api-Version': '2022-11-28',
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    return headers;
  }

  /**
   * Get default branch for the repository
   */
  async getDefaultBranch() {
    const url = `${this.apiBase}/repos/${this.owner}/${this.repo}`;
    const data = await this.fetchJson(url);
    return data.default_branch;
  }

  /**
   * Get repository tree recursively
   */
  async getTree(branch, basePath = '') {
    const url = `${this.apiBase}/repos/${this.owner}/${this.repo}/git/trees/${branch}?recursive=1`;

    try {
      const data = await this.fetchJson(url);

      if (data.truncated) {
        // Tree is too large, fall back to contents API
        return this.getTreeViaContents(branch, basePath);
      }

      let items = data.tree.map(item => ({
        path: item.path,
        type: item.type === 'blob' ? 'file' : 'tree',
        size: item.size || 0,
        sha: item.sha,
      }));

      // Filter by base path if provided
      if (basePath) {
        const normalizedBase = basePath.replace(/\/$/, '');
        items = items.filter(item => {
          return item.path === normalizedBase || item.path.startsWith(normalizedBase + '/');
        });
      }

      return items;
    } catch (err) {
      if (err.status === 404) {
        throw new Error(`Branch '${branch}' not found or repository is empty`);
      }
      throw err;
    }
  }

  /**
   * Fallback: get tree via contents API (for large repos)
   */
  async getTreeViaContents(branch, path = '') {
    const url = `${this.apiBase}/repos/${this.owner}/${this.repo}/contents/${path}?ref=${branch}`;
    const data = await this.fetchJson(url);

    if (!Array.isArray(data)) {
      // Single file
      return [{
        path: data.path,
        type: data.type === 'dir' ? 'tree' : 'file',
        size: data.size || 0,
        sha: data.sha,
      }];
    }

    const items = [];

    for (const item of data) {
      items.push({
        path: item.path,
        type: item.type === 'dir' ? 'tree' : 'file',
        size: item.size || 0,
        sha: item.sha,
      });

      // Recursively fetch directories
      if (item.type === 'dir') {
        const subItems = await this.getTreeViaContents(branch, item.path);
        items.push(...subItems);
      }
    }

    return items;
  }

  /**
   * Get raw download URL for a file
   */
  getDownloadUrl(branch, path) {
    return `${this.rawBase}/${this.owner}/${this.repo}/${branch}/${path}`;
  }

  /**
   * Get download URL via API (for private repos)
   */
  getApiDownloadUrl(branch, path) {
    return `${this.apiBase}/repos/${this.owner}/${this.repo}/contents/${path}?ref=${branch}`;
  }

  /**
   * Download file content
   */
  async downloadFile(branch, path) {
    // Try raw URL first (faster, no API limits)
    try {
      const url = this.getDownloadUrl(branch, path);
      const response = await this.fetch(url);
      return response.arrayBuffer();
    } catch (err) {
      // Fall back to API for private repos
      if (err.status === 404 && this.token) {
        const url = this.getApiDownloadUrl(branch, path);
        const data = await this.fetchJson(url);

        if (data.content && data.encoding === 'base64') {
          const binary = atob(data.content.replace(/\n/g, ''));
          const bytes = new Uint8Array(binary.length);
          for (let i = 0; i < binary.length; i++) {
            bytes[i] = binary.charCodeAt(i);
          }
          return bytes.buffer;
        }

        // Large files use download_url
        if (data.download_url) {
          const response = await this.fetch(data.download_url);
          return response.arrayBuffer();
        }
      }
      throw err;
    }
  }

  /**
   * Check rate limit status
   */
  async getRateLimit() {
    const url = `${this.apiBase}/rate_limit`;
    const data = await this.fetchJson(url);
    return {
      limit: data.rate.limit,
      remaining: data.rate.remaining,
      reset: new Date(data.rate.reset * 1000),
    };
  }
}

export default GitHubProvider;
