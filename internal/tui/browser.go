package tui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rishabnotfound/gitdigg/internal/provider"
)

type BrowserItem struct {
	Name     string
	Path     string
	Type     provider.EntryType
	Size     int64
	Children []*BrowserItem
	Parent   *BrowserItem
	Expanded bool
	Depth    int
}

type Browser struct {
	Root         *BrowserItem
	Items        []*BrowserItem
	Cursor       int
	Selected     map[string]bool
	SearchFilter string
}

func NewBrowser(entries []provider.TreeEntry) *Browser {
	b := &Browser{Selected: make(map[string]bool)}
	b.Root = buildTree(entries)
	b.Root.Expanded = true
	b.updateVisible()
	return b
}

func buildTree(entries []provider.TreeEntry) *BrowserItem {
	root := &BrowserItem{Type: provider.EntryTypeDir, Children: make([]*BrowserItem, 0)}
	nodeMap := make(map[string]*BrowserItem)
	nodeMap[""] = root

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	for _, entry := range entries {
		parts := strings.Split(entry.Path, "/")
		currentPath := ""

		for i, part := range parts {
			parentPath := currentPath
			if currentPath == "" {
				currentPath = part
			} else {
				currentPath = currentPath + "/" + part
			}

			if _, exists := nodeMap[currentPath]; !exists {
				isLast := i == len(parts)-1
				entryType := provider.EntryTypeDir
				var size int64
				if isLast {
					entryType = entry.Type
					size = entry.Size
				}

				node := &BrowserItem{
					Name:     part,
					Path:     currentPath,
					Type:     entryType,
					Size:     size,
					Children: make([]*BrowserItem, 0),
					Depth:    i,
				}

				nodeMap[currentPath] = node

				if parent, ok := nodeMap[parentPath]; ok {
					node.Parent = parent
					parent.Children = append(parent.Children, node)
				}
			}
		}
	}

	sortChildren(root)
	return root
}

func sortChildren(item *BrowserItem) {
	sort.Slice(item.Children, func(i, j int) bool {
		if item.Children[i].Type != item.Children[j].Type {
			return item.Children[i].Type == provider.EntryTypeDir
		}
		return strings.ToLower(item.Children[i].Name) < strings.ToLower(item.Children[j].Name)
	})

	for _, child := range item.Children {
		sortChildren(child)
	}
}

func (b *Browser) updateVisible() {
	b.Items = make([]*BrowserItem, 0)
	b.flatten(b.Root)
}

func (b *Browser) flatten(node *BrowserItem) {
	for _, child := range node.Children {
		if b.SearchFilter != "" && !b.matchesFilter(child) {
			continue
		}
		b.Items = append(b.Items, child)
		if child.Type == provider.EntryTypeDir && child.Expanded {
			b.flatten(child)
		}
	}
}

func (b *Browser) matchesFilter(item *BrowserItem) bool {
	filter := strings.ToLower(b.SearchFilter)
	if strings.Contains(strings.ToLower(item.Name), filter) || strings.Contains(strings.ToLower(item.Path), filter) {
		return true
	}
	if item.Type == provider.EntryTypeDir {
		for _, child := range item.Children {
			if b.matchesFilter(child) {
				return true
			}
		}
	}
	return false
}

func (b *Browser) MoveUp() {
	if b.Cursor > 0 {
		b.Cursor--
	}
}

func (b *Browser) MoveDown() {
	if b.Cursor < len(b.Items)-1 {
		b.Cursor++
	}
}

func (b *Browser) CurrentItem() *BrowserItem {
	if len(b.Items) == 0 || b.Cursor >= len(b.Items) {
		return nil
	}
	return b.Items[b.Cursor]
}

func (b *Browser) Toggle() {
	item := b.CurrentItem()
	if item == nil || item.Type != provider.EntryTypeDir {
		return
	}
	item.Expanded = !item.Expanded
	b.updateVisible()
}

func (b *Browser) Enter() {
	item := b.CurrentItem()
	if item != nil && item.Type == provider.EntryTypeDir && !item.Expanded {
		item.Expanded = true
		b.updateVisible()
	}
}

func (b *Browser) Back() {
	item := b.CurrentItem()
	if item == nil {
		return
	}

	if item.Type == provider.EntryTypeDir && item.Expanded {
		item.Expanded = false
		b.updateVisible()
		return
	}

	if item.Parent != nil && item.Parent.Path != "" {
		for i, it := range b.Items {
			if it.Path == item.Parent.Path {
				b.Cursor = i
				return
			}
		}
	}
}

func (b *Browser) ToggleSelect() {
	item := b.CurrentItem()
	if item == nil {
		return
	}
	if b.Selected[item.Path] {
		b.deselect(item)
	} else {
		b.selectItem(item)
	}
}

func (b *Browser) selectItem(item *BrowserItem) {
	b.Selected[item.Path] = true
	if item.Type == provider.EntryTypeDir {
		for _, child := range item.Children {
			b.selectItem(child)
		}
	}
}

func (b *Browser) deselect(item *BrowserItem) {
	delete(b.Selected, item.Path)
	if item.Type == provider.EntryTypeDir {
		for _, child := range item.Children {
			b.deselect(child)
		}
	}
}

func (b *Browser) SelectAll() {
	for _, item := range b.Items {
		b.selectItem(item)
	}
}

func (b *Browser) DeselectAll() {
	b.Selected = make(map[string]bool)
}

func (b *Browser) SetFilter(filter string) {
	b.SearchFilter = filter
	b.Cursor = 0
	b.updateVisible()
}

func (b *Browser) GetSelectedFiles() []string {
	var files []string
	seen := make(map[string]bool)

	var collect func(*BrowserItem)
	collect = func(item *BrowserItem) {
		if item.Type == provider.EntryTypeFile && b.Selected[item.Path] && !seen[item.Path] {
			files = append(files, item.Path)
			seen[item.Path] = true
		}
		for _, child := range item.Children {
			collect(child)
		}
	}
	collect(b.Root)
	return files
}

func (b *Browser) GetSelectedCount() int {
	count := 0
	for path := range b.Selected {
		if item := b.findItem(path); item != nil && item.Type == provider.EntryTypeFile {
			count++
		}
	}
	return count
}

func (b *Browser) findItem(path string) *BrowserItem {
	var find func(*BrowserItem) *BrowserItem
	find = func(item *BrowserItem) *BrowserItem {
		if item.Path == path {
			return item
		}
		for _, child := range item.Children {
			if found := find(child); found != nil {
				return found
			}
		}
		return nil
	}
	return find(b.Root)
}

func (item *BrowserItem) Render(selected, cursor bool) string {
	var sb strings.Builder

	sb.WriteString(strings.Repeat("  ", item.Depth))

	if cursor {
		sb.WriteString(cursorStyle.Render("> "))
	} else {
		sb.WriteString("  ")
	}

	if selected {
		sb.WriteString(checkboxChecked + " ")
	} else {
		sb.WriteString(checkboxUnchecked + " ")
	}

	if item.Type == provider.EntryTypeDir {
		arrow := ">"
		if item.Expanded {
			arrow = "v"
		}
		sb.WriteString(dirStyle.Render(fmt.Sprintf("%s %s/", arrow, item.Name)))
	} else {
		style := fileStyle
		if selected {
			style = selectedFileStyle
		}
		sb.WriteString(style.Render(item.Name))
		if item.Size > 0 {
			sb.WriteString("  " + FormatSize(item.Size))
		}
	}

	return sb.String()
}

func GetFileExtension(path string) string {
	return strings.ToLower(filepath.Ext(path))
}
