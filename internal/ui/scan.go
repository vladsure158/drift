package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/snowtema/drift/internal/protocol"
)

// ─── Types ───────────────────────────────────────

type scanResultEntry struct {
	Result   protocol.ScanResult
	Selected bool
}

// ─── Commands ────────────────────────────────────

func scanCmd(root string) tea.Cmd {
	return func() tea.Msg {
		results := protocol.ScanDir(root, 3)
		return scanDoneMsg{results: results, root: root}
	}
}

// ─── Helpers ─────────────────────────────────────

func listSubDirs(path string) []string {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}
	skipDirs := map[string]bool{
		"node_modules": true, ".next": true, "__pycache__": true,
		"dist": true, "build": true, "target": true, ".venv": true, "venv": true,
		".cache": true, "vendor": true,
	}
	var dirs []string
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || skipDirs[e.Name()] {
			continue
		}
		if e.Type()&os.ModeSymlink != 0 {
			continue
		}
		dirs = append(dirs, e.Name())
	}
	sort.Strings(dirs)
	return dirs
}

func shortPath(path string) string {
	home, _ := os.UserHomeDir()
	return strings.Replace(path, home, "~", 1)
}

// ─── State transitions ──────────────────────────

func (m *Model) enterScanBrowse() {
	home, _ := os.UserHomeDir()
	m.focus = FocusScanBrowse
	m.scanBrowsePath = home
	m.scanBrowseDirs = listSubDirs(home)
	m.scanBrowseIdx = 0
	m.scanBrowseScroll = 0
	m.scanScanning = false
}

func (m *Model) scanBrowseEnter(dir string) {
	m.scanBrowsePath = dir
	m.scanBrowseDirs = listSubDirs(dir)
	m.scanBrowseIdx = 0
	m.scanBrowseScroll = 0
}

// Layout constants for scan views.
// Browse: empty + title + empty + path + empty = 5 header lines, 1 footer.
// Results: empty + title + subtitle + empty = 4 header lines, 1 footer.
const (
	scanBrowseHeaderLines  = 5
	scanBrowseFooterLines  = 1
	scanResultsHeaderLines = 4
	scanResultsFooterLines = 1
)

func (m *Model) scanBrowseListHeight() int {
	h := m.height - scanBrowseHeaderLines - scanBrowseFooterLines
	if h < 1 {
		return 1
	}
	return h
}

func (m *Model) scanResultsListHeight() int {
	h := m.height - scanResultsHeaderLines - scanResultsFooterLines
	if h < 1 {
		return 1
	}
	return h
}

func keepInView(idx, scroll, listH int) int {
	if idx < scroll {
		scroll = idx
	}
	if idx >= scroll+listH {
		scroll = idx - listH + 1
	}
	if scroll < 0 {
		scroll = 0
	}
	return scroll
}

// ─── Update: Scan Browse ─────────────────────────

func (m Model) updateScanBrowse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.scanScanning {
		if msg.String() == "esc" {
			m.scanScanning = false
			m.focus = FocusList
		}
		return m, nil
	}

	maxIdx := len(m.scanBrowseDirs) - 1

	switch msg.String() {
	case "esc", "q":
		m.focus = FocusList
		return m, nil

	case "up", "k":
		if m.scanBrowseIdx > 0 {
			m.scanBrowseIdx--
		}

	case "down", "j":
		if m.scanBrowseIdx < maxIdx {
			m.scanBrowseIdx++
		}

	case "home", "g":
		m.scanBrowseIdx = 0
		m.scanBrowseScroll = 0

	case "end", "G":
		if maxIdx >= 0 {
			m.scanBrowseIdx = maxIdx
		}

	case "pgup", "ctrl+u":
		m.scanBrowseIdx = Clamp(m.scanBrowseIdx-m.scanBrowseListHeight(), 0, maxIdx)

	case "pgdown", "ctrl+d":
		m.scanBrowseIdx = Clamp(m.scanBrowseIdx+m.scanBrowseListHeight(), 0, maxIdx)

	case "enter", "right", "l":
		if m.scanBrowseIdx >= 0 && m.scanBrowseIdx < len(m.scanBrowseDirs) {
			newPath := filepath.Join(m.scanBrowsePath, m.scanBrowseDirs[m.scanBrowseIdx])
			m.scanBrowseEnter(newPath)
		}
		return m, nil

	case "left", "h", "backspace":
		parent := filepath.Dir(m.scanBrowsePath)
		if parent != m.scanBrowsePath {
			oldName := filepath.Base(m.scanBrowsePath)
			m.scanBrowseEnter(parent)
			for i, d := range m.scanBrowseDirs {
				if d == oldName {
					m.scanBrowseIdx = i
					break
				}
			}
		}
		return m, nil

	case "s", "S":
		m.scanScanning = true
		return m, scanCmd(m.scanBrowsePath)
	}

	m.scanBrowseScroll = keepInView(m.scanBrowseIdx, m.scanBrowseScroll, m.scanBrowseListHeight())
	return m, nil
}

// ─── Update: Scan Results ────────────────────────

func (m Model) updateScanResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxIdx := len(m.scanResults) - 1

	switch msg.String() {
	case "esc", "q":
		m.focus = FocusScanBrowse
		m.scanScanning = false
		return m, nil

	case "up", "k":
		if m.scanResultsIdx > 0 {
			m.scanResultsIdx--
		}

	case "down", "j":
		if m.scanResultsIdx < maxIdx {
			m.scanResultsIdx++
		}

	case "home", "g":
		m.scanResultsIdx = 0
		m.scanResultsScroll = 0

	case "end", "G":
		if maxIdx >= 0 {
			m.scanResultsIdx = maxIdx
		}

	case "pgup", "ctrl+u":
		m.scanResultsIdx = Clamp(m.scanResultsIdx-m.scanResultsListHeight(), 0, maxIdx)

	case "pgdown", "ctrl+d":
		m.scanResultsIdx = Clamp(m.scanResultsIdx+m.scanResultsListHeight(), 0, maxIdx)

	case " ":
		if m.scanResultsIdx >= 0 && m.scanResultsIdx < len(m.scanResults) {
			m.scanResults[m.scanResultsIdx].Selected = !m.scanResults[m.scanResultsIdx].Selected
			if m.scanResultsIdx < maxIdx {
				m.scanResultsIdx++
			}
		}

	case "a":
		for i := range m.scanResults {
			m.scanResults[i].Selected = true
		}

	case "n":
		for i := range m.scanResults {
			m.scanResults[i].Selected = false
		}

	case "enter", "i":
		count := m.initSelectedProjects()
		if count > 0 {
			m.focus = FocusList
			m.setFlash(fmt.Sprintf("Initialized %d projects", count))
			return m, tea.Batch(loadProjectsCmd, FlashCmd())
		}
		m.setFlash("No projects selected")
		return m, FlashCmd()
	}

	m.scanResultsScroll = keepInView(m.scanResultsIdx, m.scanResultsScroll, m.scanResultsListHeight())
	return m, nil
}

func (m *Model) initSelectedProjects() int {
	count := 0
	for _, r := range m.scanResults {
		if !r.Selected {
			continue
		}
		p := protocol.CreateProject(r.Result.Path)
		p.Tags = r.Result.Tags
		repo := protocol.DetectRepo(r.Result.Path)
		if repo != "" {
			p.Links.Repo = &repo
		}
		protocol.WriteProject(r.Result.Path, p)
		protocol.SyncToRegistry(r.Result.Path, p)
		protocol.AddToGitignore(r.Result.Path)
		protocol.AddClaudeMD(r.Result.Path)
		count++
	}
	return count
}

// ─── View: Scan Browse ──────────────────────────

func (m Model) renderScanBrowse() string {
	w := m.width
	displayPath := shortPath(m.scanBrowsePath)

	// Build exactly m.height lines
	lines := make([]string, 0, m.height)

	// Header: 5 lines
	lines = append(lines, "")
	lines = append(lines, "  "+AccentB.Render("SCAN")+Dim.Render(" — select directory to scan"))
	lines = append(lines, "")
	lines = append(lines, "  "+Bold.Render(Truncate(displayPath+"/", w-4)))
	lines = append(lines, "")

	listH := m.scanBrowseListHeight()

	if m.scanScanning {
		lines = append(lines, "  "+Accent.Render("Scanning..."))
		listH--
	} else if len(m.scanBrowseDirs) == 0 {
		lines = append(lines, "  "+Dim.Render("(empty)"))
		listH--
	} else {
		end := m.scanBrowseScroll + listH
		if end > len(m.scanBrowseDirs) {
			end = len(m.scanBrowseDirs)
		}
		nameW := w - 6 // "    " indent + "/" suffix margin
		if nameW < 5 {
			nameW = 5
		}
		for i := m.scanBrowseScroll; i < end; i++ {
			name := Truncate(m.scanBrowseDirs[i]+"/", nameW)
			if i == m.scanBrowseIdx {
				row := "  > " + name
				lines = append(lines, SelectedRow.Width(w).Render(row))
			} else {
				lines = append(lines, "    "+Dim.Render(name))
			}
		}
		rendered := end - m.scanBrowseScroll
		listH -= rendered
	}

	// Pad remaining list area
	for i := 0; i < listH; i++ {
		lines = append(lines, "")
	}

	// Trim to exactly m.height-1 (leave 1 for footer)
	target := m.height - scanBrowseFooterLines
	for len(lines) < target {
		lines = append(lines, "")
	}
	if len(lines) > target {
		lines = lines[:target]
	}

	// Footer
	kv := func(k, v string) string {
		return HelpKey.Render(k) + HelpLabel.Render(" "+v)
	}
	if m.scanScanning {
		lines = append(lines, " "+Dim.Render("scanning...")+"  "+kv("esc", "cancel"))
	} else {
		parts := []string{kv("↑↓", "nav"), kv("⏎", "enter dir"), kv("h", "parent"), kv("s", "scan here"), kv("esc", "cancel")}
		lines = append(lines, " "+strings.Join(parts, "  "))
	}

	return strings.Join(lines, "\n")
}

// ─── View: Scan Results ─────────────────────────

func (m Model) renderScanResults() string {
	w := m.width
	displayRoot := shortPath(m.scanRoot)

	selectedCount := 0
	for _, r := range m.scanResults {
		if r.Selected {
			selectedCount++
		}
	}

	// Build exactly m.height lines
	lines := make([]string, 0, m.height)

	// Header: 4 lines
	lines = append(lines, "")
	lines = append(lines, "  "+AccentB.Render("SCAN RESULTS")+Dim.Render(" — ")+Bold.Render(Truncate(displayRoot, w-20)))
	if len(m.scanResults) == 0 {
		lines = append(lines, "  "+Dim.Render("No new projects found"))
	} else {
		lines = append(lines, "  "+Dim.Render(fmt.Sprintf("Found %d new projects, %d selected", len(m.scanResults), selectedCount)))
	}
	lines = append(lines, "")

	listH := m.scanResultsListHeight()

	if len(m.scanResults) > 0 {
		end := m.scanResultsScroll + listH
		if end > len(m.scanResults) {
			end = len(m.scanResults)
		}

		// Row layout: "  [+] name          tags"
		// checkbox = 3 chars, spaces around = 4 total prefix "  [+] "
		prefixW := 6
		availW := w - prefixW - 2 // 2 for right margin

		for i := m.scanResultsScroll; i < end; i++ {
			r := m.scanResults[i]

			rel, _ := filepath.Rel(m.scanRoot, r.Result.Path)
			if rel == "" {
				rel = filepath.Base(r.Result.Path)
			}

			tagsRaw := ""
			if len(r.Result.Tags) > 0 {
				tagsRaw = strings.Join(r.Result.Tags, ", ")
			}

			// Split available width between name and tags
			tagsVisW := len(tagsRaw)
			if tagsVisW > 0 {
				tagsVisW += 2 // "  " gap
			}
			nameW := availW - tagsVisW
			if nameW < 12 {
				nameW = 12
				tagsVisW = availW - nameW
				if tagsVisW < 3 {
					tagsRaw = ""
					tagsVisW = 0
				} else {
					tagsRaw = Truncate(tagsRaw, tagsVisW-2)
				}
			}

			namePad := PadRight(Truncate(rel, nameW), nameW)

			if i == m.scanResultsIdx {
				// Selected row: plain text, styled as one block
				cb := "[+]"
				if !r.Selected {
					cb = "[ ]"
				}
				row := "  " + cb + " " + namePad
				if tagsRaw != "" {
					row += "  " + tagsRaw
				}
				lines = append(lines, SelectedRow.Width(w).Render(row))
			} else {
				cb := Dim.Render("[ ]")
				if r.Selected {
					cb = GreenS.Render("[+]")
				}
				row := "  " + cb + " " + namePad
				if tagsRaw != "" {
					row += "  " + Dim.Render(tagsRaw)
				}
				lines = append(lines, row)
			}
		}

		rendered := end - m.scanResultsScroll
		for i := 0; i < listH-rendered; i++ {
			lines = append(lines, "")
		}
	} else {
		for i := 0; i < listH; i++ {
			lines = append(lines, "")
		}
	}

	// Ensure exactly m.height - footerLines
	target := m.height - scanResultsFooterLines
	for len(lines) < target {
		lines = append(lines, "")
	}
	if len(lines) > target {
		lines = lines[:target]
	}

	// Footer
	kv := func(k, v string) string {
		return HelpKey.Render(k) + HelpLabel.Render(" "+v)
	}
	if len(m.scanResults) > 0 {
		parts := []string{
			kv("↑↓", "nav"),
			kv("space", "toggle"),
			kv("a", "all"),
			kv("n", "none"),
			kv("⏎", fmt.Sprintf("init %d", selectedCount)),
			kv("esc", "back"),
		}
		lines = append(lines, " "+strings.Join(parts, "  "))
	} else {
		lines = append(lines, " "+kv("esc", "back"))
	}

	return strings.Join(lines, "\n")
}
