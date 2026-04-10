import chalk from 'chalk';

const LOG_LEVELS = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3,
  silent: 4,
};

let currentLevel = LOG_LEVELS.info;

export function setLogLevel(level) {
  if (LOG_LEVELS[level] !== undefined) {
    currentLevel = LOG_LEVELS[level];
  }
}

export function debug(...args) {
  if (currentLevel <= LOG_LEVELS.debug) {
    console.log(chalk.gray('[debug]'), ...args);
  }
}

export function info(...args) {
  if (currentLevel <= LOG_LEVELS.info) {
    console.log(chalk.blue('i'), ...args);
  }
}

export function success(...args) {
  if (currentLevel <= LOG_LEVELS.info) {
    console.log(chalk.green('\u2713'), ...args);
  }
}

export function warn(...args) {
  if (currentLevel <= LOG_LEVELS.warn) {
    console.log(chalk.yellow('\u26A0'), ...args);
  }
}

export function error(...args) {
  if (currentLevel <= LOG_LEVELS.error) {
    console.error(chalk.red('\u2717'), ...args);
  }
}

export function fatal(message, exitCode = 1) {
  error(message);
  process.exit(exitCode);
}

// Styled output helpers
export const style = {
  bold: chalk.bold,
  dim: chalk.dim,
  url: chalk.cyan.underline,
  path: chalk.yellow,
  number: chalk.magenta,
  provider: chalk.blue,
  branch: chalk.green,
};

export default {
  setLogLevel,
  debug,
  info,
  success,
  warn,
  error,
  fatal,
  style,
};
