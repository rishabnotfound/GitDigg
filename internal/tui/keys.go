package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Enter       key.Binding
	Space       key.Binding
	Search      key.Binding
	Escape      key.Binding
	Quit        key.Binding
	Help        key.Binding
	Download    key.Binding
	SelectAll   key.Binding
	DeselectAll key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:          key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:        key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Left:        key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "back")),
		Right:       key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "enter")),
		Enter:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "expand")),
		Space:       key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "select")),
		Search:      key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		Escape:      key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Quit:        key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Download:    key.NewBinding(key.WithKeys("d", "ctrl+d"), key.WithHelp("d", "download")),
		SelectAll:   key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "select all")),
		DeselectAll: key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "deselect all")),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Space, k.Download, k.Search, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Space, k.SelectAll, k.DeselectAll},
		{k.Search, k.Download, k.Quit},
	}
}
