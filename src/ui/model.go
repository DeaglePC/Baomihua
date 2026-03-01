package ui

import (
	"fmt"
	"strings"
	"unicode"

	"baomihua/executor"
	"baomihua/guard"
	"baomihua/llm"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateLoading state = iota
	stateError
	stateResult
	stateDone
)

type Action int

const (
	ActionInject Action = iota
	ActionExecute
	ActionCopy
	ActionCancel
)

type menuItem struct {
	label  string
	action Action
}

type model struct {
	prompt    string
	isZH      bool
	ctx       llm.EnvContext
	state     state
	err       error
	spinner   spinner.Model
	parsed    *llm.Result
	safetyLvl guard.Level
	menuItems []menuItem
	cursor    int
	exitMsg   string
	isDone    bool
}

type errMsg struct{ err error }

func IsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func InitialModel(prompt string) model {
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		prompt:  prompt,
		isZH:    IsChinese(prompt),
		ctx:     llm.GetEnvContext(),
		state:   stateLoading,
		spinner: s,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startStreamingCmd(),
	)
}

func (m model) startStreamingCmd() tea.Cmd {
	return func() tea.Msg {
		contentChan := make(chan string)
		errChan := make(chan error)

		go llm.StreamCompletion(m.prompt, m.ctx, contentChan, errChan)

		var sb strings.Builder
		for {
			select {
			case content, ok := <-contentChan:
				if !ok {
					contentChan = nil
				} else {
					sb.WriteString(content)
				}
			case err, ok := <-errChan:
				if !ok {
					errChan = nil
				} else if err != nil {
					return errMsg{err: err}
				}
			}

			if contentChan == nil && errChan == nil {
				break
			}
		}

		res, err := llm.ParseResult(sb.String())
		if err != nil {
			return errMsg{err: err}
		}

		lvl := guard.CheckCommand(res.Command)

		var items []menuItem

		if lvl != guard.Danger {
			if m.isZH {
				items = append(items, menuItem{label: "⚡️ 直接执行 (Execute)", action: ActionExecute})
			} else {
				items = append(items, menuItem{label: "⚡️ Execute", action: ActionExecute})
			}
		}

		if m.isZH {
			items = append(items,
				menuItem{label: "🐾 插入终端 (Insert to prompt)", action: ActionInject},
				menuItem{label: "📋 复制命令 (Copy)", action: ActionCopy},
				menuItem{label: "🛑 放弃 (Cancel)", action: ActionCancel},
			)
		} else {
			items = append(items,
				menuItem{label: "🐾 Insert to prompt", action: ActionInject},
				menuItem{label: "📋 Copy", action: ActionCopy},
				menuItem{label: "🛑 Cancel", action: ActionCancel},
			)
		}

		return struct {
			res   *llm.Result
			lvl   guard.Level
			items []menuItem
		}{
			res:   res,
			lvl:   lvl,
			items: items,
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == stateResult {
			switch msg.String() {
			case "ctrl+c", "q":
				m.isDone = true
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.menuItems)-1 {
					m.cursor++
				}
			case "enter", " ":
				return m.handleChoice()
			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				idx := int(msg.String()[0] - '1')
				if idx >= 0 && idx < len(m.menuItems) {
					m.cursor = idx
					return m.handleChoice()
				}
			}
		} else {
			switch msg.String() {
			case "ctrl+c", "q":
				m.isDone = true
				return m, tea.Quit
			}
		}

	case errMsg:
		m.err = msg.err
		m.state = stateError
		return m, tea.Quit

	case struct {
		res   *llm.Result
		lvl   guard.Level
		items []menuItem
	}:
		m.parsed = msg.res
		m.safetyLvl = msg.lvl
		m.menuItems = msg.items
		m.state = stateResult

	case spinner.TickMsg:
		if m.state == stateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m model) handleChoice() (tea.Model, tea.Cmd) {
	selected := m.menuItems[m.cursor]

	switch selected.action {
	case ActionInject:
		err := executor.InjectToTerminal(m.parsed.Command)
		if err != nil {
			if m.isZH {
				m.exitMsg = fmt.Sprintf("\n❌ 插入失败: %v", err)
			} else {
				m.exitMsg = fmt.Sprintf("\n❌ Injection failed: %v", err)
			}
		} else {
			if m.isZH {
				m.exitMsg = "\n✅ 已写入终端! (请按回车执行)"
			} else {
				m.exitMsg = "\n✅ Injected into terminal! (Press Enter to execute)"
			}
		}
	case ActionExecute:
		if m.isZH {
			m.exitMsg = fmt.Sprintf("\n🚀 正在执行命令: %s", m.parsed.Command)
		} else {
			m.exitMsg = fmt.Sprintf("\n🚀 Executing command: %s", m.parsed.Command)
		}
	case ActionCopy:
		err := executor.CopyToClipboard(m.parsed.Command)
		if err != nil {
			if m.isZH {
				m.exitMsg = fmt.Sprintf("\n❌ 复制失败: %v", err)
			} else {
				m.exitMsg = fmt.Sprintf("\n❌ Copy failed: %v", err)
			}
		} else {
			if m.isZH {
				m.exitMsg = "\n✅ 已复制到剪贴板!"
			} else {
				m.exitMsg = "\n✅ Copied to clipboard!"
			}
		}
	case ActionCancel:
		if m.isZH {
			m.exitMsg = "\n🛑 已放弃执行"
		} else {
			m.exitMsg = "\n🛑 Execution canceled"
		}
	}

	m.isDone = true
	return m, tea.Quit
}

func (m model) View() string {
	if m.isDone {
		return ""
	}

	switch m.state {
	case stateError:
		if m.isZH {
			return DangerStyle.Render(fmt.Sprintf("\n❌ 发生错误: %v\n", m.err))
		}
		return DangerStyle.Render(fmt.Sprintf("\n❌ Error occurred: %v\n", m.err))
	case stateLoading:
		if m.isZH {
			return fmt.Sprintf("\n %s %s\n", m.spinner.View(), lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(fmt.Sprintf("豹米花正在潜伏思考如何 %q...", m.prompt)))
		}
		return fmt.Sprintf("\n %s %s\n", m.spinner.View(), lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(fmt.Sprintf("BaoMiHua is thinking about how to %q...", m.prompt)))

	case stateResult:
		var sb strings.Builder

		sb.WriteString("\n")
		if m.isZH {
			sb.WriteString(TitleStyle.Render("💻 命令 (Command): ") + TargetStyle.Render(m.parsed.Command) + "\n")
			sb.WriteString(TitleStyle.Render("🐆 解释 (Explanation): ") + ExplanationStyle.Render(m.parsed.Explanation) + "\n\n")
		} else {
			sb.WriteString(TitleStyle.Render("💻 Command: ") + TargetStyle.Render(m.parsed.Command) + "\n")
			sb.WriteString(TitleStyle.Render("🐆 Explanation: ") + ExplanationStyle.Render(m.parsed.Explanation) + "\n\n")
		}

		if m.safetyLvl == guard.Danger {
			if m.isZH {
				sb.WriteString(DangerStyle.Render("⚠️ 警告：豹米花察觉到极度危险的操作，请谨慎行事！") + "\n\n")
			} else {
				sb.WriteString(DangerStyle.Render("⚠️ Warning: BaoMiHua detected an extremely dangerous operation, proceed with caution!") + "\n\n")
			}
		}

		if m.isZH {
			sb.WriteString("请选择下一步动作:\n")
		} else {
			sb.WriteString("Select next action:\n")
		}

		for i, item := range m.menuItems {
			cursor := "  "
			style := ItemStyle
			if m.cursor == i {
				cursor = "> "
				style = SelectedItemStyle
			}

			sb.WriteString(style.Render(fmt.Sprintf("%s%d. %s", cursor, i+1, item.label)) + "\n")
		}

		return sb.String()
	}

	return ""
}

// RunUI is the entry point to start the BubbleTea program
func RunUI(prompt string) (*llm.Result, Action, string, error) {
	p := tea.NewProgram(InitialModel(prompt))
	m, err := p.Run()
	if err != nil {
		return nil, ActionCancel, "", err
	}

	finalModel := m.(model)
	if finalModel.state != stateResult {
		if finalModel.err != nil {
			return nil, ActionCancel, "", finalModel.err
		}
		return nil, ActionCancel, "", nil
	}

	if finalModel.cursor < len(finalModel.menuItems) {
		selected := finalModel.menuItems[finalModel.cursor]

		var sb strings.Builder
		if finalModel.parsed != nil {
			sb.WriteString("\n")
			if finalModel.isZH {
				sb.WriteString(TitleStyle.Render("💻 命令 (Command): ") + TargetStyle.Render(finalModel.parsed.Command) + "\n")
				sb.WriteString(TitleStyle.Render("🐆 解释 (Explanation): ") + ExplanationStyle.Render(finalModel.parsed.Explanation) + "\n\n")
				if finalModel.safetyLvl == guard.Danger {
					sb.WriteString(DangerStyle.Render("⚠️ 警告：豹米花察觉到极度危险的操作，请谨慎行事！") + "\n\n")
				}
			} else {
				sb.WriteString(TitleStyle.Render("💻 Command: ") + TargetStyle.Render(finalModel.parsed.Command) + "\n")
				sb.WriteString(TitleStyle.Render("🐆 Explanation: ") + ExplanationStyle.Render(finalModel.parsed.Explanation) + "\n\n")
				if finalModel.safetyLvl == guard.Danger {
					sb.WriteString(DangerStyle.Render("⚠️ Warning: BaoMiHua detected an extremely dangerous operation, proceed with caution!") + "\n\n")
				}
			}
		}
		sb.WriteString(finalModel.exitMsg + "\n")

		return finalModel.parsed, selected.action, sb.String(), nil
	}

	return finalModel.parsed, ActionCancel, "", nil
}
