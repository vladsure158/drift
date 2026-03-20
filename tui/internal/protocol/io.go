package protocol

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
)

func RegistryDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".drift")
}

func RegistryPath() string {
	return filepath.Join(RegistryDir(), "registry.json")
}

func ProjectFilePath(projectRoot string) string {
	return filepath.Join(projectRoot, ".drift", "project.json")
}

func HasProject(projectRoot string) bool {
	_, err := os.Stat(ProjectFilePath(projectRoot))
	return err == nil
}

func ReadProject(projectRoot string) (*Project, error) {
	data, err := os.ReadFile(ProjectFilePath(projectRoot))
	if err != nil {
		return nil, err
	}
	var p Project
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	if p.Tags == nil {
		p.Tags = []string{}
	}
	if p.Goals == nil {
		p.Goals = []Goal{}
	}
	if p.Notes == nil {
		p.Notes = []Note{}
	}
	return &p, nil
}

func WriteProject(projectRoot string, p *Project) error {
	dir := filepath.Join(projectRoot, ".drift")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ProjectFilePath(projectRoot), append(data, '\n'), 0644)
}

func ReadRegistry() (*Registry, error) {
	data, err := os.ReadFile(RegistryPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{Version: 1, Projects: []RegistryEntry{}}, nil
		}
		return nil, err
	}
	var r Registry
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func WriteRegistry(r *Registry) error {
	if err := os.MkdirAll(RegistryDir(), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(RegistryPath(), append(data, '\n'), 0644)
}

func SyncToRegistry(projectRoot string, p *Project) error {
	r, err := ReadRegistry()
	if err != nil {
		return err
	}
	entry := RegistryEntry{
		ID:           p.ID,
		Path:         projectRoot,
		Name:         p.Name,
		Status:       string(p.Status),
		LastActivity: p.LastActivity,
	}
	found := false
	for i, e := range r.Projects {
		if e.ID == p.ID {
			r.Projects[i] = entry
			found = true
			break
		}
	}
	if !found {
		r.Projects = append(r.Projects, entry)
	}
	return WriteRegistry(r)
}

func CreateProject(projectRoot string) *Project {
	return &Project{
		ID:           uuid.New().String(),
		Name:         filepath.Base(projectRoot),
		Status:       StatusActive,
		Progress:     0,
		Tags:         []string{},
		Created:      Now(),
		LastActivity: Now(),
		Goals:        []Goal{},
		Notes:        []Note{},
		Links:        Links{},
	}
}

func LoadAllProjects() ([]FullProject, error) {
	r, err := ReadRegistry()
	if err != nil {
		return nil, err
	}
	var projects []FullProject
	for _, entry := range r.Projects {
		p, err := ReadProject(entry.Path)
		if err != nil {
			projects = append(projects, FullProject{
				Project: Project{
					ID:           entry.ID,
					Name:         entry.Name,
					Status:       ProjectStatus(entry.Status),
					LastActivity: entry.LastActivity,
					Tags:         []string{},
					Goals:        []Goal{},
					Notes:        []Note{},
				},
				Path:    entry.Path,
				Missing: true,
			})
			continue
		}
		projects = append(projects, FullProject{
			Project: *p,
			Path:    entry.Path,
			Missing: false,
		})
	}
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastActivity > projects[j].LastActivity
	})
	return projects, nil
}

func DetectTags(dir string) []string {
	var tags []string

	// package.json
	if data, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		var pkg map[string]any
		if json.Unmarshal(data, &pkg) == nil {
			known := []string{"next", "react", "vue", "svelte", "tailwindcss", "express", "fastify", "hono", "astro", "nuxt", "angular", "typescript", "ink"}
			for _, section := range []string{"dependencies", "devDependencies"} {
				if deps, ok := pkg[section].(map[string]any); ok {
					for _, k := range known {
						if _, exists := deps[k]; exists {
							name := k
							if k == "tailwindcss" {
								name = "tailwind"
							}
							tags = appendUnique(tags, name)
						}
					}
				}
			}
		}
	}

	if fileExists(filepath.Join(dir, "pyproject.toml")) || fileExists(filepath.Join(dir, "requirements.txt")) {
		tags = appendUnique(tags, "python")
	}
	if fileExists(filepath.Join(dir, "Cargo.toml")) {
		tags = appendUnique(tags, "rust")
	}
	if fileExists(filepath.Join(dir, "go.mod")) {
		tags = appendUnique(tags, "go")
	}

	return tags
}

func DetectRepo(dir string) string {
	gitConfig := filepath.Join(dir, ".git", "config")
	data, err := os.ReadFile(gitConfig)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.Contains(line, "[remote \"origin\"]") && i+1 < len(lines) {
			for j := i + 1; j < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[j]), "["); j++ {
				if strings.Contains(lines[j], "url =") {
					parts := strings.SplitN(lines[j], "=", 2)
					if len(parts) == 2 {
						return strings.TrimSpace(parts[1])
					}
				}
			}
		}
	}
	return ""
}

// ScanResult is a discovered project not yet tracked by drift
type ScanResult struct {
	Path string
	Tags []string
}

// ScanDir recursively finds projects in a directory.
// maxDepth limits recursion (0 = root only, -1 = unlimited).
// Skips directories that already have .drift/, and known non-project dirs.
func ScanDir(root string, maxDepth int) []ScanResult {
	var results []ScanResult
	skipDirs := map[string]bool{
		"node_modules": true, ".git": true, ".next": true, "__pycache__": true,
		"dist": true, "build": true, "target": true, ".venv": true, "venv": true,
		".drift": true, ".cache": true, "vendor": true,
	}

	var walk func(dir string, depth int)
	walk = func(dir string, depth int) {
		if maxDepth >= 0 && depth > maxDepth {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}

		// Check if this directory IS a project
		isProject := false
		for _, e := range entries {
			switch e.Name() {
			case ".git", "package.json", "pyproject.toml", "Cargo.toml", "go.mod",
				"requirements.txt", "Makefile", "CMakeLists.txt", "pom.xml":
				isProject = true
			}
		}

		if isProject && !HasProject(dir) && dir != root {
			results = append(results, ScanResult{
				Path: dir,
				Tags: DetectTags(dir),
			})
			return // Don't recurse into project subdirs
		}

		// Recurse into subdirectories
		for _, e := range entries {
			if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || skipDirs[e.Name()] {
				continue
			}
			walk(filepath.Join(dir, e.Name()), depth+1)
		}
	}

	walk(root, 0)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Path < results[j].Path
	})
	return results
}

// BuildTree groups projects by their parent directory relative to a common root.
type TreeNode struct {
	Name     string      // directory name
	FullPath string      // absolute path
	Project  *FullProject // nil if just a directory
	Children []*TreeNode
}

// BuildProjectTree organizes projects into a directory tree.
func BuildProjectTree(projects []FullProject) *TreeNode {
	if len(projects) == 0 {
		return &TreeNode{Name: "~", Children: nil}
	}

	// Find common prefix
	home, _ := os.UserHomeDir()
	root := &TreeNode{Name: "~", FullPath: home}

	for i := range projects {
		p := &projects[i]
		rel, err := filepath.Rel(home, p.Path)
		if err != nil {
			rel = p.Path
		}
		parts := strings.Split(rel, string(filepath.Separator))
		insertIntoTree(root, parts, p)
	}

	// Collapse single-child directories
	collapseTree(root)

	return root
}

func insertIntoTree(node *TreeNode, parts []string, project *FullProject) {
	if len(parts) == 0 {
		return
	}
	if len(parts) == 1 {
		// This is the project directory
		node.Children = append(node.Children, &TreeNode{
			Name:     parts[0],
			FullPath: project.Path,
			Project:  project,
		})
		return
	}
	// Find or create intermediate directory
	dirName := parts[0]
	var child *TreeNode
	for _, c := range node.Children {
		if c.Name == dirName && c.Project == nil {
			child = c
			break
		}
	}
	if child == nil {
		child = &TreeNode{
			Name:     dirName,
			FullPath: filepath.Join(node.FullPath, dirName),
		}
		node.Children = append(node.Children, child)
	}
	insertIntoTree(child, parts[1:], project)
}

func collapseTree(node *TreeNode) {
	for _, c := range node.Children {
		collapseTree(c)
	}
	// If a dir has exactly one child that's also a dir (not a project), collapse
	for i, c := range node.Children {
		if c.Project == nil && len(c.Children) == 1 && c.Children[0].Project == nil {
			merged := c.Children[0]
			merged.Name = c.Name + "/" + merged.Name
			node.Children[i] = merged
			collapseTree(merged) // re-check after collapse
		}
	}
}

// FlattenTree produces a flat list with indentation info for rendering.
type TreeLine struct {
	Indent  int
	Name    string
	IsDir   bool
	Project *FullProject
	Last    bool // last child at this level (for tree drawing)
}

func FlattenTree(node *TreeNode, indent int) []TreeLine {
	var lines []TreeLine
	for i, c := range node.Children {
		last := i == len(node.Children)-1
		if c.Project != nil {
			lines = append(lines, TreeLine{
				Indent:  indent,
				Name:    c.Name,
				IsDir:   false,
				Project: c.Project,
				Last:    last,
			})
		} else {
			lines = append(lines, TreeLine{
				Indent: indent,
				Name:   c.Name,
				IsDir:  true,
				Last:   last,
			})
			lines = append(lines, FlattenTree(c, indent+1)...)
		}
	}
	return lines
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func appendUnique(slice []string, val string) []string {
	for _, s := range slice {
		if s == val {
			return slice
		}
	}
	return append(slice, val)
}
