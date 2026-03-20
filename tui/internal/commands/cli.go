package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/drift-codes/drift/internal/protocol"
	"github.com/drift-codes/drift/internal/ui"
)

var (
	cyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	green  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	red    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	dim    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	bold   = lipgloss.NewStyle().Bold(true)
	cyanB  = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
)

func statusIcon(status string) string {
	switch status {
	case "active":
		return cyan.Render("●")
	case "done":
		return green.Render("✓")
	case "idea":
		return yellow.Render("○")
	case "paused":
		return dim.Render("◇")
	case "abandoned":
		return dim.Render("✗")
	}
	return "?"
}

func miniBar(pct int) string {
	w := 5
	filled := pct * w / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", w-filled)
	if pct >= 100 {
		return green.Render(bar)
	}
	if pct >= 60 {
		return cyan.Render(bar)
	}
	if pct >= 30 {
		return yellow.Render(bar)
	}
	return dim.Render(bar)
}

func cwd() string {
	d, _ := os.Getwd()
	return d
}

func needsProject() *protocol.Project {
	p, err := protocol.ReadProject(cwd())
	if err != nil {
		fmt.Println(red.Render("  No .drift/ — run: drift init"))
		return nil
	}
	return p
}

func addToGitignore(dir string) {
	gi := filepath.Join(dir, ".gitignore")
	data, err := os.ReadFile(gi)
	if err != nil {
		return
	}
	if !strings.Contains(string(data), ".drift/") {
		os.WriteFile(gi, []byte(strings.TrimRight(string(data), "\n")+"\n.drift/\n"), 0644)
	}
}

// ─── Commands ────────────────────────────────────

func Init(dir string) {
	root := dir
	if root == "" {
		root = cwd()
	} else {
		root, _ = filepath.Abs(root)
	}

	if protocol.HasProject(root) {
		fmt.Println(yellow.Render("  Already initialized."))
		Status()
		return
	}

	p := protocol.CreateProject(root)
	p.Tags = protocol.DetectTags(root)
	repo := protocol.DetectRepo(root)
	if repo != "" {
		p.Links.Repo = &repo
	}

	protocol.WriteProject(root, p)
	protocol.SyncToRegistry(root, p)
	addToGitignore(root)

	fmt.Printf("\n  %s %s — %s\n", green.Render("✓"), bold.Render("drift init"), cyan.Render(p.Name))
	fmt.Printf("  Status: %s | Progress: %d%%\n", p.Status, p.Progress)
	if len(p.Tags) > 0 {
		fmt.Printf("  Tags: %s\n", strings.Join(p.Tags, ", "))
	}
	if p.Links.Repo != nil {
		fmt.Printf("  Repo: %s\n", *p.Links.Repo)
	}
	fmt.Println()
}

func Status() {
	p := needsProject()
	if p == nil {
		return
	}
	doneGoals := 0
	for _, g := range p.Goals {
		if g.Done {
			doneGoals++
		}
	}

	desc := dim.Render("—")
	if p.Description != nil {
		desc = *p.Description
	}

	fmt.Printf("\n  📂 %s %s %d%%\n", bold.Render(p.Name), dim.Render("["+string(p.Status)+"]"), p.Progress)
	fmt.Printf("  %s\n", desc)
	if len(p.Tags) > 0 {
		var tags []string
		for _, t := range p.Tags {
			tags = append(tags, cyan.Render(t))
		}
		fmt.Printf("  Tags: %s\n", strings.Join(tags, "  "))
	}
	fmt.Printf("  Last: %s\n", ui.TimeSince(p.LastActivity))

	if len(p.Goals) > 0 {
		fmt.Printf("\n  Goals %s\n", dim.Render(fmt.Sprintf("%d/%d", doneGoals, len(p.Goals))))
		for i, g := range p.Goals {
			icon := dim.Render("○")
			text := g.Text
			if g.Done {
				icon = green.Render("✓")
				text = dim.Render(g.Text)
			}
			fmt.Printf("  %s %s %s\n", icon, dim.Render(fmt.Sprintf("%d.", i+1)), text)
		}
	}

	if len(p.Notes) > 0 {
		fmt.Println("\n  Notes")
		n := len(p.Notes)
		start := n - 5
		if start < 0 {
			start = 0
		}
		for i := n - 1; i >= start; i-- {
			note := p.Notes[i]
			ts := ""
			if len(note.Ts) >= 16 {
				ts = note.Ts[11:16]
			}
			fmt.Printf("  %s  %s\n", dim.Render(ts), note.Text)
		}
		if start > 0 {
			fmt.Printf("  %s\n", dim.Render(fmt.Sprintf("+%d more", start)))
		}
	}

	for _, pair := range []struct{ k string; v *string }{
		{"repo", p.Links.Repo}, {"deploy", p.Links.Deploy}, {"design", p.Links.Design},
	} {
		if pair.v != nil {
			fmt.Printf("  %s %s\n", dim.Render(pair.k+":"), cyan.Render(*pair.v))
		}
	}
	fmt.Println()
}

func Note(text string) {
	p := needsProject()
	if p == nil {
		return
	}
	p.Notes = append(p.Notes, protocol.Note{Ts: protocol.Now(), Text: text})
	p.LastActivity = protocol.Now()
	protocol.WriteProject(cwd(), p)
	protocol.SyncToRegistry(cwd(), p)
	fmt.Printf("  %s Note added to %s\n", green.Render("✓"), cyan.Render(p.Name))
}

func Goal(text string) {
	p := needsProject()
	if p == nil {
		return
	}
	p.Goals = append(p.Goals, protocol.Goal{Text: text, Done: false})
	p.Progress = protocol.CalcProgress(p.Goals)
	p.LastActivity = protocol.Now()
	protocol.WriteProject(cwd(), p)
	protocol.SyncToRegistry(cwd(), p)
	fmt.Printf("  %s Goal added to %s\n", green.Render("✓"), cyan.Render(p.Name))
	printGoals(p)
}

func GoalDone(nStr string) {
	n, err := strconv.Atoi(nStr)
	if err != nil {
		fmt.Println("  Usage: drift goal done N")
		return
	}
	p := needsProject()
	if p == nil {
		return
	}
	if n < 1 || n > len(p.Goals) {
		fmt.Printf("  %s\n", red.Render(fmt.Sprintf("No goal #%d", n)))
		return
	}
	p.Goals[n-1].Done = true
	p.Progress = protocol.CalcProgress(p.Goals)
	p.LastActivity = protocol.Now()
	protocol.WriteProject(cwd(), p)
	protocol.SyncToRegistry(cwd(), p)
	fmt.Printf("  %s Goal #%d done!\n", green.Render("✓"), n)
	printGoals(p)
}

func Progress(nStr string) {
	n, err := strconv.Atoi(nStr)
	if err != nil {
		fmt.Println("  Usage: drift progress N")
		return
	}
	p := needsProject()
	if p == nil {
		return
	}
	if n < 0 {
		n = 0
	}
	if n > 100 {
		n = 100
	}
	p.Progress = n
	p.LastActivity = protocol.Now()
	protocol.WriteProject(cwd(), p)
	protocol.SyncToRegistry(cwd(), p)
	fmt.Printf("  %s Progress: %s %d%%\n", green.Render("✓"), miniBar(n), n)
}

func SetStatus(status string) {
	valid := map[string]bool{"active": true, "idea": true, "paused": true, "done": true, "abandoned": true}
	if !valid[status] {
		fmt.Printf("  %s\n", red.Render("Valid statuses: active, idea, paused, done, abandoned"))
		return
	}
	p := needsProject()
	if p == nil {
		return
	}
	p.Status = protocol.ProjectStatus(status)
	p.LastActivity = protocol.Now()
	protocol.WriteProject(cwd(), p)
	protocol.SyncToRegistry(cwd(), p)
	fmt.Printf("  %s Status: %s %s\n", green.Render("✓"), statusIcon(status), status)
}

func Describe(text string) {
	p := needsProject()
	if p == nil {
		return
	}
	if text == "" {
		p.Description = nil
	} else {
		p.Description = &text
	}
	p.LastActivity = protocol.Now()
	protocol.WriteProject(cwd(), p)
	protocol.SyncToRegistry(cwd(), p)
	fmt.Printf("  %s Description updated\n", green.Render("✓"))
}

func Tag(tags []string) {
	p := needsProject()
	if p == nil {
		return
	}
	for _, tag := range tags {
		found := false
		for _, t := range p.Tags {
			if t == tag {
				found = true
				break
			}
		}
		if !found {
			p.Tags = append(p.Tags, tag)
		}
	}
	p.LastActivity = protocol.Now()
	protocol.WriteProject(cwd(), p)
	protocol.SyncToRegistry(cwd(), p)
	var display []string
	for _, t := range p.Tags {
		display = append(display, cyan.Render(t))
	}
	fmt.Printf("  %s Tags: %s\n", green.Render("✓"), strings.Join(display, "  "))
}

func Link(linkType, url string) {
	p := needsProject()
	if p == nil {
		return
	}
	switch linkType {
	case "repo":
		p.Links.Repo = &url
	case "deploy":
		p.Links.Deploy = &url
	case "design":
		p.Links.Design = &url
	default:
		fmt.Printf("  %s\n", red.Render("Valid link types: repo, deploy, design"))
		return
	}
	p.LastActivity = protocol.Now()
	protocol.WriteProject(cwd(), p)
	protocol.SyncToRegistry(cwd(), p)
	fmt.Printf("  %s %s: %s\n", green.Render("✓"), linkType, cyan.Render(url))
}

func List(sortMode string) {
	projects, _ := protocol.LoadAllProjects()
	if len(projects) == 0 {
		fmt.Println(dim.Render("\n  No projects tracked yet."))
		fmt.Printf("  Run %s in a project directory.\n\n", cyan.Render("drift init"))
		return
	}

	// Sort
	switch sortMode {
	case "progress":
		sortProjects(projects, func(a, b protocol.FullProject) bool {
			if a.Progress != b.Progress {
				return a.Progress > b.Progress
			}
			return a.Name < b.Name
		})
	case "name":
		sortProjects(projects, func(a, b protocol.FullProject) bool { return a.Name < b.Name })
	case "status":
		order := map[string]int{"active": 0, "idea": 1, "paused": 2, "done": 3, "abandoned": 4}
		sortProjects(projects, func(a, b protocol.FullProject) bool {
			oa, ob := order[string(a.Status)], order[string(b.Status)]
			if oa != ob {
				return oa < ob
			}
			return a.LastActivity > b.LastActivity
		})
	default:
		// activity — already sorted
	}

	fmt.Printf("\n%s  %s\n\n", cyanB.Render("  drift"), dim.Render(fmt.Sprintf("— %d projects — sort: %s", len(projects), sortMode)))
	home, _ := os.UserHomeDir()
	for _, p := range projects {
		icon := statusIcon(string(p.Status))

		// Show parent dir for disambiguation
		rel, _ := filepath.Rel(home, p.Path)
		dir := filepath.Dir(rel)
		dirPrefix := ""
		if dir != "." && dir != "" {
			dirPrefix = filepath.Base(dir) + "/"
		}

		fullName := dirPrefix + p.Name
		if len(fullName) > 28 {
			fullName = fullName[:27] + "…"
		}

		// Render: dir part dim, name bright
		var nameDisplay string
		if dirPrefix != "" {
			nameDisplay = dim.Render(dirPrefix) + p.Name
			padLen := 28 - len(dirPrefix) - len(p.Name)
			if padLen > 0 {
				nameDisplay += strings.Repeat(" ", padLen)
			}
		} else {
			name := p.Name
			for len(name) < 28 {
				name += " "
			}
			nameDisplay = name
		}

		pct := fmt.Sprintf("%3d%%", p.Progress)
		bar := miniBar(p.Progress)
		ts := ui.TimeSince(p.LastActivity)
		miss := ""
		if p.Missing {
			miss = red.Render(" [missing]")
		}
		fmt.Printf("  %s %s %s %s  %s%s\n", icon, nameDisplay, bar, pct, dim.Render(ts), miss)
	}
	fmt.Println()
}

func Open(name string) {
	r, _ := protocol.ReadRegistry()
	if r == nil {
		return
	}
	var matches []protocol.RegistryEntry
	for _, p := range r.Projects {
		if strings.Contains(strings.ToLower(p.Name), strings.ToLower(name)) {
			matches = append(matches, p)
		}
	}
	if len(matches) == 0 {
		fmt.Printf("  %s\n", red.Render(fmt.Sprintf("Project \"%s\" not found", name)))
		return
	}
	if len(matches) == 1 {
		fmt.Println(matches[0].Path)
		return
	}
	fmt.Println(dim.Render("  Multiple matches:"))
	for _, m := range matches {
		fmt.Printf("  %-20s %s\n", m.Name, dim.Render(m.Path))
	}
}

func Scan(dir string, doInit bool, maxDepth int) {
	root := dir
	if root == "" {
		root = cwd()
	} else {
		root, _ = filepath.Abs(root)
	}

	results := protocol.ScanDir(root, maxDepth)

	home, _ := os.UserHomeDir()
	shortRoot := strings.Replace(root, home, "~", 1)

	if len(results) == 0 {
		fmt.Printf("\n%s\n\n", dim.Render("  No new projects found in "+shortRoot))
		return
	}

	fmt.Printf("\n  %s — found %s new projects in %s\n\n",
		cyanB.Render("drift scan"), bold.Render(strconv.Itoa(len(results))), dim.Render(shortRoot))

	for _, f := range results {
		// Show path relative to scan root
		rel, _ := filepath.Rel(root, f.Path)
		if rel == "" {
			rel = filepath.Base(f.Path)
		}
		for len(rel) < 30 {
			rel += " "
		}
		tags := dim.Render("—")
		if len(f.Tags) > 0 {
			tags = dim.Render(strings.Join(f.Tags, ", "))
		}
		fmt.Printf("  %s %s\n", rel, tags)
	}

	if !doInit {
		fmt.Printf(dim.Render("\n  Run: drift scan --init [dir]\n\n"))
		return
	}

	fmt.Printf(cyan.Render("\n  Initializing all...\n\n"))
	for _, f := range results {
		p := protocol.CreateProject(f.Path)
		p.Tags = f.Tags
		repo := protocol.DetectRepo(f.Path)
		if repo != "" {
			p.Links.Repo = &repo
		}
		protocol.WriteProject(f.Path, p)
		protocol.SyncToRegistry(f.Path, p)
		addToGitignore(f.Path)
		rel, _ := filepath.Rel(root, f.Path)
		fmt.Printf("  %s %s\n", green.Render("✓"), rel)
	}
	fmt.Println()
}

func Help() {
	fmt.Printf(`
%s
%s

%s
  %s                 Interactive TUI (fullscreen)
  %s              Initialize project in current dir
  %s             Show current project status
  %s                List all projects
  %s %s          Add a note
  %s %s          Add a goal
  %s %s         Mark goal #N done
  %s %s          Set progress (0-100)
  %s %s   Change status
  %s %s     Set description
  %s %s       Add tags
  %s %s       Set a link
  %s %s          Find untracked projects
  %s         Find & init all
  %s %s           Get project path

%s
`,
		cyanB.Render("  drift"),
		dim.Render("  Vibe-coding project manager"),
		bold.Render("Commands:"),
		cyan.Render("drift"),
		cyan.Render("drift init"),
		cyan.Render("drift status"),
		cyan.Render("drift list"),
		cyan.Render("drift note"), dim.Render("\"text\""),
		cyan.Render("drift goal"), dim.Render("\"text\""),
		cyan.Render("drift goal done"), dim.Render("N"),
		cyan.Render("drift progress"), dim.Render("N"),
		cyan.Render("drift set-status"), dim.Render("STATUS"),
		cyan.Render("drift describe"), dim.Render("\"text\""),
		cyan.Render("drift tag"), dim.Render("tag1 tag2"),
		cyan.Render("drift link"), dim.Render("type url"),
		cyan.Render("drift scan"), dim.Render("[dir]"),
		cyan.Render("drift scan --init"),
		cyan.Render("drift open"), dim.Render("name"),
		dim.Render("  Statuses: active, idea, paused, done, abandoned"),
	)
}

// helpers

func printGoals(p *protocol.Project) {
	for i, g := range p.Goals {
		icon := dim.Render("○")
		text := g.Text
		if g.Done {
			icon = green.Render("✓")
			text = dim.Render(g.Text)
		}
		fmt.Printf("  %s %s %s\n", icon, dim.Render(fmt.Sprintf("%d.", i+1)), text)
	}
	fmt.Printf("\n  Progress: %s %d%%\n", miniBar(p.Progress), p.Progress)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func sortProjects(s []protocol.FullProject, less func(a, b protocol.FullProject) bool) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && less(s[j], s[j-1]); j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
