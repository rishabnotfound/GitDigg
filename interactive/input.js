import { createInterface } from 'node:readline';

/**
 * Key codes for special keys
 */
export const KEYS = {
  UP: '\x1B[A',
  DOWN: '\x1B[B',
  RIGHT: '\x1B[C',
  LEFT: '\x1B[D',
  ENTER: '\r',
  SPACE: ' ',
  ESCAPE: '\x1B',
  CTRL_C: '\x03',
  CTRL_D: '\x04',
  TAB: '\t',
  BACKSPACE: '\x7F',
  DELETE: '\x1B[3~',
  HOME: '\x1B[H',
  END: '\x1B[F',
  PAGE_UP: '\x1B[5~',
  PAGE_DOWN: '\x1B[6~',
};

/**
 * Keyboard input handler using raw mode
 */
export class InputHandler {
  constructor() {
    this.listeners = new Map();
    this.rl = null;
    this.active = false;
  }

  /**
   * Start listening for keyboard input
   */
  start() {
    if (this.active) return;

    this.active = true;

    // Enable raw mode for character-by-character input
    if (process.stdin.isTTY) {
      process.stdin.setRawMode(true);
    }

    process.stdin.resume();
    process.stdin.setEncoding('utf8');

    this.rl = createInterface({
      input: process.stdin,
      output: process.stdout,
      terminal: true,
    });

    process.stdin.on('data', this.handleInput.bind(this));
  }

  /**
   * Stop listening for keyboard input
   */
  stop() {
    if (!this.active) return;

    this.active = false;

    if (process.stdin.isTTY) {
      process.stdin.setRawMode(false);
    }

    process.stdin.pause();

    if (this.rl) {
      this.rl.close();
      this.rl = null;
    }

    process.stdin.removeAllListeners('data');
  }

  /**
   * Handle raw input data
   */
  handleInput(data) {
    const key = data.toString();

    // Check for specific key bindings
    if (this.listeners.has(key)) {
      this.listeners.get(key)(key);
      return;
    }

    // Check for generic 'key' listener
    if (this.listeners.has('key')) {
      this.listeners.get('key')(key);
    }

    // Check for printable characters
    if (key.length === 1 && key.charCodeAt(0) >= 32 && key.charCodeAt(0) < 127) {
      if (this.listeners.has('char')) {
        this.listeners.get('char')(key);
      }
    }
  }

  /**
   * Register a key handler
   */
  on(key, handler) {
    this.listeners.set(key, handler);
    return this;
  }

  /**
   * Remove a key handler
   */
  off(key) {
    this.listeners.delete(key);
    return this;
  }

  /**
   * Clear all handlers
   */
  clear() {
    this.listeners.clear();
    return this;
  }
}

/**
 * Create a simple input prompt
 */
export async function prompt(message) {
  const rl = createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  return new Promise(resolve => {
    rl.question(message, answer => {
      rl.close();
      resolve(answer);
    });
  });
}

/**
 * Create a confirmation prompt
 */
export async function confirm(message, defaultValue = false) {
  const hint = defaultValue ? '[Y/n]' : '[y/N]';
  const answer = await prompt(`${message} ${hint} `);

  if (!answer) return defaultValue;

  return answer.toLowerCase().startsWith('y');
}

export default {
  KEYS,
  InputHandler,
  prompt,
  confirm,
};
