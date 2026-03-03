package ui

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// File type colors
	DirStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))  // Blueish
	ExeStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // Green
	TextStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // Yellow
	CodeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("213")) // Pink/Magenta
	DefaultStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")) // Default white/grey

	PromptSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	PromptCursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

type promptModel struct {
	textInput textinput.Model
	files     []fs.DirEntry
	filtered  []fs.DirEntry
	cursor    int

	showOverlay  bool
	searchPrefix string
	atIndex      int // The index in the text input where '@' was typed

	isDone bool
	prompt string // Final resulting prompt
}

func InitialPromptModel() promptModel {
	ti := textinput.New()
	ti.Placeholder = "输入你要执行的操作 (输入 @ 补全当前目录文件)..."
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 80
	ti.Prompt = "🐆 > "
	ti.PromptStyle = PromptSelectedStyle
	ti.Cursor.Style = PromptCursorStyle

	return promptModel{
		textInput: ti,
		files:     []fs.DirEntry{},
		filtered:  []fs.DirEntry{},
	}
}

func (m promptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.showOverlay {
				m.showOverlay = false
				return m, nil
			}
			m.isDone = true
			return m, tea.Quit

		case tea.KeyEnter:
			if m.showOverlay {
				// Insert the selected file
				if len(m.filtered) > 0 && m.cursor >= 0 && m.cursor < len(m.filtered) {
					selected := m.filtered[m.cursor].Name()

					// Reconstruct text
					val := m.textInput.Value()
					before := val[:m.atIndex]
					// We want to replace the '@' and whatever the search prefix currently is
					// But we also need to consider exactly where the cursor is
					// Simplest is to replace from atIndex to the current cursor position
					currPos := m.textInput.Position()
					after := ""
					if currPos < len(val) {
						after = val[currPos:]
					}

					// Add spaces around the file if it has spaces
					if strings.Contains(selected, " ") {
						selected = fmt.Sprintf("\"%s\"", selected)
					}

					m.textInput.SetValue(before + selected + " " + after)
					m.textInput.SetCursor(len(before) + len(selected) + 1)
				}
				m.showOverlay = false
				return m, nil
			} else {
				// Submit the prompt string
				m.prompt = strings.TrimSpace(m.textInput.Value())
				if m.prompt != "" {
					m.isDone = true
					return m, tea.Quit
				}
			}

		case tea.KeyUp:
			if m.showOverlay && m.cursor > 0 {
				m.cursor--
				return m, nil
			}

		case tea.KeyDown:
			if m.showOverlay && m.cursor < len(m.filtered)-1 {
				m.cursor++
				return m, nil
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)

	// After updating text input, check for `@` to show overlay
	m.updateOverlayState()

	return m, cmd
}

func (m *promptModel) updateOverlayState() {
	val := m.textInput.Value()
	pos := m.textInput.Position()

	// Find the closest '@' before the cursor
	atPos := -1
	for i := pos - 1; i >= 0; i-- {
		if val[i] == '@' {
			// Ensure it's either at the beginning or preceded by a space
			if i == 0 || val[i-1] == ' ' || val[i-1] == '\t' {
				atPos = i
				break
			}
		}
		// If we hit a space before finding '@', we might not be in an uncompleted token
		if val[i] == ' ' || val[i] == '\t' {
			break
		}
	}

	if atPos >= 0 {
		m.showOverlay = true
		m.atIndex = atPos
		m.searchPrefix = val[atPos+1 : pos]
		m.loadAndFilterFiles()
	} else {
		m.showOverlay = false
	}
}

func (m *promptModel) loadAndFilterFiles() {
	if len(m.files) == 0 {
		// Lazy load directory contents
		entries, err := os.ReadDir(".")
		if err == nil {
			m.files = entries
		}
	}

	prefix := strings.ToLower(m.searchPrefix)
	var filtered []fs.DirEntry
	for _, f := range m.files {
		// Include if matched completely, or prefix match
		if prefix == "" || strings.HasPrefix(strings.ToLower(f.Name()), prefix) || strings.Contains(strings.ToLower(f.Name()), prefix) {
			filtered = append(filtered, f)
		}
	}

	m.filtered = filtered

	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
		if m.cursor < 0 {
			m.cursor = 0
		}
	}
}

func getFileStyle(f fs.DirEntry) lipgloss.Style {
	if f.IsDir() {
		return DirStyle
	}

	name := strings.ToLower(f.Name())
	ext := filepath.Ext(name)

	if ext == ".exe" || ext == ".sh" || ext == ".bat" || ext == ".cmd" {
		return ExeStyle
	}

	if ext == ".md" || ext == ".txt" || ext == ".csv" || ext == ".json" || ext == ".yaml" || ext == ".yml" {
		return TextStyle
	}

	if ext == ".go" || ext == ".js" || ext == ".ts" || ext == ".py" || ext == ".java" || ext == ".c" || ext == ".cpp" || ext == ".rs" || ext == ".html" || ext == ".css" {
		return CodeStyle
	}

	return DefaultStyle
}

func (m promptModel) View() string {
	var sb strings.Builder

	sb.WriteString("\n")

	// Draw text input
	sb.WriteString(m.textInput.View())
	sb.WriteString("\n")

	// Draw overlay
	if m.showOverlay {
		sb.WriteString("\n")
		if len(m.filtered) == 0 {
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  (无匹配文件) / (No matching files)"))
			sb.WriteString("\n")
		} else {
			// Show top 10 matches
			maxItems := 10
			startIdx := 0
			if m.cursor >= maxItems {
				startIdx = m.cursor - maxItems + 1
			}
			endIdx := startIdx + maxItems
			if endIdx > len(m.filtered) {
				endIdx = len(m.filtered)
			}

			for i := startIdx; i < endIdx; i++ {
				f := m.filtered[i]
				cursor := "  "
				style := getFileStyle(f)

				if m.cursor == i {
					cursor = "> "
					style = style.Bold(true).Background(lipgloss.Color("236")).Underline(true)
				}

				displayName := f.Name()
				if f.IsDir() {
					displayName += "/"
				}

				sb.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(displayName)))
			}
			if len(m.filtered) > maxItems {
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("  ... and %d more", len(m.filtered)-maxItems)))
				sb.WriteString("\n")
			}
		}
	} else {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("💡 提示: 输入 @ 自动补全当前目录的文件\n"))
	}

	return sb.String()
}

// RunPromptUI displays an interactive prompt text box
func RunPromptUI() (string, error) {
	p := tea.NewProgram(InitialPromptModel())
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	finalModel := m.(promptModel)
	if finalModel.prompt != "" {
		return finalModel.prompt, nil
	}

	return "", nil
}
