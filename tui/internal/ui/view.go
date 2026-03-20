package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/drift-codes/drift/internal/protocol"
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// Help overlay takes full screen
	if m.focus == FocusHelp {
		return m.renderHelp()
	}

	listW := Clamp(m.width*38/100, 25, 44)
	detailW := m.width - listW - 3
	bodyH := m.height - 3

	header := m.renderHeader()
	list := m.renderList(listW-2, bodyH-2)
	detail := m.renderDetail(detailW-2, bodyH-2)

	// Borders: active panel gets cyan
	lb, db := ListBorderDim, DetailBorderDim
	if m.focus == FocusList {
		lb = ListBorderActive
	} else {
		db = DetailBorderActive
	}
	listBox := lb.Width(listW).Height(bodyH).Render(list)
	detailBox := db.Width(detailW).Height(bodyH).Render(detail)

	body := lipgloss.JoinHorizontal(lipgloss.Top, listBox, detailBox)
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

// ─── Header with breadcrumb ──────────────────────

func (m Model) renderHeader() string {
	// Breadcrumb
	var crumbs []string
	crumbs = append(crumbs, HeaderStyle.Render("drift"))

	if p := m.selectedProject(); p != nil && m.focus == FocusDetail {
		crumbs = append(crumbs, Dim.Render(" › "))
		crumbs = append(crumbs, Accent.Render(p.Name))
		crumbs = append(crumbs, Dim.Render(" › "))
		crumbs = append(crumbs, Dim.Render(SectionLabels[int(m.detailSection)]))
	}
	left := strings.Join(crumbs, "")

	// Count / filter indicator
	countStr := fmt.Sprintf(" %d", len(m.projects))
	if m.filterText != "" {
		countStr += fmt.Sprintf("/%d", len(m.allProjects))
		countStr += Dim.Render(" ⌕ " + m.filterText)
	}
	left += Dim.Render(countStr)

	// Sort indicator (right-aligned)
	var sortParts []string
	for i, label := range SortLabels {
		if SortMode(i) == m.sortMode {
			sortParts = append(sortParts, AccentB.Render(label))
		} else {
			sortParts = append(sortParts, Dim.Render(label))
		}
	}
	right := strings.Join(sortParts, Dim.Render("·"))

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}
	return " " + left + strings.Repeat(" ", gap) + right + " "
}

// ─── List Panel ──────────────────────────────────

func (m Model) renderList(w, h int) string {
	if len(m.projects) == 0 {
		if m.filterText != "" {
			return Dim.Render("  No matches for \"" + m.filterText + "\"\n  Esc to clear")
		}
		return Dim.Render("  No projects\n  Run drift init")
	}

	var rows []string
	end := Clamp(m.listScroll+h, 0, len(m.projects))
	for i := m.listScroll; i < end; i++ {
		p := m.projects[i]
		selected := i == m.listIdx
		dimmed := m.focus != FocusList

		icon := StatusIcons[string(p.Status)]
		color := StatusColors[string(p.Status)]
		pct := fmt.Sprintf("%3d%%", p.Progress)
		ts := TimeSince(p.LastActivity)

		nameW := w - 14
		if nameW < 5 {
			nameW = 5
		}
		name := PadRight(Truncate(p.Name, nameW), nameW)

		if selected && !dimmed {
			row := fmt.Sprintf(" %s %s  %s %s", icon, name, pct, ts)
			rows = append(rows, SelectedRow.Width(w).Render(row))
		} else if dimmed {
			row := fmt.Sprintf(" %s %s  %s %s",
				Dim.Render(icon), Dim.Render(name), Dim.Render(pct), Dim.Render(PadRight(ts, 4)))
			if selected {
				rows = append(rows, DimSelectedRow.Width(w).Render(row))
			} else {
				rows = append(rows, row)
			}
		} else {
			iconS := lipgloss.NewStyle().Foreground(color).Render(icon)
			rows = append(rows, fmt.Sprintf(" %s %s  %s %s",
				iconS, name, Dim.Render(pct), Dim.Render(PadRight(ts, 4))))
		}
	}

	for len(rows) < h {
		rows = append(rows, "")
	}
	return strings.Join(rows, "\n")
}

// ─── Detail Panel ────────────────────────────────

func (m Model) renderDetail(w, h int) string {
	p := m.selectedProject()
	if p == nil {
		return Dim.Render(" No project selected")
	}

	var lines []string
	add := func(s string) {
		lines = append(lines, s)
	}

	focused := m.focus == FocusDetail
	home, _ := os.UserHomeDir()

	// ── INFO (fixed 7 lines) ──
	infoActive := focused && m.detailSection == SectionInfo
	st := StatusColors[string(p.Status)]
	nameS := lipgloss.NewStyle().Foreground(st).Bold(true).Render(p.Name)
	meta := Dim.Render(fmt.Sprintf("  %s  %d%%", string(p.Status), p.Progress))

	marker := "  "
	if infoActive {
		marker = AccentB.Render("▸ ")
	}
	add(marker + nameS + meta) // line 1: name + status

	// line 2: description (always present)
	if p.Description != nil {
		add("    " + Truncate(*p.Description, w-6))
	} else {
		add("    " + Dim.Render("—"))
	}

	// line 3: tags (always present)
	if len(p.Tags) > 0 {
		var tags []string
		for _, t := range p.Tags {
			tags = append(tags, TagStyle.Render("#"+t))
		}
		add("    " + Truncate(strings.Join(tags, " "), w-6))
	} else {
		add("    " + Dim.Render("no tags"))
	}

	// line 4: path
	add("    " + Dim.Render(Truncate(strings.Replace(p.Path, home, "~", 1), w-6)))

	// line 5: updated
	add("    " + Dim.Render("updated "+TimeSince(p.LastActivity)+" ago"))

	// line 6: status selector (always shown when focused, dim placeholder otherwise)
	var statuses []string
	for i, s := range protocol.AllStatuses {
		if p.Status == s {
			if infoActive {
				statuses = append(statuses, AccentB.Render(fmt.Sprintf("[%d]%s", i+1, string(s))))
			} else {
				statuses = append(statuses, Bold.Render(fmt.Sprintf("[%s]", string(s))))
			}
		} else if infoActive {
			statuses = append(statuses, Dim.Render(fmt.Sprintf(" %d %s", i+1, string(s))))
		}
	}
	if len(statuses) > 0 {
		add("    " + strings.Join(statuses, " "))
	} else {
		add("")
	}

	// line 7: separator
	add("")

	// ── GOALS ──
	goalsActive := focused && m.detailSection == SectionGoals
	doneCount := 0
	for _, g := range p.Goals {
		if g.Done {
			doneCount++
		}
	}

	marker = "  "
	if goalsActive {
		marker = AccentB.Render("▸ ")
	}
	add(marker + Bold.Render(fmt.Sprintf("goals %d/%d", doneCount, len(p.Goals))))

	if len(p.Goals) == 0 && focused {
		add("    " + Dim.Render("G to add a goal"))
	}
	for i, g := range p.Goals {
		isCursor := goalsActive && i == m.detailCursor
		var icon, text string
		if g.Done {
			icon = GreenS.Render("✓")
			text = Dim.Render(g.Text)
		} else {
			icon = Dim.Render("○")
			text = g.Text
		}
		if isCursor {
			prefix := AccentB.Render("› ")
			if g.Done {
				add(prefix + "  " + GreenS.Render("✓ ") + Accent.Render(g.Text))
			} else {
				add(prefix + "  " + Dim.Render("○ ") + AccentB.Render(g.Text))
			}
		} else {
			add(fmt.Sprintf("    %s %s", icon, text))
		}
	}
	add("")

	// ── NOTES ──
	notesActive := focused && m.detailSection == SectionNotes
	marker = "  "
	if notesActive {
		marker = AccentB.Render("▸ ")
	}
	add(marker + Bold.Render(fmt.Sprintf("notes %d", len(p.Notes))))

	if len(p.Notes) == 0 && focused {
		add("    " + Dim.Render("N to add a note"))
	}
	remaining := h - len(lines) - 1
	if remaining < 1 {
		remaining = 1
	}
	noteStart := len(p.Notes) - remaining
	if noteStart < 0 {
		noteStart = 0
	}
	for i := len(p.Notes) - 1; i >= noteStart; i-- {
		n := p.Notes[i]
		var ts string
		if len(n.Ts) >= 16 {
			ts = n.Ts[5:10] + " " + n.Ts[11:16]
		} else {
			ts = n.Ts
		}
		text := Truncate(n.Text, w-16)
		add(fmt.Sprintf("    %s %s", Dim.Render(ts), text))
	}
	if noteStart > 0 {
		add(Dim.Render(fmt.Sprintf("    +%d more", noteStart)))
	}

	// ── LINKS ──
	for _, pair := range []struct{ k, v string }{
		{"repo", ptrStr(p.Links.Repo)},
		{"deploy", ptrStr(p.Links.Deploy)},
		{"design", ptrStr(p.Links.Design)},
	} {
		if pair.v != "" {
			add(fmt.Sprintf("    %s %s", Dim.Render(pair.k), Accent.Render(pair.v)))
		}
	}

	// Fill to height
	for len(lines) < h {
		lines = append(lines, "")
	}
	if len(lines) > h {
		lines = lines[:h]
	}

	return strings.Join(lines, "\n")
}

// ─── Footer ──────────────────────────────────────

func (m Model) renderFooter() string {
	// Confirm dialog
	if m.confirmMsg != "" {
		return " " + YellowS.Render(m.confirmMsg)
	}

	// Flash message
	if m.flash != "" {
		return " " + FlashStyle.Render(m.flash)
	}

	// Text input
	if m.inputMode != InputNone {
		return " " + m.textInput.View()
	}

	// Context-sensitive help
	kv := func(k, v string) string {
		return HelpKey.Render(k) + HelpLabel.Render(" "+v)
	}

	var parts []string
	switch m.focus {
	case FocusList:
		parts = append(parts, kv("↑↓", "nav"), kv("⏎", "open"), kv("s", "sort"), kv("/", "filter"), kv(":", "jump"))
		parts = append(parts, kv("?", "help"), kv("q", "quit"))
	case FocusDetail:
		switch m.detailSection {
		case SectionInfo:
			parts = append(parts, kv("↑↓", "nav"), kv("tab", "section"), kv("1-5", "status"), kv("D", "describe"))
			parts = append(parts, kv("[/]", "prev/next"), kv("?", "help"), kv("esc", "back"))
		case SectionGoals:
			parts = append(parts, kv("↑↓", "nav"), kv("⏎", "toggle"), kv("g", "add"), kv("x", "delete"))
			parts = append(parts, kv("tab", "section"), kv("[/]", "prev/next"), kv("esc", "back"))
		case SectionNotes:
			parts = append(parts, kv("n", "add note"), kv("tab", "section"), kv("[/]", "prev/next"), kv("esc", "back"))
		}
	}

	return " " + strings.Join(parts, "  ")
}

// ─── Help Screen ─────────────────────────────────

func (m Model) renderHelp() string {
	title := AccentB.Render("  drift — keyboard shortcuts\n")

	sections := []struct {
		header string
		keys   [][2]string
	}{
		{"Navigation", [][2]string{
			{"↑/↓ or j/k", "Move up/down"},
			{"Enter or →", "Open project detail"},
			{"Esc or ←", "Back to list"},
			{"Tab", "Next section in detail"},
			{"Shift+Tab", "Previous section"},
			{"[ / ]", "Previous/next project (in detail)"},
			{"g / G", "First / last project"},
			{"Ctrl+U/D", "Page up / page down"},
		}},
		{"Actions", [][2]string{
			{"n", "Add note"},
			{"g", "Add goal"},
			{"D", "Edit description"},
			{"Space/Enter", "Toggle goal done (on goal)"},
			{"x", "Delete goal (on goal, with confirm)"},
			{"1-5", "Set status: 1=active 2=idea 3=paused 4=done 5=abandoned"},
		}},
		{"Search & Sort", [][2]string{
			{"/", "Filter projects by name (live search)"},
			{":", "Jump to project by name"},
			{"s", "Cycle sort mode (recent → progress → name → status)"},
			{"Esc", "Clear filter"},
		}},
		{"General", [][2]string{
			{"?", "Toggle this help"},
			{"q", "Quit (from list)"},
			{"Ctrl+C", "Force quit"},
		}},
	}

	var out strings.Builder
	out.WriteString(title)
	out.WriteString("\n")

	for _, sec := range sections {
		out.WriteString("  " + Bold.Render(sec.header) + "\n")
		for _, kv := range sec.keys {
			key := HelpKey.Render(PadRight(kv[0], 16))
			out.WriteString("    " + key + " " + kv[1] + "\n")
		}
		out.WriteString("\n")
	}

	out.WriteString(Dim.Render("  Press ? or Esc to close"))

	// Center vertically
	content := out.String()
	lines := strings.Split(content, "\n")
	padTop := (m.height - len(lines)) / 3
	if padTop < 0 {
		padTop = 0
	}

	var result strings.Builder
	for i := 0; i < padTop; i++ {
		result.WriteString("\n")
	}
	result.WriteString(content)

	// Fill remaining height
	total := padTop + len(lines)
	for total < m.height {
		result.WriteString("\n")
		total++
	}

	return result.String()
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
