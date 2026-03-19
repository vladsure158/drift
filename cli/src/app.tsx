import React, { useState, useEffect, useMemo } from "react";
import { Box, Text, useApp, useInput } from "ink";
import TextInput from "ink-text-input";
import {
  readRegistry, loadAllProjects, readProject, writeProject,
  syncToRegistry, calcProgress, now,
  type Project, type ProjectStatus,
} from "./lib/protocol.js";
import { useTerminalSize } from "./hooks/useTerminalSize.js";

type FullProject = Project & { _path: string; _missing: boolean };

// ─── View State Machine ──────────────────────────

type ViewMode =
  | "list"
  | "detail"
  | "input-note"
  | "input-goal"
  | "input-desc"
  | "pick-status"
  | "pick-goal-done";

// ─── Sort ────────────────────────────────────────

type SortMode = "activity" | "progress" | "name" | "status";
const SORT_MODES: SortMode[] = ["activity", "progress", "name", "status"];
const SORT_LABELS: Record<SortMode, string> = {
  activity: "recent", progress: "progress", name: "name", status: "status",
};
const STATUS_ORDER: Record<string, number> = {
  active: 0, idea: 1, paused: 2, done: 3, abandoned: 4,
};
const ALL_STATUSES: ProjectStatus[] = ["active", "idea", "paused", "done", "abandoned"];

function sortProjects(projects: FullProject[], mode: SortMode): FullProject[] {
  const s = [...projects];
  switch (mode) {
    case "activity": return s.sort((a, b) => b.lastActivity.localeCompare(a.lastActivity));
    case "progress": return s.sort((a, b) => b.progress - a.progress || a.name.localeCompare(b.name));
    case "name":     return s.sort((a, b) => a.name.localeCompare(b.name));
    case "status":   return s.sort((a, b) => (STATUS_ORDER[a.status] ?? 9) - (STATUS_ORDER[b.status] ?? 9) || b.lastActivity.localeCompare(a.lastActivity));
    default: return s;
  }
}

// ─── Helpers ─────────────────────────────────────

const ST: Record<string, { icon: string; color: string }> = {
  active: { icon: "●", color: "cyan" },
  done: { icon: "✓", color: "green" },
  idea: { icon: "○", color: "yellow" },
  paused: { icon: "◇", color: "gray" },
  abandoned: { icon: "✗", color: "gray" },
};

function timeSince(iso: string): string {
  if (!iso) return "";
  const diff = Date.now() - new Date(iso).getTime();
  const m = Math.floor(diff / 60000);
  if (m < 1) return "now";
  if (m < 60) return `${m}m`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h`;
  const d = Math.floor(h / 24);
  if (d < 7) return `${d}d`;
  return `${Math.floor(d / 7)}w`;
}

// ─── Project Row ─────────────────────────────────

function ProjectRow({ p, selected, dimmed, width }: {
  p: FullProject; selected: boolean; dimmed: boolean; width: number;
}) {
  const st = ST[p.status] ?? ST.active;
  const pct = String(p.progress).padStart(3);
  const time = timeSince(p.lastActivity).padStart(3);
  const maxName = Math.max(8, width - 15);
  const name = p.name.length > maxName ? p.name.slice(0, maxName - 1) + "…" : p.name;

  if (selected && !dimmed) {
    return (
      <Box><Text backgroundColor="gray" color="white" bold>
        {` ${st.icon} ${name.padEnd(maxName)}  ${pct}%  ${time} `}
      </Text></Box>
    );
  }

  return (
    <Box>
      <Text dimColor={dimmed}>
        {" "}<Text color={dimmed ? undefined : st.color as any} dimColor={dimmed}>{st.icon}</Text>{" "}
        {name.padEnd(maxName)}
        {"  "}<Text dimColor>{pct}%</Text>
        {"  "}<Text dimColor>{time}</Text>{" "}
      </Text>
    </Box>
  );
}

// ─── Detail Panel ────────────────────────────────

function DetailPanel({ p, maxHeight, focused, goalHighlight }: {
  p: FullProject | null; maxHeight: number; focused: boolean; goalHighlight?: number;
}) {
  if (!p) return <Text dimColor> Select a project</Text>;

  const st = ST[p.status] ?? ST.active;
  const doneGoals = p.goals.filter(g => g.done).length;
  const shortPath = p._path.replace(process.env.HOME ?? "", "~");
  const usedLines = 7 + (p.tags.length > 0 ? 2 : 0) + (p.goals.length > 0 ? p.goals.length + 1 : 0);
  const notesRoom = Math.max(1, maxHeight - usedLines - 4);
  const showNotes = p.notes.slice(-notesRoom).reverse();

  return (
    <Box flexDirection="column" paddingX={1} height={maxHeight} overflowY="hidden">
      <Box>
        <Text bold color={st.color as any}>{p.name}</Text>
        <Text dimColor>  {p.status}  {p.progress}%</Text>
        {p._missing && <Text color="red"> [missing]</Text>}
      </Box>

      <Text dimColor={!focused}>{p.description ?? "no description — press D to add"}</Text>

      {p.tags.length > 0 && (
        <Box marginTop={1} gap={1} flexWrap="wrap">
          {p.tags.map(tag => <Text key={tag} color="cyan" dimColor>#{tag}</Text>)}
        </Box>
      )}

      <Box marginTop={1}><Text dimColor>{shortPath}</Text></Box>

      {/* Goals */}
      <Box flexDirection="column" marginTop={1}>
        <Text bold>
          goals <Text dimColor>{doneGoals}/{p.goals.length}</Text>
          {p.goals.length === 0 && <Text dimColor> — press G to add</Text>}
        </Text>
        {p.goals.map((g, i) => {
          const isHighlighted = goalHighlight === i;
          return (
            <Box key={i}>
              <Text
                color={isHighlighted ? "yellow" : g.done ? "green" : "gray"}
                bold={isHighlighted}
              >
                {g.done ? " ✓ " : ` ${i + 1} `}
              </Text>
              <Text
                dimColor={g.done && !isHighlighted}
                strikethrough={g.done}
                bold={isHighlighted}
              >
                {g.text}
              </Text>
            </Box>
          );
        })}
      </Box>

      {/* Notes */}
      <Box flexDirection="column" marginTop={1}>
        <Text bold>
          notes <Text dimColor>{p.notes.length}</Text>
          {p.notes.length === 0 && <Text dimColor> — press N to add</Text>}
        </Text>
        {showNotes.map((n, i) => (
          <Box key={i} gap={1}>
            <Text dimColor>{n.ts.slice(5, 10)} {n.ts.slice(11, 16)}</Text>
            <Text>{n.text.length > 55 ? n.text.slice(0, 54) + "…" : n.text}</Text>
          </Box>
        ))}
        {p.notes.length > showNotes.length && (
          <Text dimColor> +{p.notes.length - showNotes.length} more</Text>
        )}
      </Box>

      {/* Links */}
      {Object.entries(p.links).some(([, v]) => v) && (
        <Box flexDirection="column" marginTop={1}>
          {Object.entries(p.links).map(([k, v]) =>
            v ? <Box key={k} gap={1}><Text dimColor>{k}</Text><Text color="cyan">{v}</Text></Box> : null
          )}
        </Box>
      )}
    </Box>
  );
}

// ─── Footer per mode ─────────────────────────────

function FooterBar({ mode }: { mode: ViewMode }) {
  const items: { key: string; label: string }[] = [];

  switch (mode) {
    case "list":
      items.push({ key: "↑↓", label: "nav" }, { key: "⏎", label: "select" }, { key: "s", label: "sort" }, { key: "q", label: "quit" });
      break;
    case "detail":
      items.push(
        { key: "n", label: "note" }, { key: "g", label: "goal" }, { key: "d", label: "done #" },
        { key: "D", label: "describe" }, { key: "1-5", label: "status" }, { key: "o", label: "open" },
        { key: "esc", label: "back" },
      );
      break;
    case "input-note":
      items.push({ key: "type", label: "enter note" }, { key: "⏎", label: "save" }, { key: "esc", label: "cancel" });
      break;
    case "input-goal":
      items.push({ key: "type", label: "enter goal" }, { key: "⏎", label: "save" }, { key: "esc", label: "cancel" });
      break;
    case "input-desc":
      items.push({ key: "type", label: "enter description" }, { key: "⏎", label: "save" }, { key: "esc", label: "cancel" });
      break;
    case "pick-status":
      items.push(
        { key: "1", label: "active" }, { key: "2", label: "idea" }, { key: "3", label: "paused" },
        { key: "4", label: "done" }, { key: "5", label: "abandoned" }, { key: "esc", label: "cancel" },
      );
      break;
    case "pick-goal-done":
      items.push({ key: "#", label: "goal number to complete" }, { key: "esc", label: "cancel" });
      break;
  }

  return (
    <Box paddingX={1} height={1} gap={2}>
      {items.map(({ key, label }) => (
        <Text key={key + label}>
          <Text bold color="cyan">{key}</Text>
          <Text dimColor>{" " + label}</Text>
        </Text>
      ))}
    </Box>
  );
}

// ─── Input Bar ───────────────────────────────────

function InputBar({ label, value, onChange, onSubmit }: {
  label: string; value: string; onChange: (v: string) => void; onSubmit: (v: string) => void;
}) {
  return (
    <Box paddingX={1} height={1}>
      <Text bold color="cyan">{label}: </Text>
      <TextInput value={value} onChange={onChange} onSubmit={onSubmit} />
    </Box>
  );
}

// ─── Flash Message ───────────────────────────────

function FlashMessage({ text }: { text: string | null }) {
  if (!text) return null;
  return (
    <Box paddingX={1} height={1}>
      <Text color="green">{text}</Text>
    </Box>
  );
}

// ─── App ─────────────────────────────────────────

export function App() {
  const { exit } = useApp();
  const { columns, rows } = useTerminalSize();

  const [allProjects, setAllProjects] = useState<FullProject[]>([]);
  const [sortMode, setSortMode] = useState<SortMode>("activity");
  const [idx, setIdx] = useState(0);
  const [scrollOffset, setScrollOffset] = useState(0);
  const [mode, setMode] = useState<ViewMode>("list");
  const [inputValue, setInputValue] = useState("");
  const [flash, setFlash] = useState<string | null>(null);

  // Load
  useEffect(() => {
    setAllProjects(loadAllProjects(readRegistry()));
  }, []);

  const projects = useMemo(() => sortProjects(allProjects, sortMode), [allProjects, sortMode]);
  const listHeight = Math.max(5, rows - 3);

  // Keep selection in view
  useEffect(() => {
    if (idx < scrollOffset) setScrollOffset(idx);
    if (idx >= scrollOffset + listHeight) setScrollOffset(idx - listHeight + 1);
  }, [idx, listHeight, scrollOffset]);

  // Flash auto-clear
  useEffect(() => {
    if (!flash) return;
    const t = setTimeout(() => setFlash(null), 2000);
    return () => clearTimeout(t);
  }, [flash]);

  // Reload project data after mutation
  function reload() {
    setAllProjects(loadAllProjects(readRegistry()));
  }

  // Mutate the selected project
  function mutate(fn: (p: Project) => void) {
    const sel = projects[idx];
    if (!sel || sel._missing) return;
    const p = readProject(sel._path);
    if (!p) return;
    fn(p);
    p.lastActivity = now();
    writeProject(sel._path, p);
    syncToRegistry(sel._path, p);
    reload();
  }

  const selected = projects[idx] ?? null;

  // ─── Input modes: only TextInput handles keys ──
  // For input modes, we DON'T call useInput (TextInput takes over)
  // We handle Esc via a wrapper

  // ─── Keyboard for non-input modes ──────────────
  useInput((input, key) => {
    // Input modes: only Esc works here (TextInput handles the rest)
    if (mode === "input-note" || mode === "input-goal" || mode === "input-desc") {
      if (key.escape) { setMode("detail"); setInputValue(""); }
      return;
    }

    // Quit
    if (input === "q" && mode === "list") { exit(); return; }
    if (key.ctrl && input === "c") { exit(); return; }

    // ── LIST mode ──
    if (mode === "list") {
      if (key.upArrow || input === "k") setIdx(i => i > 0 ? i - 1 : projects.length - 1);
      if (key.downArrow || input === "j") setIdx(i => i < projects.length - 1 ? i + 1 : 0);
      if (key.return) setMode("detail");
      if (input === "s") {
        setSortMode(prev => SORT_MODES[(SORT_MODES.indexOf(prev) + 1) % SORT_MODES.length]);
        setIdx(0); setScrollOffset(0);
      }
      if (key.ctrl && input === "u") setIdx(i => Math.max(0, i - listHeight));
      if (key.ctrl && input === "d") setIdx(i => Math.min(projects.length - 1, i + listHeight));
      if (input === "g" && !key.shift) { setIdx(0); setScrollOffset(0); }
      if (input === "G") setIdx(projects.length - 1);
      return;
    }

    // ── DETAIL mode ──
    if (mode === "detail") {
      if (key.escape || input === "q") { setMode("list"); return; }
      // Navigate between projects while in detail
      if (key.upArrow || input === "k") setIdx(i => i > 0 ? i - 1 : projects.length - 1);
      if (key.downArrow || input === "j") setIdx(i => i < projects.length - 1 ? i + 1 : 0);
      // Actions
      if (input === "n") { setMode("input-note"); setInputValue(""); return; }
      if (input === "g" && !key.shift) { setMode("input-goal"); setInputValue(""); return; }
      if (input === "D") { setMode("input-desc"); setInputValue(selected?.description ?? ""); return; }
      if (input === "d") { setMode("pick-goal-done"); return; }
      // Status: 1-5
      if (input === "1") { mutate(p => { p.status = "active"; }); setFlash("✓ active"); return; }
      if (input === "2") { mutate(p => { p.status = "idea"; }); setFlash("✓ idea"); return; }
      if (input === "3") { mutate(p => { p.status = "paused"; }); setFlash("✓ paused"); return; }
      if (input === "4") { mutate(p => { p.status = "done"; }); setFlash("✓ done"); return; }
      if (input === "5") { mutate(p => { p.status = "abandoned"; }); setFlash("✓ abandoned"); return; }
      // Open
      if (input === "o") {
        if (selected) setFlash(`cd ${selected._path}`);
        return;
      }
      return;
    }

    // ── PICK-STATUS mode ──
    if (mode === "pick-status") {
      if (key.escape) { setMode("detail"); return; }
      const statusMap: Record<string, ProjectStatus> = { "1": "active", "2": "idea", "3": "paused", "4": "done", "5": "abandoned" };
      if (input in statusMap) {
        mutate(p => { p.status = statusMap[input]; });
        setFlash(`✓ Status: ${statusMap[input]}`);
        setMode("detail");
      }
      return;
    }

    // ── PICK-GOAL-DONE mode ──
    if (mode === "pick-goal-done") {
      if (key.escape) { setMode("detail"); return; }
      const goalIdx = parseInt(input, 10) - 1;
      if (!isNaN(goalIdx) && selected && goalIdx >= 0 && goalIdx < selected.goals.length) {
        mutate(p => {
          p.goals[goalIdx].done = !p.goals[goalIdx].done;
          p.progress = calcProgress(p.goals);
        });
        const wasToggled = selected.goals[goalIdx].done;
        setFlash(`✓ Goal #${goalIdx + 1} ${wasToggled ? "undone" : "done"}`);
        setMode("detail");
      }
      return;
    }
  });

  // ─── Submit handlers for text input ────────────

  function onSubmitNote(value: string) {
    if (value.trim()) {
      mutate(p => { p.notes.push({ ts: now(), text: value.trim() }); });
      setFlash("✓ Note added");
    }
    setMode("detail"); setInputValue("");
  }

  function onSubmitGoal(value: string) {
    if (value.trim()) {
      mutate(p => {
        p.goals.push({ text: value.trim(), done: false });
        p.progress = calcProgress(p.goals);
      });
      setFlash("✓ Goal added");
    }
    setMode("detail"); setInputValue("");
  }

  function onSubmitDesc(value: string) {
    mutate(p => { p.description = value.trim() || null; });
    setFlash("✓ Description updated");
    setMode("detail"); setInputValue("");
  }

  // ─── Layout ────────────────────────────────────

  const listW = Math.min(40, Math.floor(columns * 0.38));
  const visibleProjects = projects.slice(scrollOffset, scrollOffset + listHeight);
  const listDimmed = mode !== "list";
  const isInputMode = mode === "input-note" || mode === "input-goal" || mode === "input-desc";
  const inputLabel = mode === "input-note" ? "note" : mode === "input-goal" ? "goal" : mode === "input-desc" ? "desc" : "";

  if (projects.length === 0) {
    return (
      <Box flexDirection="column" width={columns} height={rows} alignItems="center" justifyContent="center">
        <Text dimColor>No projects tracked.</Text>
        <Text>Run <Text color="cyan" bold>drift init</Text> in a project directory.</Text>
      </Box>
    );
  }

  return (
    <Box flexDirection="column" width={columns} height={rows}>
      {/* ── Header ── */}
      <Box paddingX={1} justifyContent="space-between" height={1}>
        <Box>
          <Text bold color="cyan">drift</Text>
          <Text dimColor> {projects.length} projects</Text>
          {projects.length > listHeight && (
            <Text dimColor> [{scrollOffset + 1}–{Math.min(scrollOffset + listHeight, projects.length)}]</Text>
          )}
        </Box>
        <Box gap={1}>
          {SORT_MODES.map((m, i) => (
            <Text key={m} color={sortMode === m ? "cyan" : undefined} dimColor={sortMode !== m} bold={sortMode === m}>
              {SORT_LABELS[m]}
            </Text>
          ))}
        </Box>
      </Box>

      {/* ── Main ── */}
      <Box flexGrow={1} height={listHeight}>
        {/* List */}
        <Box
          flexDirection="column"
          width={listW}
          borderStyle="single"
          borderColor={listDimmed ? "gray" : "cyan"}
          borderRight={true}
          borderLeft={false}
          borderTop={false}
          borderBottom={false}
          height={listHeight}
          overflowY="hidden"
        >
          {visibleProjects.map((p, vi) => (
            <ProjectRow
              key={p.id}
              p={p}
              selected={scrollOffset + vi === idx}
              dimmed={listDimmed}
              width={listW - 1}
            />
          ))}
        </Box>

        {/* Detail */}
        <Box flexDirection="column" flexGrow={1} height={listHeight}>
          <DetailPanel
            p={selected}
            maxHeight={listHeight}
            focused={mode !== "list"}
            goalHighlight={mode === "pick-goal-done" ? -1 : undefined}
          />
        </Box>
      </Box>

      {/* ── Bottom bar ── */}
      {flash ? (
        <FlashMessage text={flash} />
      ) : isInputMode ? (
        <InputBar
          label={inputLabel}
          value={inputValue}
          onChange={setInputValue}
          onSubmit={
            mode === "input-note" ? onSubmitNote :
            mode === "input-goal" ? onSubmitGoal :
            onSubmitDesc
          }
        />
      ) : (
        <FooterBar mode={mode} />
      )}
    </Box>
  );
}
