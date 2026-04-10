/**
 * File tree browser state management
 */
export class FileBrowser {
  constructor(tree) {
    this.tree = tree;
    this.cursor = 0;
    this.selected = new Set();
    this.expanded = new Set();
    this.searchQuery = '';
    this.viewOffset = 0;
    this.viewHeight = 20;

    // Build tree structure
    this.buildTree();
  }

  /**
   * Build hierarchical tree from flat list
   */
  buildTree() {
    this.nodes = [];
    this.nodeMap = new Map();

    // Sort items by path
    const sortedItems = [...this.tree].sort((a, b) => a.path.localeCompare(b.path));

    // Create nodes with depth info
    for (const item of sortedItems) {
      const parts = item.path.split('/');
      const depth = parts.length - 1;
      const name = parts[parts.length - 1];

      const node = {
        ...item,
        name,
        depth,
        isDirectory: item.type === 'tree',
      };

      this.nodes.push(node);
      this.nodeMap.set(item.path, node);
    }

    // Filter to visible nodes
    this.updateVisibleNodes();
  }

  /**
   * Update the list of visible nodes based on expanded state
   */
  updateVisibleNodes() {
    this.visibleNodes = [];

    for (const node of this.nodes) {
      // Check if all parent directories are expanded
      const parts = node.path.split('/');
      let visible = true;

      for (let i = 1; i < parts.length; i++) {
        const parentPath = parts.slice(0, i).join('/');
        if (!this.expanded.has(parentPath)) {
          visible = false;
          break;
        }
      }

      // Apply search filter
      if (visible && this.searchQuery) {
        const query = this.searchQuery.toLowerCase();
        visible = node.path.toLowerCase().includes(query) ||
                  node.name.toLowerCase().includes(query);
      }

      if (visible) {
        this.visibleNodes.push(node);
      }
    }

    // Ensure cursor is within bounds
    if (this.cursor >= this.visibleNodes.length) {
      this.cursor = Math.max(0, this.visibleNodes.length - 1);
    }

    this.updateViewOffset();
  }

  /**
   * Update view offset to keep cursor visible
   */
  updateViewOffset() {
    if (this.cursor < this.viewOffset) {
      this.viewOffset = this.cursor;
    } else if (this.cursor >= this.viewOffset + this.viewHeight) {
      this.viewOffset = this.cursor - this.viewHeight + 1;
    }
  }

  /**
   * Move cursor up
   */
  moveUp() {
    if (this.cursor > 0) {
      this.cursor--;
      this.updateViewOffset();
    }
  }

  /**
   * Move cursor down
   */
  moveDown() {
    if (this.cursor < this.visibleNodes.length - 1) {
      this.cursor++;
      this.updateViewOffset();
    }
  }

  /**
   * Page up
   */
  pageUp() {
    this.cursor = Math.max(0, this.cursor - this.viewHeight);
    this.updateViewOffset();
  }

  /**
   * Page down
   */
  pageDown() {
    this.cursor = Math.min(this.visibleNodes.length - 1, this.cursor + this.viewHeight);
    this.updateViewOffset();
  }

  /**
   * Go to top
   */
  goToTop() {
    this.cursor = 0;
    this.updateViewOffset();
  }

  /**
   * Go to bottom
   */
  goToBottom() {
    this.cursor = this.visibleNodes.length - 1;
    this.updateViewOffset();
  }

  /**
   * Toggle selection of current item
   */
  toggleSelect() {
    const node = this.visibleNodes[this.cursor];
    if (!node) return;

    if (node.isDirectory) {
      // Toggle all files under this directory
      const prefix = node.path + '/';
      const filesUnder = this.nodes.filter(n => !n.isDirectory && n.path.startsWith(prefix));
      const allSelected = filesUnder.every(f => this.selected.has(f.path));

      for (const file of filesUnder) {
        if (allSelected) {
          this.selected.delete(file.path);
        } else {
          this.selected.add(file.path);
        }
      }
    } else {
      // Toggle single file
      if (this.selected.has(node.path)) {
        this.selected.delete(node.path);
      } else {
        this.selected.add(node.path);
      }
    }
  }

  /**
   * Toggle expand/collapse of directory
   */
  toggleExpand() {
    const node = this.visibleNodes[this.cursor];
    if (!node || !node.isDirectory) return;

    if (this.expanded.has(node.path)) {
      this.expanded.delete(node.path);
    } else {
      this.expanded.add(node.path);
    }

    this.updateVisibleNodes();
  }

  /**
   * Expand all directories
   */
  expandAll() {
    for (const node of this.nodes) {
      if (node.isDirectory) {
        this.expanded.add(node.path);
      }
    }
    this.updateVisibleNodes();
  }

  /**
   * Collapse all directories
   */
  collapseAll() {
    this.expanded.clear();
    this.updateVisibleNodes();
  }

  /**
   * Set search query
   */
  setSearch(query) {
    this.searchQuery = query;
    this.cursor = 0;
    this.updateVisibleNodes();
  }

  /**
   * Clear search
   */
  clearSearch() {
    this.searchQuery = '';
    this.updateVisibleNodes();
  }

  /**
   * Select all visible files
   */
  selectAll() {
    for (const node of this.visibleNodes) {
      if (!node.isDirectory) {
        this.selected.add(node.path);
      }
    }
  }

  /**
   * Deselect all
   */
  deselectAll() {
    this.selected.clear();
  }

  /**
   * Get current node
   */
  getCurrentNode() {
    return this.visibleNodes[this.cursor];
  }

  /**
   * Get selected files
   */
  getSelectedFiles() {
    return Array.from(this.selected);
  }

  /**
   * Get visible nodes for rendering
   */
  getVisibleSlice() {
    return this.visibleNodes.slice(this.viewOffset, this.viewOffset + this.viewHeight);
  }

  /**
   * Check if a node is selected
   */
  isSelected(node) {
    if (node.isDirectory) {
      const prefix = node.path + '/';
      const filesUnder = this.nodes.filter(n => !n.isDirectory && n.path.startsWith(prefix));
      return filesUnder.length > 0 && filesUnder.every(f => this.selected.has(f.path));
    }
    return this.selected.has(node.path);
  }

  /**
   * Check if a node is partially selected (some children selected)
   */
  isPartiallySelected(node) {
    if (!node.isDirectory) return false;

    const prefix = node.path + '/';
    const filesUnder = this.nodes.filter(n => !n.isDirectory && n.path.startsWith(prefix));
    const selectedCount = filesUnder.filter(f => this.selected.has(f.path)).length;

    return selectedCount > 0 && selectedCount < filesUnder.length;
  }

  /**
   * Set view height
   */
  setViewHeight(height) {
    this.viewHeight = height;
    this.updateViewOffset();
  }
}

export default FileBrowser;
