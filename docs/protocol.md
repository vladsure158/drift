# drift Protocol Specification

Version: 1.0.0

## Overview

drift is an open protocol for tracking vibe-coding project metadata. It defines a file-based format that any tool — CLI, AI assistant, TUI, web dashboard — can read and write.

The protocol solves three core problems for developers who run many projects in parallel:
1. **No central registry** — projects are scattered across the filesystem with no index
2. **Lost context** — switching between projects means losing track of what was done and what's next
3. **No quick save** — there's no fast way to capture progress before switching

## Design Principles

- **File-based** — plain JSON files, no database, no server
- **Human-readable** — open any file in a text editor and understand it
- **Tolerant** — missing fields default to null/empty; unknown fields are preserved
- **Local-first** — all data on your machine, nothing sent anywhere
- **Tool-agnostic** — works with any implementation (Claude skill, CLI, TUI, web)
- **Git-friendly** — JSON with stable key ordering; `.drift/` can be gitignored or committed

## File Layout

### Per-project metadata

```
<project-root>/
  .drift/
    project.json    # Project metadata
```

Each project that opts into drift tracking gets a `.drift/` directory at its root containing a single `project.json` file.

### Global registry

```
~/.drift/
  registry.json     # Index of all known projects
```

The registry lives in the user's home directory and serves as the central index of all drift-tracked projects. It contains minimal data — just enough to list projects without reading each one.

## Schemas

### project.json

The canonical metadata file for a single project.

```jsonc
{
  // Required
  "id": "550e8400-e29b-41d4-a716-446655440000",  // UUID v4, immutable after creation
  "name": "landing-saas",                          // Human-readable name
  "created": "2026-03-19T10:00:00Z",               // ISO 8601, immutable after creation

  // Status tracking
  "status": "active",          // One of: idea, active, paused, done, abandoned
  "progress": 65,              // Integer 0-100; auto-calculated from goals if present
  "lastActivity": "2026-03-19T14:30:00Z",  // ISO 8601, updated on any mutation

  // Context
  "description": "AI startup landing page with pricing and auth",  // Why this project exists
  "tags": ["next.js", "tailwind", "shadcn"],                       // Freeform tags

  // Goals — ordered checklist
  "goals": [
    { "text": "Hero section", "done": true },
    { "text": "Pricing page", "done": false },
    { "text": "Deploy to Vercel", "done": false }
  ],

  // Notes — append-only log
  "notes": [
    { "ts": "2026-03-19T12:00:00Z", "text": "Added hero with gradient bg" },
    { "ts": "2026-03-19T14:30:00Z", "text": "Stuck on Stripe integration, need API key" }
  ],

  // Links — optional external references
  "links": {
    "repo": "https://github.com/user/landing-saas",
    "deploy": "https://landing-saas.vercel.app",
    "design": "https://figma.com/file/abc123"
  }
}
```

**Field rules:**

| Field | Type | Required | Default | Mutability |
|-------|------|----------|---------|------------|
| `id` | UUID v4 string | yes | generated | immutable |
| `name` | string | yes | directory name | mutable |
| `created` | ISO 8601 | yes | now | immutable |
| `status` | enum | no | `"active"` | mutable |
| `progress` | integer 0-100 | no | `0` | mutable / auto |
| `lastActivity` | ISO 8601 | no | `created` | auto-updated |
| `description` | string \| null | no | `null` | mutable |
| `tags` | string[] | no | `[]` | mutable |
| `goals` | goal[] | no | `[]` | mutable |
| `notes` | note[] | no | `[]` | append-only |
| `links` | object | no | `{}` | mutable |

**Status values:**

| Status | Meaning |
|--------|---------|
| `idea` | Just an idea, not started yet |
| `active` | Currently being worked on |
| `paused` | Intentionally on hold (will return) |
| `done` | Completed, shipped |
| `abandoned` | Intentionally dropped (won't return) |

**Progress auto-calculation:**
When `goals` array is non-empty, `progress` is automatically calculated as:
```
progress = round(done_goals / total_goals * 100)
```
If `progress` is explicitly set and goals exist, the explicit value takes precedence.

**Notes are append-only:**
Notes are never edited or deleted. They form a chronological log. Each note has:
- `ts` — ISO 8601 timestamp (required)
- `text` — freeform text (required)

### registry.json

The global index of all known projects.

```jsonc
{
  "version": 1,
  "projects": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "path": "/Users/artem/Develop/landing-saas",
      "name": "landing-saas",
      "status": "active",
      "lastActivity": "2026-03-19T14:30:00Z"
    },
    {
      "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
      "path": "/Users/artem/Develop/ai-chatbot",
      "name": "ai-chatbot",
      "status": "active",
      "lastActivity": "2026-03-18T20:00:00Z"
    }
  ]
}
```

**Field rules:**

| Field | Type | Required | Source |
|-------|------|----------|--------|
| `version` | integer | yes | always `1` |
| `projects` | project-ref[] | yes | — |
| `projects[].id` | UUID v4 | yes | from project.json |
| `projects[].path` | absolute path | yes | filesystem location |
| `projects[].name` | string | yes | from project.json |
| `projects[].status` | enum | yes | from project.json |
| `projects[].lastActivity` | ISO 8601 | yes | from project.json |

**Registry is a cache:**
The registry is a convenience index. The source of truth is always the individual `project.json` files. If a registry entry conflicts with a project.json, the project.json wins. Implementations should sync the registry when possible.

**Stale entries:**
If a registry entry points to a path that no longer exists or no longer contains `.drift/project.json`, implementations should mark it as stale or remove it on next access.

## Operations

The protocol defines these logical operations. Implementations may expose them however they want (CLI commands, natural language, UI buttons).

### init

Create `.drift/project.json` in the current directory and register in `~/.drift/registry.json`.

**Auto-enrichment:** Implementations SHOULD auto-detect:
- `name` — from directory name
- `tags` — from `package.json` (dependencies), `pyproject.toml`, `Cargo.toml`, `go.mod`
- `links.repo` — from `git remote get-url origin`
- `links.deploy` — from `.vercel/project.json` if present

**Gitignore:** Implementations SHOULD add `.drift/` to the project's `.gitignore` if it exists, unless the user explicitly wants to track drift metadata in git.

### note

Append a timestamped note to `project.json.notes[]` and update `lastActivity`.

### goal / goal done

Add a goal to `project.json.goals[]` or mark one as done. Recalculate `progress` if auto-calculation is active.

### status

Read and display current project metadata.

### set-status

Change `project.json.status` and sync to registry.

### progress

Manually set `project.json.progress` (overrides auto-calculation).

### list

Read `~/.drift/registry.json` and display all projects. Optionally enrich with data from individual project.json files.

### scan

Walk a directory tree looking for projects (by `.git/`, `package.json`, `pyproject.toml`, `Cargo.toml`, `go.mod`, etc.). Offer to `init` each found project.

### open

Return the filesystem path of a project by name or id (for use with `cd`).

## Versioning

The protocol uses a `version` field in registry.json. The current version is `1`.

Future versions will be backward-compatible where possible. If a breaking change is needed, the version number increments and implementations should migrate old data.

project.json intentionally has no version field — it relies on tolerant parsing (unknown fields preserved, missing fields defaulted).

## Conventions

- **Encoding:** UTF-8
- **Line endings:** LF (Unix-style)
- **JSON formatting:** 2-space indentation, trailing newline
- **Timestamps:** ISO 8601 with timezone (prefer UTC with `Z` suffix)
- **UUIDs:** v4, lowercase, with hyphens
- **File permissions:** user-readable/writable (0644)
