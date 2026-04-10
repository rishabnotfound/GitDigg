package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rishabnotfound/gitdig/internal/download"
	"github.com/rishabnotfound/gitdig/internal/provider"
)

type ViewState int

const (
	ViewBrowser ViewState = iota
	ViewDownloading
	ViewComplete
)

type Model struct {
	Provider    provider.Provider
	RepoInfo    *provider.RepoInfo
	Entries     []provider.TreeEntry
	Ref         string
	Browser     *Browser
	Search      SearchModel
	Progress    ProgressModel
	Help        help.Model
	Keys        KeyMap
	State       ViewState
	Width       int
	Height      int
	Ready       bool
	Error       error
	ShowHelp    bool
	OutputDir   string
	Concurrency int
	Flat        bool
}

type TreeLoadedMsg struct {
	Entries []provider.TreeEntry
	Error   error
}

func (m Model) Init() tea.Cmd {
	return m.loadTree()
}

func (m Model) loadTree() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		ref := m.Ref
		if ref == "" {
			var err error
			ref, err = m.Provider.GetDefaultBranch(ctx, m.RepoInfo.Owner, m.RepoInfo.Repo)
			if err != nil {
				return TreeLoadedMsg{Error: err}
			}
			m.Ref = ref
		}

		entries, err := m.Provider.GetTree(ctx, m.RepoInfo.Owner, m.RepoInfo.Repo, provider.TreeOptions{
			Ref:       ref,
			Recursive: true,
		})

		return TreeLoadedMsg{Entries: entries, Error: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Progress.Width = msg.Width - 4
		m.Help.Width = msg.Width
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, m.Keys.Quit) && !m.Search.Active {
			return m, tea.Quit
		}

		switch m.State {
		case ViewBrowser:
			return m.updateBrowser(msg)
		case ViewDownloading:
			return m, nil
		case ViewComplete:
			return m, tea.Quit
		}

	case TreeLoadedMsg:
		if msg.Error != nil {
			m.Error = msg.Error
			return m, nil
		}
		m.Entries = msg.Entries
		m.Browser = NewBrowser(msg.Entries)
		m.Ready = true
		return m, nil

	case ProgressUpdateMsg:
		m.Progress, _ = m.Progress.Update(msg)
		return m, nil

	case DownloadCompleteMsg:
		m.Progress, _ = m.Progress.Update(msg)
		m.State = ViewComplete
		return m, nil
	}

	return m, nil
}

func (m Model) updateBrowser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.Search.Active {
		var cmd tea.Cmd
		m.Search, cmd = m.Search.Update(msg)
		return m, cmd
	}

	switch {
	case key.Matches(msg, m.Keys.Up):
		m.Browser.MoveUp()
	case key.Matches(msg, m.Keys.Down):
		m.Browser.MoveDown()
	case key.Matches(msg, m.Keys.Enter), key.Matches(msg, m.Keys.Right):
		m.Browser.Enter()
	case key.Matches(msg, m.Keys.Left):
		m.Browser.Back()
	case key.Matches(msg, m.Keys.Space):
		m.Browser.ToggleSelect()
		m.Browser.MoveDown()
	case key.Matches(msg, m.Keys.SelectAll):
		m.Browser.SelectAll()
	case key.Matches(msg, m.Keys.DeselectAll):
		m.Browser.DeselectAll()
	case key.Matches(msg, m.Keys.Search):
		m.Search.OnChange = func(v string) { m.Browser.SetFilter(v) }
		return m, m.Search.Focus()
	case key.Matches(msg, m.Keys.Escape):
		m.Search.Clear()
		m.Browser.SetFilter("")
	case key.Matches(msg, m.Keys.Download):
		return m.startDownload()
	case key.Matches(msg, m.Keys.Help):
		m.ShowHelp = !m.ShowHelp
	}

	return m, nil
}

func (m Model) startDownload() (tea.Model, tea.Cmd) {
	files := m.Browser.GetSelectedFiles()
	if len(files) == 0 {
		return m, nil
	}

	var selected []provider.TreeEntry
	selMap := make(map[string]bool)
	for _, f := range files {
		selMap[f] = true
	}
	for _, e := range m.Entries {
		if selMap[e.Path] {
			selected = append(selected, e)
		}
	}

	m.State = ViewDownloading
	m.Progress = NewProgressModel(len(files))
	m.Progress.Width = m.Width - 4

	return m, m.runDownload(selected)
}

func (m Model) runDownload(entries []provider.TreeEntry) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		opts := download.Options{
			Concurrency: m.Concurrency,
			OutputDir:   m.OutputDir,
			Flat:        m.Flat,
			Ref:         m.Ref,
		}

		mgr := download.NewManager(m.Provider, m.RepoInfo.Owner, m.RepoInfo.Repo, opts)
		err := mgr.Download(ctx, entries)

		return DownloadCompleteMsg{Error: err}
	}
}

func (m Model) View() string {
	if m.Error != nil {
		return appStyle.Render(errorStyle.Render(fmt.Sprintf("Error: %v", m.Error)))
	}

	if !m.Ready {
		return appStyle.Render(subtitleStyle.Render("Loading..."))
	}

	var content string
	switch m.State {
	case ViewBrowser:
		content = m.browserView()
	case ViewDownloading, ViewComplete:
		content = m.Progress.View()
	}

	return appStyle.Render(content)
}

func (m Model) browserView() string {
	var sb strings.Builder

	header := fmt.Sprintf("%s/%s", m.RepoInfo.Owner, m.RepoInfo.Repo)
	if m.Ref != "" {
		header += fmt.Sprintf(" @ %s", m.Ref)
	}
	sb.WriteString(headerStyle.Render(header))
	sb.WriteString("\n")

	if m.Search.Active {
		sb.WriteString(m.Search.View())
		sb.WriteString("\n")
	} else if m.Browser.SearchFilter != "" {
		sb.WriteString(subtitleStyle.Render(fmt.Sprintf("Filter: %s", m.Browser.SearchFilter)))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	listHeight := m.Height - 10
	if listHeight < 5 {
		listHeight = 5
	}

	startIdx := 0
	if m.Browser.Cursor >= listHeight {
		startIdx = m.Browser.Cursor - listHeight + 1
	}

	endIdx := startIdx + listHeight
	if endIdx > len(m.Browser.Items) {
		endIdx = len(m.Browser.Items)
	}

	for i := startIdx; i < endIdx; i++ {
		item := m.Browser.Items[i]
		sb.WriteString(item.Render(m.Browser.Selected[item.Path], i == m.Browser.Cursor))
		sb.WriteString("\n")
	}

	for i := endIdx - startIdx; i < listHeight; i++ {
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	status := fmt.Sprintf("%d files selected", m.Browser.GetSelectedCount())
	if len(m.Browser.Items) > 0 {
		status += fmt.Sprintf(" | %d/%d", m.Browser.Cursor+1, len(m.Browser.Items))
	}
	sb.WriteString(statusStyle.Render(status))
	sb.WriteString("\n")

	if m.ShowHelp {
		sb.WriteString("\n")
		sb.WriteString(m.Help.View(m.Keys))
	} else {
		sb.WriteString(helpStyle.Render("? help | space select | d download | / search | q quit"))
	}

	return sb.String()
}

func NewModel(p provider.Provider, info *provider.RepoInfo, ref, outputDir string, concurrency int, flat bool) Model {
	return Model{
		Provider:    p,
		RepoInfo:    info,
		Ref:         ref,
		OutputDir:   outputDir,
		Concurrency: concurrency,
		Flat:        flat,
		Search:      NewSearchModel(),
		Help:        help.New(),
		Keys:        DefaultKeyMap(),
		State:       ViewBrowser,
	}
}

func Run(p provider.Provider, info *provider.RepoInfo, ref, outputDir string, concurrency int, flat bool) error {
	m := NewModel(p, info, ref, outputDir, concurrency, flat)
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}
