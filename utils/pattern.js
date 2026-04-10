import { minimatch } from 'minimatch';

/**
 * Check if a path matches any of the given patterns
 */
export function matchesAny(path, patterns) {
  if (!patterns || patterns.length === 0) {
    return true;
  }

  return patterns.some(pattern => matches(path, pattern));
}

/**
 * Check if a path matches a single pattern
 */
export function matches(path, pattern) {
  // Handle exact matches
  if (path === pattern) {
    return true;
  }

  // Handle directory prefix matches (e.g., "src" matches "src/file.js")
  if (!pattern.includes('*') && !pattern.includes('?')) {
    const normalizedPattern = pattern.replace(/\/$/, '');
    if (path === normalizedPattern || path.startsWith(normalizedPattern + '/')) {
      return true;
    }
  }

  // Use minimatch for glob patterns
  return minimatch(path, pattern, {
    matchBase: true,
    dot: true,
    nocase: false,
  });
}

/**
 * Filter a list of files by patterns
 */
export function filterByPatterns(files, patterns) {
  if (!patterns || patterns.length === 0) {
    return files;
  }

  return files.filter(file => {
    const path = typeof file === 'string' ? file : file.path;
    return matchesAny(path, patterns);
  });
}

/**
 * Check if a pattern is a glob pattern
 */
export function isGlobPattern(pattern) {
  return /[*?[\]{}!]/.test(pattern);
}

/**
 * Normalize patterns for consistent matching
 */
export function normalizePatterns(patterns) {
  if (!patterns) return [];

  return patterns.map(p => {
    // Remove leading/trailing slashes
    let normalized = p.replace(/^\/+|\/+$/g, '');

    // Convert ** at start to match any prefix
    if (normalized.startsWith('**/')) {
      normalized = normalized.slice(3);
    }

    return normalized;
  });
}

/**
 * Check if a path is within a base directory
 */
export function isWithinPath(filePath, basePath) {
  if (!basePath) return true;

  const normalizedBase = basePath.replace(/\/$/, '');
  return filePath === normalizedBase || filePath.startsWith(normalizedBase + '/');
}

/**
 * Get the relative path from a base path
 */
export function relativePath(filePath, basePath) {
  if (!basePath) return filePath;

  const normalizedBase = basePath.replace(/\/$/, '');
  if (filePath === normalizedBase) {
    return filePath.split('/').pop();
  }

  if (filePath.startsWith(normalizedBase + '/')) {
    return filePath.slice(normalizedBase.length + 1);
  }

  return filePath;
}

export default {
  matches,
  matchesAny,
  filterByPatterns,
  isGlobPattern,
  normalizePatterns,
  isWithinPath,
  relativePath,
};
