package ui

import (
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/snowtema/drift/internal/protocol"
)

// ─── List Mode ───────────────────────────────────

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxIdx := m.maxListIdx()

	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "?":
		m.prevFocus = m.focus
		m.focus = FocusHelp
		return m, nil
	case "up", "k":
		if maxIdx >= 0 {
			m.listIdx = (m.listIdx - 1 + maxIdx + 1) % (maxIdx + 1)
			// In tree mode, skip directory-only lines
			if m.viewMode == ViewTree {
				m.skipDirLines(-1)
			}
		}
	case "down", "j":
		if maxIdx >= 0 {
			m.listIdx = (m.listIdx + 1) % (maxIdx + 1)
			if m.viewMode == ViewTree {
				m.skipDirLines(1)
			}
		}
	case "home", "g":
		m.listIdx = 0
		m.listScroll = 0
		if m.viewMode == ViewTree {
			m.skipDirLines(1)
		}
	case "end", "G":
		if maxIdx >= 0 {
			m.listIdx = maxIdx
			if m.viewMode == ViewTree {
				m.skipDirLines(-1)
			}
		}
	case "pgup", "ctrl+u":
		m.listIdx = Clamp(m.listIdx-m.listHeight(), 0, maxIdx)
	case "pgdown", "ctrl+d":
		m.listIdx = Clamp(m.listIdx+m.listHeight(), 0, maxIdx)

	// Enter detail
	case "enter", "right", "l":
		if m.selectedProject() != nil {
			m.focus = FocusDetail
			m.detailSection = SectionInfo
			m.detailCursor = 0
			m.detailScroll = 0
		}

	// Sort
	case "s":
		m.sortMode = (m.sortMode + 1) % 4
		m.applyFilterAndSort()
		m.listIdx = 0
		m.listScroll = 0

	// Toggle tree/flat view
	case "t":
		if m.viewMode == ViewFlat {
			m.viewMode = ViewTree
		} else {
			m.viewMode = ViewFlat
		}
		m.listIdx = 0
		m.listScroll = 0

	// Filter
	case "/":
		m.startInput(InputFilter, "/")

	// Jump
	case ":":
		m.startInput(InputJump, ":")

	// Clear filter
	case "esc":
		if m.filterText != "" {
			m.filterText = ""
			m.applyFilterAndSort()
			m.listIdx = 0
			m.listScroll = 0
		}
	}

	m.keepListInView()
	return m, nil
}

// ─── Detail Mode ─────────────────────────────────

func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	p := m.selectedProject()
	if p == nil {
		m.focus = FocusList
		return m, nil
	}

	switch msg.String() {
	// Back to list
	case "esc", "left", "h":
		m.focus = FocusList
		return m, nil

	case "?":
		m.prevFocus = m.focus
		m.focus = FocusHelp
		return m, nil

	// Navigate within detail: sections + items
	case "up", "k":
		if m.detailSection == SectionGoals && m.detailCursor > 0 {
			m.detailCursor--
		} else if m.detailSection == SectionNotes && m.detailCursor > 0 {
			m.detailCursor--
		} else if int(m.detailSection) > 0 {
			m.detailSection--
			// Set cursor to last item of previous section
			if m.detailSection == SectionGoals && len(p.Goals) > 0 {
				m.detailCursor = len(p.Goals) - 1
			} else {
				m.detailCursor = 0
			}
		}

	case "down", "j":
		maxCursor := m.maxCursorForSection(p)
		if m.detailCursor < maxCursor-1 {
			m.detailCursor++
		} else if int(m.detailSection) < int(SectionNotes) {
			m.detailSection++
			m.detailCursor = 0
		}

	// Sections via Tab
	case "tab":
		m.detailSection = DetailSection((int(m.detailSection) + 1) % 3)
		m.detailCursor = 0
	case "shift+tab":
		sec := int(m.detailSection) - 1
		if sec < 0 {
			sec = 2
		}
		m.detailSection = DetailSection(sec)
		m.detailCursor = 0

	// Switch project without leaving detail
	case "[":
		n := len(m.projects)
		if n > 0 {
			m.listIdx = (m.listIdx - 1 + n) % n
			m.keepListInView()
			m.resetDetail()
		}
	case "]":
		n := len(m.projects)
		if n > 0 {
			m.listIdx = (m.listIdx + 1) % n
			m.keepListInView()
			m.resetDetail()
		}

	// ── Actions ──

	// Add note
	case "n":
		m.startInput(InputNote, "note:")

	// Add goal
	case "g":
		m.startInput(InputGoal, "goal:")

	// Edit description
	case "D":
		desc := ""
		if p.Description != nil {
			desc = *p.Description
		}
		m.startInputWithValue(InputDesc, "desc:", desc)

	// Launch Claude Code in project directory
	case "c":
		return m, m.launchClaude(p)

	// Toggle goal done (Space or Enter on a goal)
	case "enter", " ":
		if m.detailSection == SectionGoals && len(p.Goals) > 0 && m.detailCursor < len(p.Goals) {
			idx := m.detailCursor
			m.mutate(func(proj *protocol.Project) {
				proj.Goals[idx].Done = !proj.Goals[idx].Done
				proj.Progress = protocol.CalcProgress(proj.Goals)
			})
			m.setFlash("✓ goal toggled")
			return m, FlashCmd()
		}

	// Delete goal (x on a goal, with confirm)
	case "x":
		if m.detailSection == SectionGoals && len(p.Goals) > 0 && m.detailCursor < len(p.Goals) {
			idx := m.detailCursor
			goalText := p.Goals[idx].Text
			m.confirm("Delete goal \""+Truncate(goalText, 30)+"\"? y/N", func() {
				m.mutate(func(proj *protocol.Project) {
					proj.Goals = append(proj.Goals[:idx], proj.Goals[idx+1:]...)
					proj.Progress = protocol.CalcProgress(proj.Goals)
				})
				if m.detailCursor >= len(m.selectedProject().Goals) && m.detailCursor > 0 {
					m.detailCursor--
				}
				m.setFlash("✓ goal deleted")
			})
			return m, nil
		}

	// Status: 1-5
	case "1":
		m.setStatus(protocol.StatusActive)
		return m, FlashCmd()
	case "2":
		m.setStatus(protocol.StatusIdea)
		return m, FlashCmd()
	case "3":
		m.setStatus(protocol.StatusPaused)
		return m, FlashCmd()
	case "4":
		m.setStatus(protocol.StatusDone)
		return m, FlashCmd()
	case "5":
		m.setStatus(protocol.StatusAbandoned)
		return m, FlashCmd()
	}

	return m, nil
}

func (m *Model) setStatus(s protocol.ProjectStatus) {
	m.mutate(func(p *protocol.Project) { p.Status = s })
	m.setFlash("✓ " + string(s))
}

func (m *Model) resetDetail() {
	m.detailSection = SectionInfo
	m.detailCursor = 0
	m.detailScroll = 0
}

func (m *Model) maxCursorForSection(p *protocol.FullProject) int {
	switch m.detailSection {
	case SectionGoals:
		return len(p.Goals)
	case SectionNotes:
		return len(p.Notes)
	default:
		return 1
	}
}

// ─── Input Mode ──────────────────────────────────

func (m Model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		mode := m.inputMode
		m.inputMode = InputNone
		m.textInput.Blur()
		// If was filter, clear filter text
		if mode == InputFilter {
			m.filterText = ""
			m.applyFilterAndSort()
		}
		return m, nil

	case "enter":
		val := strings.TrimSpace(m.textInput.Value())
		mode := m.inputMode
		m.inputMode = InputNone
		m.textInput.Blur()

		switch mode {
		case InputNote:
			if val != "" {
				m.mutate(func(p *protocol.Project) {
					p.Notes = append(p.Notes, protocol.Note{Ts: protocol.Now(), Text: val})
				})
				m.setFlash("✓ note added")
				return m, FlashCmd()
			}
		case InputGoal:
			if val != "" {
				m.mutate(func(p *protocol.Project) {
					p.Goals = append(p.Goals, protocol.Goal{Text: val, Done: false})
					p.Progress = protocol.CalcProgress(p.Goals)
				})
				m.setFlash("✓ goal added")
				return m, FlashCmd()
			}
		case InputDesc:
			m.mutate(func(p *protocol.Project) {
				if val == "" {
					p.Description = nil
				} else {
					p.Description = &val
				}
			})
			m.setFlash("✓ description updated")
			return m, FlashCmd()
		case InputFilter:
			m.filterText = val
			m.applyFilterAndSort()
			m.listIdx = 0
			m.listScroll = 0
		case InputJump:
			if val != "" {
				m.jumpToProject(val)
			}
		}
		return m, nil
	}

	// Live filter: update as user types
	if m.inputMode == InputFilter {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		m.filterText = m.textInput.Value()
		m.applyFilterAndSort()
		m.listIdx = 0
		m.listScroll = 0
		return m, cmd
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *Model) jumpToProject(name string) {
	name = strings.ToLower(name)
	for i, p := range m.projects {
		if strings.Contains(strings.ToLower(p.Name), name) {
			m.listIdx = i
			m.keepListInView()
			m.focus = FocusDetail
			m.resetDetail()
			m.setFlash("→ " + p.Name)
			return
		}
	}
	m.setFlash("not found: " + name)
}

// ─── Confirm Dialog ──────────────────────────────

func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.confirmAction != nil {
			m.confirmAction()
		}
		m.confirmMsg = ""
		m.confirmAction = nil
		return m, FlashCmd()
	case "n", "N", "esc":
		m.confirmMsg = ""
		m.confirmAction = nil
		return m, nil
	}
	return m, nil
}

// ─── Launch Claude Code ──────────────────────────

type claudeExitMsg struct{}

func (m *Model) launchClaude(p *protocol.FullProject) tea.Cmd {
	c := exec.Command("claude")
	c.Dir = p.Path

	return tea.ExecProcess(c, func(err error) tea.Msg {
		return claudeExitMsg{}
	})
}

// ─── Help Overlay ────────────────────────────────

func (m Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?", "esc", "q":
		m.focus = m.prevFocus
		return m, nil
	}
	return m, nil
}
