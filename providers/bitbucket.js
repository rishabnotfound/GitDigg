import { BaseProvider } from './base.js';
import { API_URLS } from '../lib/constants.js';

/**
 * Bitbucket API provider
 */
export class BitbucketProvider extends BaseProvider {
  get name() {
    return 'bitbucket';
  }

  get apiBase() {
    return API_URLS.bitbucket;
  }

  getHeaders() {
    const headers = {
      'Accept': 'application/json',
      'User-Agent': 'gitdigg-cli',
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    return headers;
  }

  /**
   * Get default branch (main branch) for the repository
   */
  async getDefaultBranch() {
    const url = `${this.apiBase}/repositories/${this.owner}/${this.repo}`;
    const data = await this.fetchJson(url);
    return data.mainbranch?.name || 'main';
  }

  /**
   * Get repository tree recursively
   */
  async getTree(branch, basePath = '') {
    const items = [];
    let url = `${this.apiBase}/repositories/${this.owner}/${this.repo}/src/${encodeURIComponent(branch)}/${basePath}`;

    while (url) {
      try {
        const data = await this.fetchJson(url);

        if (data.values) {
          for (const item of data.values) {
            const entry = {
              path: item.path,
              type: item.type === 'commit_directory' ? 'tree' : 'file',
              size: item.size || 0,
            };

            items.push(entry);

            // Recursively fetch directories
            if (item.type === 'commit_directory') {
              const subItems = await this.getTree(branch, item.path);
              items.push(...subItems);
            }
          }
        }

        url = data.next || null;
      } catch (err) {
        if (err.status === 404) {
          throw new Error(`Branch '${branch}' not found or repository is empty`);
        }
        throw err;
      }
    }

    return items;
  }

  /**
   * Get raw download URL for a file
   */
  getDownloadUrl(branch, path) {
    return `${this.apiBase}/repositories/${this.owner}/${this.repo}/src/${encodeURIComponent(branch)}/${path}`;
  }

  /**
   * Download file content
   */
  async downloadFile(branch, path) {
    const url = this.getDownloadUrl(branch, path);

    // Bitbucket returns raw content when Accept header is not JSON
    const response = await fetch(url, {
      headers: {
        'User-Agent': 'gitdigg-cli',
        ...(this.token ? { 'Authorization': `Bearer ${this.token}` } : {}),
      },
    });

    if (!response.ok) {
      const error = new Error(`HTTP ${response.status}: ${response.statusText}`);
      error.status = response.status;
      throw error;
    }

    return response.arrayBuffer();
  }

  /**
   * Get file metadata
   */
  async getFileInfo(branch, path) {
    const url = `${this.apiBase}/repositories/${this.owner}/${this.repo}/src/${encodeURIComponent(branch)}/${path}?format=meta`;
    return this.fetchJson(url);
  }
}

export default BitbucketProvider;
