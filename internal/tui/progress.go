package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rishabnotfound/gitdig/internal/download"
)

type ProgressModel struct {
	TotalFiles     int
	CompletedFiles int
	FailedFiles    int
	CurrentFile    string
	Done           bool
	Error          error
	Width          int
}

type ProgressUpdateMsg struct {
	Update *download.ProgressUpdate
}

type DownloadCompleteMsg struct {
	Error error
}

func NewProgressModel(totalFiles int) ProgressModel {
	return ProgressModel{TotalFiles: totalFiles, Width: 60}
}

func (p ProgressModel) Init() tea.Cmd {
	return nil
}

func (p ProgressModel) Update(msg tea.Msg) (ProgressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ProgressUpdateMsg:
		p.CompletedFiles = int(msg.Update.CompletedFiles)
		p.FailedFiles = int(msg.Update.FailedFiles)
		p.CurrentFile = msg.Update.CurrentFile
	case DownloadCompleteMsg:
		p.Done = true
		p.Error = msg.Error
	case tea.WindowSizeMsg:
		p.Width = msg.Width - 4
		if p.Width < 20 {
			p.Width = 20
		}
	}
	return p, nil
}

func (p ProgressModel) View() string {
	var sb strings.Builder

	if p.Done {
		if p.Error != nil {
			sb.WriteString(errorStyle.Render(fmt.Sprintf("Failed: %v", p.Error)))
		} else {
			sb.WriteString(lipgloss.NewStyle().Foreground(secondaryColor).Bold(true).Render(
				fmt.Sprintf("Done! %d files downloaded", p.CompletedFiles),
			))
			if p.FailedFiles > 0 {
				sb.WriteString(errorStyle.Render(fmt.Sprintf(" (%d failed)", p.FailedFiles)))
			}
		}
		return sb.String()
	}

	sb.WriteString(titleStyle.Render("Downloading..."))
	sb.WriteString("\n\n")

	progress := float64(p.CompletedFiles+p.FailedFiles) / float64(p.TotalFiles)
	barWidth := p.Width - 20
	if barWidth < 10 {
		barWidth = 10
	}
	filled := int(progress * float64(barWidth))
	empty := barWidth - filled

	bar := progressBarStyle.Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(subtleColor).Render(strings.Repeat("░", empty))

	sb.WriteString(fmt.Sprintf("  %s %3.0f%%\n", bar, progress*100))
	sb.WriteString("\n")
	sb.WriteString(progressTextStyle.Render(fmt.Sprintf("  Files: %d / %d", p.CompletedFiles+p.FailedFiles, p.TotalFiles)))

	if p.FailedFiles > 0 {
		sb.WriteString(errorStyle.Render(fmt.Sprintf(" (%d failed)", p.FailedFiles)))
	}
	sb.WriteString("\n")

	if p.CurrentFile != "" {
		file := p.CurrentFile
		maxLen := p.Width - 10
		if len(file) > maxLen {
			file = "..." + file[len(file)-maxLen+3:]
		}
		sb.WriteString(progressTextStyle.Render(fmt.Sprintf("  Current: %s", file)))
	}

	return sb.String()
}
