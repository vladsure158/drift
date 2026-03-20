package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/snowtema/drift/internal/protocol"
)

// ─── State Machine ───────────────────────────────

type Focus int

const (
	FocusList   Focus = iota
	FocusDetail
	FocusHelp // ? overlay
)

type InputMode int

const (
	InputNone   InputMode = iota
	InputNote
	InputGoal
	InputDesc
	InputFilter // / filter in list
	InputJump   // : jump to project
)

type SortMode int

const (
	SortActivity SortMode = iota
	SortProgress
	SortName
	SortStatus
)

var SortLabels = []string{"recent", "progress", "name", "status"}

type ListViewMode int

const (
	ViewFlat ListViewMode = iota
	ViewTree
)

type DetailSection int

const (
	SectionInfo  DetailSection = iota
	SectionGoals
	SectionNotes
)

var SectionLabels = []string{"info", "goals", "notes"}

// ─── Model ───────────────────────────────────────

type Model struct {
	// Data
	allProjects []protocol.FullProject // unfiltered
	projects    []protocol.FullProject // filtered + sorted view
	filterText  string
	treeLines   []protocol.TreeLine    // tree view cache
	viewMode    ListViewMode

	// Layout
	width, height int

	// List
	listIdx    int
	listScroll int

	// Focus & mode
	focus     Focus
	prevFocus Focus // for returning from help

	// Sort
	sortMode SortMode

	// Detail
	detailSection DetailSection
	detailCursor  int
	detailScroll  int // scroll offset for detail content

	// Input
	inputMode InputMode
	textInput textinput.Model

	// Flash
	flash      string
	flashTicks int

	// Confirm dialog
	confirmMsg    string
	confirmAction func()

	// Header & theme
	version         string
	bannerCollapsed bool
	themeIdx        int
}

func NewModel(version string) Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 200

	// Load saved preferences
	cfg := protocol.ReadConfig()
	themeIdx := 0
	for i, t := range Themes {
		if t.Name == cfg.Theme {
			themeIdx = i
			break
		}
	}
	ApplyTheme(Themes[themeIdx])

	var viewMode ListViewMode
	if cfg.ViewMode == "tree" {
		viewMode = ViewTree
	}

	var sortMode SortMode
	for i, label := range SortLabels {
		if label == cfg.SortMode {
			sortMode = SortMode(i)
			break
		}
	}

	return Model{
		focus:           FocusList,
		textInput:       ti,
		version:         version,
		themeIdx:        themeIdx,
		bannerCollapsed: cfg.BannerCollapsed,
		viewMode:        viewMode,
		sortMode:        sortMode,
	}
}

func (m *Model) saveConfig() {
	viewMode := "flat"
	if m.viewMode == ViewTree {
		viewMode = "tree"
	}
	protocol.WriteConfig(protocol.UIConfig{
		Theme:           Themes[m.themeIdx].Name,
		ViewMode:        viewMode,
		SortMode:        SortLabels[int(m.sortMode)],
		BannerCollapsed: m.bannerCollapsed,
	})
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, loadProjectsCmd)
}

// ─── Messages ────────────────────────────────────

type projectsLoadedMsg struct{ projects []protocol.FullProject }
type flashTickMsg struct{}

func loadProjectsCmd() tea.Msg {
	projects, _ := protocol.LoadAllProjects()
	return projectsLoadedMsg{projects}
}

func FlashCmd() tea.Cmd {
	return tea.Tick(time.Second, func(_ time.Time) tea.Msg { return flashTickMsg{} })
}

// ─── Update Router ───────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case projectsLoadedMsg:
		m.allProjects = msg.projects
		m.applyFilterAndSort()
		return m, nil

	case claudeExitMsg:
		// Reload projects after returning from Claude Code
		return m, loadProjectsCmd

	case flashTickMsg:
		m.flashTicks--
		if m.flashTicks <= 0 {
			m.flash = ""
			return m, nil
		}
		return m, tea.Tick(time.Second, func(_ time.Time) tea.Msg { return flashTickMsg{} })

	case tea.KeyMsg:
		// Global: Ctrl+C always quits
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Confirm dialog takes priority
		if m.confirmMsg != "" {
			return m.updateConfirm(msg)
		}

		// Input mode
		if m.inputMode != InputNone {
			return m.updateInput(msg)
		}

		// Help overlay
		if m.focus == FocusHelp {
			return m.updateHelp(msg)
		}

		// Normal focus
		switch m.focus {
		case FocusList:
			return m.updateList(msg)
		case FocusDetail:
			return m.updateDetail(msg)
		}
	}
	return m, nil
}

// ─── Data helpers ────────────────────────────────

func (m *Model) applyFilterAndSort() {
	// Filter
	if m.filterText == "" {
		m.projects = make([]protocol.FullProject, len(m.allProjects))
		copy(m.projects, m.allProjects)
	} else {
		ft := strings.ToLower(m.filterText)
		m.projects = nil
		for _, p := range m.allProjects {
			if strings.Contains(strings.ToLower(p.Name), ft) ||
				(p.Description != nil && strings.Contains(strings.ToLower(*p.Description), ft)) ||
				strings.Contains(strings.ToLower(strings.Join(p.Tags, " ")), ft) {
				m.projects = append(m.projects, p)
			}
		}
	}
	// Sort
	m.sortProjects()
	// Build tree
	tree := protocol.BuildProjectTree(m.projects)
	m.treeLines = protocol.FlattenTree(tree, 0)
	// Clamp index
	maxIdx := m.maxListIdx()
	if m.listIdx > maxIdx {
		m.listIdx = maxIdx
	}
	if m.listIdx < 0 {
		m.listIdx = 0
	}
}

func (m *Model) maxListIdx() int {
	if m.viewMode == ViewTree {
		return len(m.treeLines) - 1
	}
	return len(m.projects) - 1
}

// selectedProjectFromTree returns the project at the current tree cursor
func (m *Model) selectedProjectFromTree() *protocol.FullProject {
	if m.listIdx < 0 || m.listIdx >= len(m.treeLines) {
		return nil
	}
	return m.treeLines[m.listIdx].Project
}

func (m *Model) sortProjects() {
	switch m.sortMode {
	case SortActivity:
		sortBy(m.projects, func(a, b protocol.FullProject) bool {
			return a.LastActivity > b.LastActivity
		})
	case SortProgress:
		sortBy(m.projects, func(a, b protocol.FullProject) bool {
			if a.Progress != b.Progress {
				return a.Progress > b.Progress
			}
			return a.Name < b.Name
		})
	case SortName:
		sortBy(m.projects, func(a, b protocol.FullProject) bool {
			return a.Name < b.Name
		})
	case SortStatus:
		order := map[string]int{"active": 0, "idea": 1, "paused": 2, "done": 3, "abandoned": 4}
		sortBy(m.projects, func(a, b protocol.FullProject) bool {
			oa, ob := order[string(a.Status)], order[string(b.Status)]
			if oa != ob {
				return oa < ob
			}
			return a.LastActivity > b.LastActivity
		})
	}
}

func sortBy(s []protocol.FullProject, less func(a, b protocol.FullProject) bool) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && less(s[j], s[j-1]); j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func (m *Model) selectedProject() *protocol.FullProject {
	if m.viewMode == ViewTree {
		return m.selectedProjectFromTree()
	}
	if m.listIdx < 0 || m.listIdx >= len(m.projects) {
		return nil
	}
	return &m.projects[m.listIdx]
}

func (m *Model) listHeight() int {
	// Must match what View() passes to renderList: bodyH-2
	hh := m.headerHeight()
	h := m.height - hh - 2 - 2
	if h < 3 {
		return 3
	}
	return h
}

// skipDirLines moves cursor past directory-only lines in tree mode
func (m *Model) skipDirLines(direction int) {
	max := len(m.treeLines)
	if max == 0 {
		return
	}
	for attempts := 0; attempts < max; attempts++ {
		if m.listIdx >= 0 && m.listIdx < max && m.treeLines[m.listIdx].IsDir {
			m.listIdx += direction
			if m.listIdx < 0 {
				m.listIdx = max - 1
			}
			if m.listIdx >= max {
				m.listIdx = 0
			}
		} else {
			break
		}
	}
}

func (m *Model) keepListInView() {
	lh := m.listHeight()
	maxIdx := m.maxListIdx()
	if m.listIdx < 0 {
		m.listIdx = 0
	}
	if m.listIdx > maxIdx {
		m.listIdx = maxIdx
	}
	if m.listIdx < m.listScroll {
		m.listScroll = m.listIdx
	}
	if m.listIdx >= m.listScroll+lh {
		m.listScroll = m.listIdx - lh + 1
	}
	if m.listScroll < 0 {
		m.listScroll = 0
	}
}

func (m *Model) mutate(fn func(*protocol.Project)) {
	p := m.selectedProject()
	if p == nil || p.Missing {
		return
	}
	projectID := p.ID // remember ID before re-sort
	proj, err := protocol.ReadProject(p.Path)
	if err != nil {
		return
	}
	fn(proj)
	proj.LastActivity = protocol.Now()
	protocol.WriteProject(p.Path, proj)
	protocol.SyncToRegistry(p.Path, proj)
	all, _ := protocol.LoadAllProjects()
	m.allProjects = all
	m.applyFilterAndSort()
	// Restore selection to same project after re-sort
	if m.viewMode == ViewTree {
		for i, line := range m.treeLines {
			if line.Project != nil && line.Project.ID == projectID {
				m.listIdx = i
				break
			}
		}
	} else {
		for i, pp := range m.projects {
			if pp.ID == projectID {
				m.listIdx = i
				break
			}
		}
	}
	m.keepListInView()
}

func (m *Model) setFlash(msg string) {
	m.flash = msg
	m.flashTicks = 2
}

func (m *Model) confirm(msg string, action func()) {
	m.confirmMsg = msg
	m.confirmAction = action
}

func (m *Model) startInput(mode InputMode, prompt string) {
	m.inputMode = mode
	m.textInput.SetValue("")
	m.textInput.Prompt = prompt + " "
	m.textInput.Focus()
}

func (m *Model) startInputWithValue(mode InputMode, prompt, value string) {
	m.inputMode = mode
	m.textInput.SetValue(value)
	m.textInput.Prompt = prompt + " "
	m.textInput.Focus()
}
