import chalk from 'chalk';
import { Title } from '../lib/constants.js';

/**
 * Terminal renderer for interactive mode
 */
export class Renderer {
  constructor() {
    this.width = process.stdout.columns || 80;
    this.height = process.stdout.rows || 24;

    // Listen for terminal resize
    process.stdout.on('resize', () => {
      this.width = process.stdout.columns || 80;
      this.height = process.stdout.rows || 24;
    });
  }

  /**
   * Clear the screen
   */
  clear() {
    process.stdout.write('\x1B[2J\x1B[H');
  }

  /**
   * Move cursor to position
   */
  moveTo(x, y) {
    process.stdout.write(`\x1B[${y};${x}H`);
  }

  /**
   * Hide cursor
   */
  hideCursor() {
    process.stdout.write('\x1B[?25l');
  }

  /**
   * Show cursor
   */
  showCursor() {
    process.stdout.write('\x1B[?25h');
  }

  /**
   * Write text at current position
   */
  write(text) {
    process.stdout.write(text);
  }

  /**
   * Write a line
   */
  writeLine(text = '') {
    process.stdout.write(text + '\n');
  }

  /**
   * Render the header
   */
  renderHeader(repo, branch, selectedCount) {
    const title = chalk.bold.cyan(Title) + chalk.dim(' - Interactive Mode');
    const repoInfo = chalk.yellow(repo) + chalk.dim('@') + chalk.green(branch);
    const selection = selectedCount > 0
      ? chalk.magenta(` [${selectedCount} selected]`)
      : '';

    this.writeLine(title);
    this.writeLine(repoInfo + selection);
    this.writeLine(chalk.dim('\u2500'.repeat(Math.min(this.width, 60))));
  }

  /**
   * Render the file tree
   */
  renderTree(browser) {
    const nodes = browser.getVisibleSlice();
    const cursorIndex = browser.cursor - browser.viewOffset;

    for (let i = 0; i < nodes.length; i++) {
      const node = nodes[i];
      const isCursor = i === cursorIndex;
      const isSelected = browser.isSelected(node);
      const isPartial = browser.isPartiallySelected(node);

      this.renderNode(node, isCursor, isSelected, isPartial, browser.expanded.has(node.path));
    }

    // Fill remaining lines
    const remaining = browser.viewHeight - nodes.length;
    for (let i = 0; i < remaining; i++) {
      this.writeLine();
    }
  }

  /**
   * Render a single tree node
   */
  renderNode(node, isCursor, isSelected, isPartial, isExpanded) {
    const indent = '  '.repeat(node.depth);

    // Selection indicator
    let checkbox;
    if (isSelected) {
      checkbox = chalk.green('[x]');
    } else if (isPartial) {
      checkbox = chalk.yellow('[-]');
    } else {
      checkbox = chalk.dim('[ ]');
    }

    // File/folder icon
    let icon;
    if (node.isDirectory) {
      icon = isExpanded ? chalk.yellow('\u25BC') : chalk.yellow('\u25B6');
    } else {
      icon = chalk.dim('\u2022');
    }

    // Name styling
    let name;
    if (node.isDirectory) {
      name = chalk.bold.blue(node.name + '/');
    } else {
      name = node.name;
    }

    // Cursor highlight
    const line = `${indent}${checkbox} ${icon} ${name}`;
    if (isCursor) {
      this.writeLine(chalk.inverse(line));
    } else {
      this.writeLine(line);
    }
  }

  /**
   * Render the footer with keybindings
   */
  renderFooter(searchMode = false, searchQuery = '') {
    this.writeLine(chalk.dim('\u2500'.repeat(Math.min(this.width, 60))));

    if (searchMode) {
      this.writeLine(chalk.cyan('Search: ') + searchQuery + chalk.dim('\u2588'));
      this.writeLine(chalk.dim('Enter: confirm | Esc: cancel'));
    } else {
      this.writeLine(
        chalk.dim('\u2191\u2193') + ' navigate  ' +
        chalk.dim('Space') + ' select  ' +
        chalk.dim('\u2192') + ' expand  ' +
        chalk.dim('/') + ' search  ' +
        chalk.dim('Enter') + ' download'
      );
      this.writeLine(
        chalk.dim('a') + ' select all  ' +
        chalk.dim('e') + ' expand all  ' +
        chalk.dim('q') + ' quit'
      );
    }
  }

  /**
   * Render the full screen
   */
  render(browser, repo, branch, searchMode = false, searchQuery = '') {
    this.clear();
    this.hideCursor();

    this.renderHeader(repo, branch, browser.selected.size);
    this.renderTree(browser);
    this.renderFooter(searchMode, searchQuery);
  }

  /**
   * Render a message screen
   */
  renderMessage(title, message) {
    this.clear();
    this.writeLine(chalk.bold(title));
    this.writeLine();
    this.writeLine(message);
  }

  /**
   * Cleanup on exit
   */
  cleanup() {
    this.showCursor();
    this.clear();
  }
}

export default Renderer;
