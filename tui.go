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
	screen       Screen
	files        []FileItem
	currentFile  int
	toDelete     []FileItem
	toSkip       []FileItem
	spinner      int
	progress     int
	maxProgress  int
	totalSize    int64
	deletedSize  int64
	err          error
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
	case "s":
		m.files[m.currentFile].Skipped = true
		return m.nextFile()
	case "u":
		if m.currentFile > 0 {
			m.currentFile--
			m.files[m.currentFile].Decided = false
			m.files[m.currentFile].Skipped = false
		}
		return m, nil
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
	for {
		m.currentFile++
		if m.currentFile >= len(m.files) {
			m.prepareConfirmation()
			m.screen = ScreenConfirm
			break
		}
		if !m.files[m.currentFile].Skipped {
			break
		}
	}
	return m, nil
}

func (m *model) prepareConfirmation() {
	m.toDelete = []FileItem{}
	m.toSkip = []FileItem{}
	m.totalSize = 0
	
	for _, file := range m.files {
		if file.Decided && !file.Keep {
			m.toDelete = append(m.toDelete, file)
			m.totalSize += file.Size
		} else if file.Skipped {
			m.toSkip = append(m.toSkip, file)
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
		
		sizeStr := formatSize(file.Size)
		dateStr := file.ModTime.Format("2006-01-02 15:04")
		
		content := fmt.Sprintf("%s\n%s\n\nSize: %s\nModified: %s", 
			fileType, file.Path, sizeStr, dateStr)
		
		if file.Preview != "" {
			content += "\n\nPreview:\n" + file.Preview
		}
		
		fileBox := fileStyle.Render(content)
		
		keepBtn := keepButtonStyle.Render("✓ Keep (→/l/y)")
		deleteBtn := deleteButtonStyle.Render("✗ Delete (←/h/n)")
		skipBtn := buttonStyle.Render("↷ Skip (s)")
		
		buttons := lipgloss.JoinHorizontal(lipgloss.Top, keepBtn, "  ", deleteBtn, "  ", skipBtn)
		
		progress := fmt.Sprintf("Progress: %d/%d", m.currentFile+1, len(m.files))
		
		controls := "Controls: u=undo last | q=quit"
		
		return fmt.Sprintf("\n%s\n\n%s\n\n%s\n\n%s\n%s",
			titleStyle.Render("File Review"),
			fileBox,
			buttons,
			progress,
			controls,
		)

	case ScreenConfirm:
		if len(m.toDelete) == 0 {
			skippedInfo := ""
			if len(m.toSkip) > 0 {
				skippedInfo = fmt.Sprintf("\n%d files skipped for later review.", len(m.toSkip))
			}
			return "\n" + titleStyle.Render("Complete") + "\n\nNo files selected for deletion." + skippedInfo + "\n\nPress q to quit"
		}
		
		var deleteList strings.Builder
		for _, file := range m.toDelete {
			deleteList.WriteString(fmt.Sprintf("  %s (%s)\n", file.Path, formatSize(file.Size)))
		}
		
		sizeInfo := fmt.Sprintf("Total size: %s", formatSize(m.totalSize))
		skippedInfo := ""
		if len(m.toSkip) > 0 {
			skippedInfo = fmt.Sprintf("\n%d files skipped.", len(m.toSkip))
		}
		
		return fmt.Sprintf("\n%s\n\nFiles to delete (%d):\n%s\n%s%s\n\nConfirm deletion? (y/n)",
			titleStyle.Render("Confirmation"),
			len(m.toDelete),
			deleteList.String(),
			sizeInfo,
			skippedInfo,
		)

	case ScreenProgress:
		bar := progressStyle.Render(fmt.Sprintf("%s Deleting files... %d/%d", 
			spinnerFrames[m.spinner], m.progress, m.maxProgress))
		return fmt.Sprintf("\n%s\n\n%s", titleStyle.Render("Progress"), bar)

	case ScreenComplete:
		stats := fmt.Sprintf("Files deleted: %d\nSpace freed: %s", 
			len(m.toDelete), formatSize(m.totalSize))
		
		skippedInfo := ""
		if len(m.toSkip) > 0 {
			skippedInfo = fmt.Sprintf("\n%d files were skipped.", len(m.toSkip))
		}
		
		return fmt.Sprintf("\n%s\n\nDeletion complete!\n\n%s%s\n\nPress q to quit",
			titleStyle.Render("Complete"), stats, skippedInfo)

	}

	return ""
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
