import { BaseProvider } from './base.js';
import { API_URLS, RAW_URLS } from '../lib/constants.js';

/**
 * GitLab API provider
 */
export class GitLabProvider extends BaseProvider {
  get name() {
    return 'gitlab';
  }

  get apiBase() {
    return API_URLS.gitlab;
  }

  get rawBase() {
    return RAW_URLS.gitlab;
  }

  /**
   * Get encoded project ID (owner/repo URL encoded)
   */
  get projectId() {
    return encodeURIComponent(`${this.owner}/${this.repo}`);
  }

  getHeaders() {
    const headers = {
      'Accept': 'application/json',
      'User-Agent': 'gitdigg-cli',
    };

    if (this.token) {
      headers['PRIVATE-TOKEN'] = this.token;
    }

    return headers;
  }

  /**
   * Get default branch for the repository
   */
  async getDefaultBranch() {
    const url = `${this.apiBase}/projects/${this.projectId}`;
    const data = await this.fetchJson(url);
    return data.default_branch;
  }

  /**
   * Get repository tree recursively
   */
  async getTree(branch, basePath = '') {
    const items = [];
    let page = 1;
    const perPage = 100;

    while (true) {
      let url = `${this.apiBase}/projects/${this.projectId}/repository/tree?ref=${encodeURIComponent(branch)}&recursive=true&per_page=${perPage}&page=${page}`;

      if (basePath) {
        url += `&path=${encodeURIComponent(basePath)}`;
      }

      try {
        const data = await this.fetchJson(url);

        if (!Array.isArray(data) || data.length === 0) {
          break;
        }

        for (const item of data) {
          items.push({
            path: item.path,
            type: item.type === 'tree' ? 'tree' : 'file',
            name: item.name,
            id: item.id,
          });
        }

        if (data.length < perPage) {
          break;
        }

        page++;
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
    const encodedPath = encodeURIComponent(path);
    return `${this.apiBase}/projects/${this.projectId}/repository/files/${encodedPath}/raw?ref=${encodeURIComponent(branch)}`;
  }

  /**
   * Download file content
   */
  async downloadFile(branch, path) {
    const url = this.getDownloadUrl(branch, path);
    const response = await this.fetch(url);
    return response.arrayBuffer();
  }

  /**
   * Get file metadata
   */
  async getFileInfo(branch, path) {
    const encodedPath = encodeURIComponent(path);
    const url = `${this.apiBase}/projects/${this.projectId}/repository/files/${encodedPath}?ref=${encodeURIComponent(branch)}`;
    return this.fetchJson(url);
  }
}

export default GitLabProvider;
