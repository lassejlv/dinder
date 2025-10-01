package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Screen int

const (
	ScreenLoading Screen = iota
	ScreenReview
	ScreenConfirm
	ScreenProgress
	ScreenComplete
)

type model struct {
	screen      Screen
	files       []FileItem
	currentFile int
	toDelete    []FileItem
	spinner     int
	progress    int
	maxProgress int
	err         error
}

type filesLoadedMsg []FileItem
type deletionCompleteMsg struct{}
type tickMsg time.Time

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	fileStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			Width(50)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			Padding(0, 3).
			MarginTop(1)

	keepButtonStyle = buttonStyle.Copy().
			Background(lipgloss.Color("#04B575"))

	deleteButtonStyle = buttonStyle.Copy().
			Background(lipgloss.Color("#FF5F56"))

	progressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
)

func initialModel() model {
	return model{
		screen:  ScreenLoading,
		spinner: 0,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tick(),
		loadFiles,
	)
}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loadFiles() tea.Msg {
	files, err := scanDirectory(".")
	if err != nil {
		return err
	}
	return filesLoadedMsg(files)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.screen {
		case ScreenReview:
			return m.handleReviewInput(msg)
		case ScreenConfirm:
			return m.handleConfirmInput(msg)
		case ScreenComplete:
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
		
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case filesLoadedMsg:
		m.files = []FileItem(msg)
		if len(m.files) == 0 {
			m.screen = ScreenComplete
		} else {
			m.screen = ScreenReview
		}
		return m, nil

	case deletionCompleteMsg:
		m.screen = ScreenComplete
		return m, nil

	case tickMsg:
		if m.screen == ScreenLoading || m.screen == ScreenProgress {
			m.spinner = (m.spinner + 1) % len(spinnerFrames)
			return m, tick()
		}

	case error:
		m.err = msg
		return m, tea.Quit
	}

	return m, nil
}

func (m model) handleReviewInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "right", "l", "y":
		m.files[m.currentFile].Keep = true
		m.files[m.currentFile].Decided = true
		return m.nextFile()
	case "left", "h", "n":
		m.files[m.currentFile].Keep = false
		m.files[m.currentFile].Decided = true
		return m.nextFile()
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handleConfirmInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		m.screen = ScreenProgress
		m.maxProgress = len(m.toDelete)
		return m, tea.Batch(tick(), m.deleteFiles())
	case "n", "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) nextFile() (tea.Model, tea.Cmd) {
	m.currentFile++
	if m.currentFile >= len(m.files) {
		m.prepareConfirmation()
		m.screen = ScreenConfirm
	}
	return m, nil
}

func (m *model) prepareConfirmation() {
	m.toDelete = []FileItem{}
	for _, file := range m.files {
		if file.Decided && !file.Keep {
			m.toDelete = append(m.toDelete, file)
		}
	}
}

func (m model) deleteFiles() tea.Cmd {
	return func() tea.Msg {
		for _, file := range m.toDelete {
			os.RemoveAll(file.Path)
		}
		return deletionCompleteMsg{}
	}
}

func (m model) View() string {
	switch m.screen {
	case ScreenLoading:
		return fmt.Sprintf("\n%s Loading files...\n", spinnerFrames[m.spinner])

	case ScreenReview:
		if m.currentFile >= len(m.files) {
			return "No more files to review"
		}
		
		file := m.files[m.currentFile]
		fileType := "FILE"
		if file.IsDir {
			fileType = "DIR"
		}
		
		content := fmt.Sprintf("%s\n%s\n\nSize: %d bytes", 
			fileType, file.Path, file.Size)
		
		fileBox := fileStyle.Render(content)
		
		keepBtn := keepButtonStyle.Render("✓ Keep (→/l/y)")
		deleteBtn := deleteButtonStyle.Render("✗ Delete (←/h/n)")
		
		buttons := lipgloss.JoinHorizontal(lipgloss.Top, keepBtn, "  ", deleteBtn)
		
		progress := fmt.Sprintf("Progress: %d/%d", m.currentFile+1, len(m.files))
		
		return fmt.Sprintf("\n%s\n\n%s\n\n%s\n\n%s\n\nPress q to quit",
			titleStyle.Render("File Review"),
			fileBox,
			buttons,
			progress,
		)

	case ScreenConfirm:
		if len(m.toDelete) == 0 {
			return "\n" + titleStyle.Render("Complete") + "\n\nNo files selected for deletion.\n\nPress q to quit"
		}
		
		var deleteList strings.Builder
		for _, file := range m.toDelete {
			deleteList.WriteString(fmt.Sprintf("  %s\n", file.Path))
		}
		
		return fmt.Sprintf("\n%s\n\nFiles to delete (%d):\n%s\nConfirm deletion? (y/n)",
			titleStyle.Render("Confirmation"),
			len(m.toDelete),
			deleteList.String(),
		)

	case ScreenProgress:
		bar := progressStyle.Render(fmt.Sprintf("%s Deleting files... %d/%d", 
			spinnerFrames[m.spinner], m.progress, m.maxProgress))
		return fmt.Sprintf("\n%s\n\n%s", titleStyle.Render("Progress"), bar)

	case ScreenComplete:
		return fmt.Sprintf("\n%s\n\nDeletion complete!\n\nPress q to quit",
			titleStyle.Render("Complete"))
	}

	return ""
}
