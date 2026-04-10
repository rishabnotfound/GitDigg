import { getToken } from '../lib/config.js';
import { DEFAULTS } from '../lib/constants.js';

/**
 * Base provider class with common functionality
 */
export class BaseProvider {
  constructor(owner, repo, options = {}) {
    this.owner = owner;
    this.repo = repo;
    this.options = options;
    this.token = options.token || getToken(this.name);
    this.timeout = options.timeout || DEFAULTS.timeout;
  }

  /**
   * Provider name (override in subclass)
   */
  get name() {
    throw new Error('Provider must implement name getter');
  }

  /**
   * Get default headers for API requests
   */
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
   * Make an authenticated API request
   */
  async fetch(url, options = {}) {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          ...this.getHeaders(),
          ...options.headers,
        },
        signal: controller.signal,
      });

      if (!response.ok) {
        const error = new Error(`HTTP ${response.status}: ${response.statusText}`);
        error.status = response.status;
        error.url = url;
        throw error;
      }

      return response;
    } finally {
      clearTimeout(timeoutId);
    }
  }

  /**
   * Make an API request and parse JSON
   */
  async fetchJson(url, options = {}) {
    const response = await this.fetch(url, options);
    return response.json();
  }

  /**
   * Get the default branch for the repository
   */
  async getDefaultBranch() {
    throw new Error('Provider must implement getDefaultBranch');
  }

  /**
   * Get repository tree (file listing)
   * Returns: Array of { path, type: 'file'|'tree', size? }
   */
  async getTree(branch, path = '') {
    throw new Error('Provider must implement getTree');
  }

  /**
   * Get the raw download URL for a file
   */
  getDownloadUrl(branch, path) {
    throw new Error('Provider must implement getDownloadUrl');
  }

  /**
   * Download a file's content
   */
  async downloadFile(branch, path) {
    const url = this.getDownloadUrl(branch, path);
    const response = await this.fetch(url);
    return response.arrayBuffer();
  }

  /**
   * Check if path is a directory in the tree
   */
  async isDirectory(branch, path) {
    const tree = await this.getTree(branch, path);
    return tree.some(item => item.path === path && item.type === 'tree');
  }

  /**
   * List files matching a pattern within a path
   */
  async listFiles(branch, basePath = '') {
    const tree = await this.getTree(branch, basePath);
    return tree.filter(item => item.type === 'blob' || item.type === 'file');
  }
}

export default BaseProvider;
