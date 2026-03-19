# drift

Lightweight project tracker for vibe coders. Track dozens of AI-assisted projects without leaving your terminal.

```
drift — 6 projects

  STATUS   PROGRESS  NAME             LAST ACTIVITY
  ●active  ████░ 65% landing-saas     2h ago
  ●active  █████ 90% ai-chatbot       5h ago
  ✓done    █████100% portfolio-v2     2d ago
  ○idea    ░░░░░  0% crypto-tracker   3d ago
  ◊paused  ██░░░ 30% email-tool       1w ago
  ✗abandoned       0% old-experiment  3w ago
```

## The Problem

You vibe-code 3-7 projects a day. After a week you have 20+ folders and can't remember:
- What each project does
- Where you left off
- What's next
- Which ones are worth continuing

Jira is overkill. Notion is too slow. You need something that works at the speed of `ls`.

## How It Works

drift is a **protocol** — a simple file format (`.drift/project.json`) that any tool can read and write. The reference implementation is a **Claude Code skill** that works with zero installation.

```
your-project/
  .drift/
    project.json    ← project metadata, goals, notes

~/.drift/
  registry.json     ← index of all your projects
```

## Quick Start

### Option 1: Claude Code Skill (recommended, zero install)

```bash
# Copy the skill to your Claude Code skills directory
cp -r skills/drift ~/.claude/skills/drift

# Now in any Claude Code session:
/drift init
/drift note "added hero section"
/drift goal "pricing page"
/drift goal done 1
/drift list
```

### Option 2: Manual (works with any editor)

Create `.drift/project.json` in your project:

```json
{
  "id": "generate-a-uuid-here",
  "name": "my-project",
  "description": "What this project does",
  "status": "active",
  "progress": 0,
  "tags": ["next.js", "tailwind"],
  "created": "2026-03-19T10:00:00Z",
  "lastActivity": "2026-03-19T10:00:00Z",
  "goals": [
    { "text": "Hero section", "done": true },
    { "text": "Pricing page", "done": false }
  ],
  "notes": [
    { "ts": "2026-03-19T12:00:00Z", "text": "Added gradient background" }
  ],
  "links": {
    "repo": "https://github.com/you/my-project",
    "deploy": null,
    "design": null
  }
}
```

## Commands

| Command | Description |
|---------|-------------|
| `drift init` | Initialize drift in current project |
| `drift status` | Show current project status |
| `drift note "text"` | Add a timestamped note |
| `drift goal "text"` | Add a goal |
| `drift goal done N` | Mark goal #N as done |
| `drift progress N` | Set progress (0-100) |
| `drift set-status active` | Change status (idea/active/paused/done/abandoned) |
| `drift list` | List all projects |
| `drift scan ~/Develop` | Find untracked projects |
| `drift open name` | Get project path |
| `drift describe "text"` | Set project description |
| `drift tag next.js ai` | Add tags |
| `drift link deploy https://...` | Set a link |

## Protocol

drift is protocol-first. The `.drift/` format is the product — implementations are just consumers.

- [Protocol Specification](docs/protocol.md)
- [project.json Schema](docs/schema/project.schema.json)
- [registry.json Schema](docs/schema/registry.schema.json)

Anyone can build a drift-compatible tool: CLI in any language, TUI, VS Code extension, web dashboard. As long as it reads and writes the `.drift/` format, it's compatible.

## Design Principles

- **File-based** — plain JSON, no database, no server
- **Zero-friction** — `drift init` and you're done
- **Auto-enrichment** — detects stack, repo URL, deploy URL automatically
- **Tool-agnostic** — works with Claude, Cursor, any AI assistant, or no AI at all
- **Local-first** — your data stays on your machine
- **Append-only notes** — never lose context

## Roadmap

- [x] Protocol specification
- [x] JSON Schema
- [x] Claude Code skill
- [ ] CLI (TypeScript, `npx drift-cli`)
- [ ] TUI (Norton Commander-style dual panel)
- [ ] Web dashboard
- [ ] MCP server for Claude Code
- [ ] VS Code extension

## License

[MIT](LICENSE)
