import { writeFile, mkdir } from 'node:fs/promises';
import { dirname, join } from 'node:path';
import { DEFAULTS } from '../lib/constants.js';

/**
 * Download a single file with retry support
 */
export async function downloadFile(provider, branch, filePath, outputDir, options = {}) {
  const retries = options.retries ?? DEFAULTS.retries;
  const retryDelay = options.retryDelay ?? DEFAULTS.retryDelay;
  const flat = options.flat ?? false;

  let lastError;

  for (let attempt = 1; attempt <= retries; attempt++) {
    try {
      // Download file content
      const content = await provider.downloadFile(branch, filePath);

      // Determine output path
      const outputPath = flat
        ? join(outputDir, filePath.split('/').pop())
        : join(outputDir, filePath);

      // Ensure directory exists
      await mkdir(dirname(outputPath), { recursive: true });

      // Write file
      await writeFile(outputPath, Buffer.from(content));

      return {
        success: true,
        path: filePath,
        outputPath,
        size: content.byteLength,
        attempt,
      };
    } catch (err) {
      lastError = err;

      // Don't retry on 404 (file not found) or 403 (forbidden)
      if (err.status === 404 || err.status === 403) {
        break;
      }

      // Wait before retry
      if (attempt < retries) {
        await sleep(retryDelay * attempt);
      }
    }
  }

  return {
    success: false,
    path: filePath,
    error: lastError?.message || 'Unknown error',
    status: lastError?.status,
  };
}

/**
 * Sleep for a given number of milliseconds
 */
function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Format file size for display
 */
export function formatSize(bytes) {
  if (bytes === 0) return '0 B';

  const units = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  const size = bytes / Math.pow(1024, i);

  return `${size.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

export default {
  downloadFile,
  formatSize,
};
