import ora from 'ora';
import chalk from 'chalk';

/**
 * Create a spinner instance
 */
export function createSpinner(text) {
  return ora({
    text,
    color: 'cyan',
    spinner: 'dots',
  });
}

/**
 * Create a progress bar for downloads
 */
export function createProgressBar(total) {
  let current = 0;
  let currentFile = '';
  let spinner = null;
  let startTime = null;

  const render = () => {
    const percent = Math.round((current / total) * 100);
    const filled = Math.round(percent / 5);
    const empty = 20 - filled;
    const bar = chalk.cyan('\u2588'.repeat(filled)) + chalk.gray('\u2591'.repeat(empty));

    const elapsed = startTime ? (Date.now() - startTime) / 1000 : 0;
    const rate = elapsed > 0 ? Math.round(current / elapsed) : 0;

    const status = `${bar} ${percent}% (${current}/${total})`;
    const rateText = rate > 0 ? chalk.dim(` ${rate} files/s`) : '';
    const fileText = currentFile ? chalk.dim(` ${truncate(currentFile, 30)}`) : '';

    return `${status}${rateText}${fileText}`;
  };

  return {
    start() {
      startTime = Date.now();
      spinner = ora({
        text: render(),
        color: 'cyan',
        spinner: 'dots',
      }).start();
    },

    update(completed, file = '') {
      current = completed;
      currentFile = file;
      if (spinner) {
        spinner.text = render();
      }
    },

    stop() {
      if (spinner) {
        spinner.stop();
        spinner = null;
      }
    },

    succeed(text) {
      if (spinner) {
        spinner.succeed(text);
        spinner = null;
      }
    },

    fail(text) {
      if (spinner) {
        spinner.fail(text);
        spinner = null;
      }
    },
  };
}

/**
 * Truncate a string to max length
 */
function truncate(str, maxLength) {
  if (str.length <= maxLength) return str;
  return '...' + str.slice(-(maxLength - 3));
}

/**
 * Create a simple counter display
 */
export function createCounter(label, total) {
  let current = 0;

  return {
    increment() {
      current++;
      process.stdout.write(`\r${label}: ${current}/${total}`);
    },

    done() {
      process.stdout.write('\n');
    },

    get current() {
      return current;
    },
  };
}

export default {
  createSpinner,
  createProgressBar,
  createCounter,
};
