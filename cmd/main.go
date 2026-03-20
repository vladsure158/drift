package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/snowtema/drift/internal/commands"
	"github.com/snowtema/drift/internal/ui"
)

var version = "dev"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		// No args: launch TUI
		launchTUI()
		return
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "init":
		dir := ""
		if len(rest) > 0 {
			dir = rest[0]
		}
		commands.Init(dir)

	case "status", "s":
		commands.Status()

	case "list", "ls":
		sort := "activity"
		for _, a := range rest {
			if strings.HasPrefix(a, "--sort=") {
				sort = strings.TrimPrefix(a, "--sort=")
			}
		}
		commands.List(sort)

	case "note", "n":
		text := strings.Join(rest, " ")
		if text == "" {
			fmt.Println("  Usage: drift note \"text\"")
			os.Exit(1)
		}
		commands.Note(text)

	case "goal", "g":
		if len(rest) > 0 && rest[0] == "done" {
			if len(rest) < 2 {
				fmt.Println("  Usage: drift goal done N")
				os.Exit(1)
			}
			commands.GoalDone(rest[1])
		} else {
			text := strings.Join(rest, " ")
			if text == "" {
				fmt.Println("  Usage: drift goal \"text\"")
				os.Exit(1)
			}
			commands.Goal(text)
		}

	case "progress", "p":
		if len(rest) == 0 {
			fmt.Println("  Usage: drift progress N")
			os.Exit(1)
		}
		commands.Progress(rest[0])

	case "set-status":
		if len(rest) == 0 {
			fmt.Println("  Usage: drift set-status STATUS")
			os.Exit(1)
		}
		commands.SetStatus(rest[0])

	case "describe", "desc":
		commands.Describe(strings.Join(rest, " "))

	case "tag":
		if len(rest) == 0 {
			fmt.Println("  Usage: drift tag tag1 tag2")
			os.Exit(1)
		}
		commands.Tag(rest)

	case "link":
		if len(rest) < 2 {
			fmt.Println("  Usage: drift link type url")
			os.Exit(1)
		}
		commands.Link(rest[0], rest[1])

	case "open", "o":
		if len(rest) == 0 {
			fmt.Println("  Usage: drift open name")
			os.Exit(1)
		}
		commands.Open(rest[0])

	case "scan":
		doInit := false
		dir := ""
		maxDepth := 3 // default recursive depth
		for _, a := range rest {
			if a == "--init" {
				doInit = true
			} else if strings.HasPrefix(a, "--depth=") {
				fmt.Sscanf(strings.TrimPrefix(a, "--depth="), "%d", &maxDepth)
			} else {
				dir = a
			}
		}
		commands.Scan(dir, doInit, maxDepth)

	case "help", "-h", "--help":
		commands.Help()

	case "tui":
		launchTUI()

	default:
		fmt.Printf("  Unknown command: %s\n", cmd)
		commands.Help()
		os.Exit(1)
	}
}

func launchTUI() {
	p := tea.NewProgram(
		ui.NewModel(version),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
