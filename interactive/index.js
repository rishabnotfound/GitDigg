import { FileBrowser } from './browser.js';
import { Renderer } from './renderer.js';
import { InputHandler, KEYS } from './input.js';
import { download } from '../downloader/index.js';
import { createSpinner } from '../utils/spinner.js';
import logger from '../utils/logger.js';

/**
 * Start interactive mode
 */
export async function startInteractive(provider, owner, repo, branch, options = {}) {
  const renderer = new Renderer();
  const input = new InputHandler();

  // Calculate view height (leave room for header/footer - logo is ~10 lines, footer is ~3)
  const viewHeight = Math.max(10, renderer.height - 16);

  // Fetch repository tree
  renderer.clear();
  const spinner = createSpinner('Fetching repository tree...');
  spinner.start();

  let tree;
  try {
    tree = await provider.getTree(branch);
    spinner.succeed(`Found ${tree.length} items`);
  } catch (err) {
    spinner.fail(`Failed: ${err.message}`);
    return;
  }

  // Create browser
  const browser = new FileBrowser(tree);
  browser.setViewHeight(viewHeight);

  // State
  let searchMode = false;
  let searchQuery = '';
  let running = true;

  // Initial render
  const repoId = `${owner}/${repo}`;
  renderer.render(browser, repoId, branch, searchMode, searchQuery);

  // Handle keyboard input
  return new Promise(resolve => {
    const cleanup = () => {
      input.stop();
      renderer.cleanup();
      running = false;
    };

    const rerender = () => {
      if (running) {
        renderer.render(browser, repoId, branch, searchMode, searchQuery);
      }
    };

    input.on(KEYS.CTRL_C, () => {
      cleanup();
      resolve({ cancelled: true, files: [] });
    });

    input.on('q', () => {
      if (!searchMode) {
        cleanup();
        resolve({ cancelled: true, files: [] });
      }
    });

    input.on(KEYS.ESCAPE, () => {
      if (searchMode) {
        searchMode = false;
        searchQuery = '';
        browser.clearSearch();
        rerender();
      } else {
        cleanup();
        resolve({ cancelled: true, files: [] });
      }
    });

    input.on(KEYS.UP, () => {
      if (!searchMode) {
        browser.moveUp();
        rerender();
      }
    });

    input.on(KEYS.DOWN, () => {
      if (!searchMode) {
        browser.moveDown();
        rerender();
      }
    });

    input.on(KEYS.PAGE_UP, () => {
      if (!searchMode) {
        browser.pageUp();
        rerender();
      }
    });

    input.on(KEYS.PAGE_DOWN, () => {
      if (!searchMode) {
        browser.pageDown();
        rerender();
      }
    });

    input.on(KEYS.HOME, () => {
      if (!searchMode) {
        browser.goToTop();
        rerender();
      }
    });

    input.on(KEYS.END, () => {
      if (!searchMode) {
        browser.goToBottom();
        rerender();
      }
    });

    input.on(KEYS.SPACE, () => {
      if (!searchMode) {
        browser.toggleSelect();
        browser.moveDown();
        rerender();
      }
    });

    input.on(KEYS.RIGHT, () => {
      if (!searchMode) {
        browser.toggleExpand();
        rerender();
      }
    });

    input.on(KEYS.LEFT, () => {
      if (!searchMode) {
        const node = browser.getCurrentNode();
        if (node && node.isDirectory && browser.expanded.has(node.path)) {
          browser.toggleExpand();
          rerender();
        }
      }
    });

    input.on(KEYS.ENTER, async () => {
      if (searchMode) {
        searchMode = false;
        browser.setSearch(searchQuery);
        rerender();
      } else {
        // Download selected files
        const files = browser.getSelectedFiles();

        if (files.length === 0) {
          // If nothing selected, select current item
          const current = browser.getCurrentNode();
          if (current && !current.isDirectory) {
            files.push(current.path);
          }
        }

        if (files.length === 0) {
          return;
        }

        cleanup();

        // Perform download
        logger.info(`Downloading ${files.length} files...`);

        const result = await download(provider, {
          branch,
          paths: files,
          outputDir: options.outputDir || '.',
          concurrency: options.concurrency,
          flat: options.flat,
        });

        resolve({ cancelled: false, files, result });
      }
    });

    input.on('/', () => {
      if (!searchMode) {
        searchMode = true;
        searchQuery = '';
        rerender();
      }
    });

    input.on(KEYS.BACKSPACE, () => {
      if (searchMode && searchQuery.length > 0) {
        searchQuery = searchQuery.slice(0, -1);
        rerender();
      }
    });

    input.on('a', () => {
      if (!searchMode) {
        if (browser.selected.size > 0) {
          browser.deselectAll();
        } else {
          browser.selectAll();
        }
        rerender();
      }
    });

    input.on('e', () => {
      if (!searchMode) {
        if (browser.expanded.size > 0) {
          browser.collapseAll();
        } else {
          browser.expandAll();
        }
        rerender();
      }
    });

    input.on('char', (char) => {
      if (searchMode) {
        searchQuery += char;
        rerender();
      }
    });

    input.start();
  });
}

export default {
  startInteractive,
};
