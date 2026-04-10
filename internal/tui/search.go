package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type SearchModel struct {
	Input    textinput.Model
	Active   bool
	OnChange func(string)
}

func NewSearchModel() SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Prompt = searchPromptStyle.Render(" / ")
	ti.CharLimit = 100
	ti.Width = 30
	return SearchModel{Input: ti}
}

func (s SearchModel) Init() tea.Cmd {
	return nil
}

func (s SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	if !s.Active {
		return s, nil
	}

	var cmd tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "enter" || keyMsg.String() == "esc" {
			s.Active = false
			s.Input.Blur()
			return s, nil
		}
	}

	s.Input, cmd = s.Input.Update(msg)

	if s.OnChange != nil {
		s.OnChange(s.Input.Value())
	}

	return s, cmd
}

func (s SearchModel) View() string {
	if !s.Active {
		return ""
	}
	return searchStyle.Render(s.Input.View())
}

func (s *SearchModel) Focus() tea.Cmd {
	s.Active = true
	return s.Input.Focus()
}

func (s *SearchModel) Clear() {
	s.Input.SetValue("")
	if s.OnChange != nil {
		s.OnChange("")
	}
}
