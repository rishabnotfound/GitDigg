import { DEFAULTS } from '../lib/constants.js';

/**
 * Async queue for concurrent downloads with limited concurrency
 */
export class DownloadQueue {
  constructor(options = {}) {
    this.concurrency = options.concurrency ?? DEFAULTS.concurrency;
    this.onProgress = options.onProgress || (() => {});
    this.onFileComplete = options.onFileComplete || (() => {});
    this.onFileError = options.onFileError || (() => {});

    this.queue = [];
    this.running = 0;
    this.completed = 0;
    this.failed = 0;
    this.results = [];
  }

  /**
   * Add a task to the queue
   */
  add(task) {
    this.queue.push(task);
  }

  /**
   * Add multiple tasks to the queue
   */
  addAll(tasks) {
    this.queue.push(...tasks);
  }

  /**
   * Process all tasks in the queue
   */
  async run() {
    const total = this.queue.length;
    const executing = new Set();

    while (this.queue.length > 0 || executing.size > 0) {
      // Start new tasks up to concurrency limit
      while (this.queue.length > 0 && executing.size < this.concurrency) {
        const task = this.queue.shift();
        const promise = this.execute(task, total).then(result => {
          executing.delete(promise);
          return result;
        });
        executing.add(promise);
      }

      // Wait for at least one task to complete
      if (executing.size > 0) {
        await Promise.race(executing);
      }
    }

    return {
      total,
      completed: this.completed,
      failed: this.failed,
      results: this.results,
    };
  }

  /**
   * Execute a single task
   */
  async execute(task, total) {
    this.running++;

    try {
      const result = await task();

      if (result.success) {
        this.completed++;
        this.onFileComplete(result, this.completed, total);
      } else {
        this.failed++;
        this.onFileError(result, this.failed, total);
      }

      this.results.push(result);
      this.onProgress(this.completed + this.failed, total, result);

      return result;
    } catch (err) {
      this.failed++;
      const result = {
        success: false,
        error: err.message,
      };
      this.results.push(result);
      this.onFileError(result, this.failed, total);
      return result;
    } finally {
      this.running--;
    }
  }

  /**
   * Get current progress
   */
  getProgress() {
    const total = this.queue.length + this.completed + this.failed + this.running;
    return {
      total,
      completed: this.completed,
      failed: this.failed,
      running: this.running,
      pending: this.queue.length,
      percent: total > 0 ? Math.round(((this.completed + this.failed) / total) * 100) : 0,
    };
  }
}

/**
 * Simple concurrent map with limited concurrency
 */
export async function concurrentMap(items, fn, concurrency = DEFAULTS.concurrency) {
  const results = [];
  const executing = new Set();

  for (const item of items) {
    const promise = fn(item).then(result => {
      executing.delete(promise);
      return result;
    });

    executing.add(promise);
    results.push(promise);

    if (executing.size >= concurrency) {
      await Promise.race(executing);
    }
  }

  return Promise.all(results);
}

export default {
  DownloadQueue,
  concurrentMap,
};
