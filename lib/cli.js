import { program } from 'commander';
import { createProvider } from '../providers/index.js';
import { download } from '../downloader/index.js';
import { startInteractive } from '../interactive/index.js';
import { mergeOptions, loadConfig } from './config.js';
import { normalizePatterns } from '../utils/pattern.js';
import logger, { setLogLevel, style } from '../utils/logger.js';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

// Get package version
const __dirname = dirname(fileURLToPath(import.meta.url));
const pkg = JSON.parse(readFileSync(join(__dirname, '..', 'package.json'), 'utf-8'));

/**
 * Run the CLI
 */
export async function run(argv) {
  program
    .name('gitdigg')
    .description('Download specific files from git repositories without cloning')
    .version(pkg.version)
    .argument('<repository>', 'Repository URL or owner/repo shorthand')
    .argument('[paths...]', 'Files or directories to download (supports glob patterns)')
    .option('-b, --branch <branch>', 'Branch, tag, or commit to download from')
    .option('-o, --output <dir>', 'Output directory', '.')
    .option('-i, --interactive', 'Interactive mode - browse and select files')
    .option('-c, --concurrency <n>', 'Number of concurrent downloads', parseInt, 5)
    .option('--flat', 'Download all files to output directory without preserving structure')
    .option('-v, --verbose', 'Verbose output')
    .option('-q, --quiet', 'Suppress output')
    .option('--retries <n>', 'Number of retries for failed downloads', parseInt, 3)
    .action(async (repository, paths, opts) => {
      try {
        await execute(repository, paths, opts);
      } catch (err) {
        if (opts.verbose) {
          console.error(err);
        }
        logger.fatal(err.message);
      }
    });

  await program.parseAsync(argv);
}

/**
 * Execute the download command
 */
async function execute(repository, paths, opts) {
  // Load config and merge with CLI options
  loadConfig();
  const options = mergeOptions(opts);

  // Set log level
  if (options.verbose) {
    setLogLevel('debug');
  } else if (options.quiet) {
    setLogLevel('silent');
  }

  // Create provider from repository URL
  const { provider, owner, repo, path: urlPath, branch: urlBranch } = createProvider(repository);

  // Determine branch
  const branch = options.branch || urlBranch || await provider.getDefaultBranch();

  // Log info
  if (!options.quiet) {
    logger.info(`Repository: ${style.provider(provider.name)} ${style.bold(`${owner}/${repo}`)}`);
    logger.info(`Branch: ${style.branch(branch)}`);
  }

  // Interactive mode
  if (options.interactive) {
    const result = await startInteractive(provider, owner, repo, branch, {
      outputDir: options.outputDir,
      concurrency: options.concurrency,
      flat: options.flat,
    });

    if (result.cancelled) {
      logger.info('Cancelled');
      return;
    }

    return;
  }

  // Build download paths from arguments and URL path
  let downloadPaths = [...paths];
  if (urlPath && !downloadPaths.includes(urlPath)) {
    downloadPaths.unshift(urlPath);
  }

  // Separate patterns (containing globs) from paths
  const patterns = [];
  const exactPaths = [];

  for (const p of downloadPaths) {
    if (/[*?[\]{}]/.test(p)) {
      patterns.push(p);
    } else {
      exactPaths.push(p);
    }
  }

  // Normalize patterns
  const normalizedPatterns = normalizePatterns(patterns);

  // If no paths specified and not interactive, download everything
  if (downloadPaths.length === 0) {
    logger.info('No paths specified, downloading entire repository');
  }

  // Execute download
  const result = await download(provider, {
    branch,
    paths: exactPaths,
    patterns: normalizedPatterns,
    outputDir: options.outputDir,
    concurrency: options.concurrency,
    flat: options.flat,
    verbose: options.verbose,
    quiet: options.quiet,
    retries: options.retries,
  });

  // Exit with error if any downloads failed
  if (result.failed > 0) {
    process.exitCode = 1;
  }
}

export default { run };
