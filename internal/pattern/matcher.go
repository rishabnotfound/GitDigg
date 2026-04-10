package pattern

import (
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rishabnotfound/gitdigg/internal/provider"
)

type Matcher struct {
	patterns []string
}

func NewMatcher(patterns []string) *Matcher {
	return &Matcher{patterns: normalize(patterns)}
}

func normalize(patterns []string) []string {
	result := make([]string, 0, len(patterns))
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		result = append(result, strings.TrimPrefix(p, "/"))
	}
	return result
}

func (m *Matcher) Match(entries []provider.TreeEntry) []provider.TreeEntry {
	if len(m.patterns) == 0 {
		return entries
	}

	var matched []provider.TreeEntry
	seen := make(map[string]bool)

	for _, entry := range entries {
		if seen[entry.Path] {
			continue
		}
		if m.matches(entry) {
			matched = append(matched, entry)
			seen[entry.Path] = true
		}
	}

	return matched
}

func (m *Matcher) matches(entry provider.TreeEntry) bool {
	for _, pattern := range m.patterns {
		if m.matchPattern(pattern, entry) {
			return true
		}
	}
	return false
}

func (m *Matcher) matchPattern(pattern string, entry provider.TreeEntry) bool {
	path := entry.Path

	if pattern == path {
		return true
	}

	if strings.HasSuffix(pattern, "/") {
		dir := strings.TrimSuffix(pattern, "/")
		if path == dir || strings.HasPrefix(path, dir+"/") {
			return true
		}
	}

	if !hasGlob(pattern) {
		if strings.HasPrefix(path, pattern+"/") || path == pattern {
			return true
		}
	}

	if hasGlob(pattern) {
		if matched, _ := doublestar.Match(pattern, path); matched {
			return true
		}
		if !strings.HasPrefix(pattern, "**/") {
			if matched, _ := doublestar.Match("**/"+pattern, path); matched {
				return true
			}
		}
	}

	return false
}

func hasGlob(s string) bool {
	return strings.ContainsAny(s, "*?[{")
}

func (m *Matcher) ExpandDirectories(entries []provider.TreeEntry) []provider.TreeEntry {
	matched := m.Match(entries)
	if len(matched) == 0 {
		return nil
	}

	dirs := make(map[string]bool)
	for _, entry := range matched {
		if entry.Type == provider.EntryTypeDir {
			dirs[entry.Path] = true
		}
	}

	if len(dirs) > 0 {
		for _, entry := range entries {
			if entry.Type != provider.EntryTypeFile {
				continue
			}
			dir := filepath.Dir(entry.Path)
			for dir != "." && dir != "" {
				if dirs[dir] {
					matched = append(matched, entry)
					break
				}
				dir = filepath.Dir(dir)
			}
		}
	}

	seen := make(map[string]bool)
	var result []provider.TreeEntry
	for _, entry := range matched {
		if !seen[entry.Path] {
			seen[entry.Path] = true
			result = append(result, entry)
		}
	}

	return result
}

func FilterFiles(entries []provider.TreeEntry) []provider.TreeEntry {
	var files []provider.TreeEntry
	for _, entry := range entries {
		if entry.Type == provider.EntryTypeFile {
			files = append(files, entry)
		}
	}
	return files
}
