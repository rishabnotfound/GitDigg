import { mkdir } from 'node:fs/promises';
import { resolve } from 'node:path';
import { DownloadQueue } from './queue.js';
import { downloadFile, formatSize } from './file.js';
import { filterByPatterns, isWithinPath } from '../utils/pattern.js';
import { createSpinner, createProgressBar } from '../utils/spinner.js';
import logger from '../utils/logger.js';

/**
 * Main download orchestrator
 */
export async function download(provider, options = {}) {
  const {
    branch,
    paths = [],
    patterns = [],
    outputDir = '.',
    concurrency = 5,
    flat = false,
    verbose = false,
    quiet = false,
  } = options;

  // Resolve output directory
  const outDir = resolve(outputDir);
  await mkdir(outDir, { recursive: true });

  // Get the branch to use
  const targetBranch = branch || await getBranch(provider);

  if (verbose) {
    logger.info(`Using branch: ${logger.style.branch(targetBranch)}`);
  }

  // Fetch repository tree
  const spinner = quiet ? null : createSpinner('Fetching repository tree...');
  spinner?.start();

  let tree;
  try {
    tree = await provider.getTree(targetBranch);
    spinner?.succeed(`Found ${tree.length} items in repository`);
  } catch (err) {
    spinner?.fail(`Failed to fetch tree: ${err.message}`);
    throw err;
  }

  // Filter to only files
  let files = tree.filter(item => item.type === 'file' || item.type === 'blob');

  // Filter by paths if specified
  if (paths.length > 0) {
    files = files.filter(file => {
      return paths.some(p => isWithinPath(file.path, p) || file.path === p);
    });
  }

  // Filter by patterns if specified
  if (patterns.length > 0) {
    files = filterByPatterns(files, patterns);
  }

  if (files.length === 0) {
    if (!quiet) {
      logger.warn('No files match the specified criteria');
    }
    return { total: 0, completed: 0, failed: 0, results: [] };
  }

  if (!quiet) {
    logger.info(`Downloading ${logger.style.number(files.length)} files to ${logger.style.path(outDir)}`);
  }

  // Create download queue
  const progressBar = quiet ? null : createProgressBar(files.length);
  progressBar?.start();

  const queue = new DownloadQueue({
    concurrency,
    onProgress: (completed, total, result) => {
      progressBar?.update(completed, result.path);
    },
    onFileComplete: (result) => {
      if (verbose && !quiet) {
        logger.success(`${result.path} (${formatSize(result.size)})`);
      }
    },
    onFileError: (result) => {
      if (!quiet) {
        logger.error(`${result.path}: ${result.error}`);
      }
    },
  });

  // Add download tasks
  for (const file of files) {
    queue.add(() => downloadFile(provider, targetBranch, file.path, outDir, {
      flat,
      retries: options.retries,
      retryDelay: options.retryDelay,
    }));
  }

  // Run downloads
  const results = await queue.run();

  progressBar?.stop();

  // Summary
  if (!quiet) {
    const totalSize = results.results
      .filter(r => r.success)
      .reduce((sum, r) => sum + (r.size || 0), 0);

    if (results.failed > 0) {
      logger.warn(`Completed: ${results.completed}/${results.total} files (${results.failed} failed)`);
    } else {
      logger.success(`Downloaded ${results.completed} files (${formatSize(totalSize)})`);
    }
  }

  return results;
}

/**
 * Get branch to use, with fallback detection
 */
async function getBranch(provider) {
  try {
    return await provider.getDefaultBranch();
  } catch (err) {
    // Try common branch names
    for (const branch of ['main', 'master']) {
      try {
        await provider.getTree(branch);
        return branch;
      } catch {
        continue;
      }
    }
    throw new Error('Could not determine default branch');
  }
}

/**
 * Download a single file or directory
 */
export async function downloadPath(provider, branch, path, outputDir, options = {}) {
  const tree = await provider.getTree(branch, path);

  // Check if path is a single file
  const exactMatch = tree.find(item => item.path === path && item.type !== 'tree');
  if (exactMatch) {
    return downloadFile(provider, branch, path, outputDir, options);
  }

  // Download all files under the path
  const files = tree.filter(item => item.type === 'file' || item.type === 'blob');

  return download(provider, {
    branch,
    paths: [path],
    outputDir,
    ...options,
  });
}

export default {
  download,
  downloadPath,
};
