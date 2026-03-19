import { existsSync, readFileSync, writeFileSync, readdirSync, statSync } from "node:fs";
import { resolve, basename, join } from "node:path";
import { execSync } from "node:child_process";
import chalk from "chalk";
import {
  hasProject, readProject, writeProject, createProject,
  syncToRegistry, readRegistry, loadAllProjects, calcProgress,
  now, type Project, type ProjectStatus,
} from "./protocol.js";

const CWD = process.cwd();

function statusIcon(status: string): string {
  const icons: Record<string, string> = {
    active: chalk.cyan("●"),
    done: chalk.green("✓"),
    idea: chalk.yellow("○"),
    paused: chalk.dim("◊"),
    abandoned: chalk.dim("✗"),
  };
  return icons[status] ?? "?";
}

function miniBar(pct: number): string {
  const w = 5;
  const filled = Math.round((pct / 100) * w);
  const bar = "█".repeat(filled) + "░".repeat(w - filled);
  if (pct >= 100) return chalk.green(bar);
  if (pct >= 60) return chalk.cyan(bar);
  if (pct >= 30) return chalk.yellow(bar);
  return chalk.dim(bar);
}

function timeSince(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;
  const weeks = Math.floor(days / 7);
  if (weeks < 4) return `${weeks}w ago`;
  return `${Math.floor(days / 30)}mo ago`;
}

function detectTags(dir: string): string[] {
  const tags: string[] = [];
  const pkgPath = join(dir, "package.json");
  if (existsSync(pkgPath)) {
    try {
      const pkg = JSON.parse(readFileSync(pkgPath, "utf-8"));
      const allDeps = { ...pkg.dependencies, ...pkg.devDependencies };
      const known = ["next", "react", "vue", "svelte", "tailwindcss", "express", "fastify", "hono", "astro", "nuxt", "angular", "solid-js", "ink", "typescript"];
      for (const k of known) {
        if (k in allDeps) tags.push(k === "tailwindcss" ? "tailwind" : k);
      }
    } catch {}
  }
  if (existsSync(join(dir, "pyproject.toml")) || existsSync(join(dir, "requirements.txt"))) tags.push("python");
  if (existsSync(join(dir, "Cargo.toml"))) tags.push("rust");
  if (existsSync(join(dir, "go.mod"))) tags.push("go");
  return tags;
}

function detectRepo(dir: string): string | null {
  try {
    return execSync("git remote get-url origin", { cwd: dir, encoding: "utf-8", stdio: ["pipe", "pipe", "pipe"] }).trim() || null;
  } catch { return null; }
}

function addToGitignore(dir: string) {
  const gi = join(dir, ".gitignore");
  if (!existsSync(gi)) return;
  const content = readFileSync(gi, "utf-8");
  if (!content.includes(".drift/")) {
    writeFileSync(gi, content.trimEnd() + "\n.drift/\n");
  }
}

// ─── Commands ────────────────────────────────────────

export function cmdInit(dir?: string) {
  const root = dir ? resolve(dir) : CWD;

  if (hasProject(root)) {
    console.log(chalk.yellow("  Проект уже инициализирован."));
    cmdStatus(root);
    return;
  }

  const project = createProject(root);
  project.tags = detectTags(root);
  project.links.repo = detectRepo(root);

  writeProject(root, project);
  syncToRegistry(root, project);
  addToGitignore(root);

  console.log(`\n  ${chalk.green("✓")} ${chalk.bold("drift init")} — ${chalk.cyan(project.name)}`);
  console.log(`  Status: ${project.status} | Progress: ${project.progress}%`);
  if (project.tags.length) console.log(`  Tags: ${project.tags.join(", ")}`);
  if (project.links.repo) console.log(`  Repo: ${project.links.repo}`);
  console.log();
}

export function cmdStatus(dir?: string) {
  const root = dir ?? CWD;
  const p = readProject(root);
  if (!p) {
    console.log(chalk.dim("  Проект не инициализирован. Запусти: drift init"));
    return;
  }

  const doneGoals = p.goals.filter(g => g.done).length;
  console.log(`\n  📂 ${chalk.bold(p.name)} ${chalk.dim(`[${p.status}]`)} ${p.progress}%`);
  console.log(`  ${p.description ?? chalk.dim("нет описания")}`);
  if (p.tags.length) console.log(`  Tags: ${p.tags.map(t => chalk.cyan(t)).join("  ")}`);
  console.log(`  Last: ${timeSince(p.lastActivity)}`);

  if (p.goals.length) {
    console.log(`\n  Goals ${chalk.dim(`${doneGoals}/${p.goals.length}`)}`);
    p.goals.forEach((g, i) => {
      const icon = g.done ? chalk.green("✓") : chalk.dim("○");
      const text = g.done ? chalk.dim.strikethrough(g.text) : g.text;
      console.log(`  ${icon} ${chalk.dim(`${i + 1}.`)} ${text}`);
    });
  }

  if (p.notes.length) {
    const recent = p.notes.slice(-5).reverse();
    console.log(`\n  Notes`);
    recent.forEach(n => {
      console.log(`  ${chalk.dim(n.ts.slice(11, 16))}  ${n.text}`);
    });
    if (p.notes.length > 5) console.log(chalk.dim(`  +${p.notes.length - 5} more`));
  }

  const links = Object.entries(p.links).filter(([, v]) => v);
  if (links.length) {
    console.log();
    links.forEach(([k, v]) => console.log(`  ${chalk.dim(k + ":")} ${chalk.cyan(v)}`));
  }
  console.log();
}

export function cmdNote(text: string) {
  const p = readProject(CWD);
  if (!p) { console.log(chalk.red("  Нет .drift/ — запусти drift init")); return; }

  p.notes.push({ ts: now(), text });
  p.lastActivity = now();
  writeProject(CWD, p);
  syncToRegistry(CWD, p);
  console.log(`  ${chalk.green("✓")} Note added to ${chalk.cyan(p.name)}`);
}

export function cmdGoal(text: string) {
  const p = readProject(CWD);
  if (!p) { console.log(chalk.red("  Нет .drift/ — запусти drift init")); return; }

  p.goals.push({ text, done: false });
  p.progress = calcProgress(p.goals);
  p.lastActivity = now();
  writeProject(CWD, p);
  syncToRegistry(CWD, p);

  console.log(`  ${chalk.green("✓")} Goal added to ${chalk.cyan(p.name)}`);
  printGoals(p);
}

export function cmdGoalDone(n: number) {
  const p = readProject(CWD);
  if (!p) { console.log(chalk.red("  Нет .drift/ — запусти drift init")); return; }
  if (n < 1 || n > p.goals.length) { console.log(chalk.red(`  Нет цели #${n}`)); return; }

  p.goals[n - 1].done = true;
  p.progress = calcProgress(p.goals);
  p.lastActivity = now();
  writeProject(CWD, p);
  syncToRegistry(CWD, p);

  console.log(`  ${chalk.green("✓")} Goal #${n} done!`);
  printGoals(p);

  if (p.goals.every(g => g.done)) {
    console.log(chalk.green("\n  🎉 All goals done! Run: drift set-status done"));
  }
}

export function cmdProgress(n: number) {
  const p = readProject(CWD);
  if (!p) { console.log(chalk.red("  Нет .drift/ — запусти drift init")); return; }

  p.progress = Math.max(0, Math.min(100, n));
  p.lastActivity = now();
  writeProject(CWD, p);
  syncToRegistry(CWD, p);
  console.log(`  ${chalk.green("✓")} Progress: ${miniBar(p.progress)} ${p.progress}%`);
}

export function cmdSetStatus(status: string) {
  const valid: ProjectStatus[] = ["idea", "active", "paused", "done", "abandoned"];
  if (!valid.includes(status as ProjectStatus)) {
    console.log(chalk.red(`  Допустимые статусы: ${valid.join(", ")}`));
    return;
  }
  const p = readProject(CWD);
  if (!p) { console.log(chalk.red("  Нет .drift/ — запусти drift init")); return; }

  p.status = status as ProjectStatus;
  p.lastActivity = now();
  writeProject(CWD, p);
  syncToRegistry(CWD, p);
  console.log(`  ${chalk.green("✓")} Status: ${statusIcon(status)} ${status}`);
}

export function cmdDescribe(text: string) {
  const p = readProject(CWD);
  if (!p) { console.log(chalk.red("  Нет .drift/ — запусти drift init")); return; }

  p.description = text;
  p.lastActivity = now();
  writeProject(CWD, p);
  syncToRegistry(CWD, p);
  console.log(`  ${chalk.green("✓")} Description updated`);
}

export function cmdTag(tags: string[]) {
  const p = readProject(CWD);
  if (!p) { console.log(chalk.red("  Нет .drift/ — запусти drift init")); return; }

  for (const tag of tags) {
    if (!p.tags.includes(tag)) p.tags.push(tag);
  }
  p.lastActivity = now();
  writeProject(CWD, p);
  syncToRegistry(CWD, p);
  console.log(`  ${chalk.green("✓")} Tags: ${p.tags.map(t => chalk.cyan(t)).join("  ")}`);
}

export function cmdLink(type: string, url: string) {
  const p = readProject(CWD);
  if (!p) { console.log(chalk.red("  Нет .drift/ — запусти drift init")); return; }

  p.links[type] = url;
  p.lastActivity = now();
  writeProject(CWD, p);
  syncToRegistry(CWD, p);
  console.log(`  ${chalk.green("✓")} ${type}: ${chalk.cyan(url)}`);
}

const STATUS_ORDER: Record<string, number> = {
  active: 0, idea: 1, paused: 2, done: 3, abandoned: 4,
};

type SortMode = "activity" | "progress" | "name" | "status";

function sortProjectsCli(projects: ReturnType<typeof loadAllProjects>, mode: SortMode) {
  const sorted = [...projects];
  switch (mode) {
    case "activity": return sorted.sort((a, b) => b.lastActivity.localeCompare(a.lastActivity));
    case "progress": return sorted.sort((a, b) => b.progress - a.progress || a.name.localeCompare(b.name));
    case "name": return sorted.sort((a, b) => a.name.localeCompare(b.name));
    case "status": return sorted.sort((a, b) => (STATUS_ORDER[a.status] ?? 9) - (STATUS_ORDER[b.status] ?? 9) || b.lastActivity.localeCompare(a.lastActivity));
    default: return sorted;
  }
}

export function cmdList(sort: SortMode = "activity") {
  const registry = readRegistry();
  const projects = sortProjectsCli(loadAllProjects(registry), sort);

  if (projects.length === 0) {
    console.log(chalk.dim("\n  No projects tracked yet."));
    console.log(`  Run ${chalk.cyan("drift init")} in a project directory.\n`);
    return;
  }

  console.log(`\n${chalk.cyan.bold("  drift")}  ${chalk.dim(`— ${projects.length} projects — sort: ${sort}`)}\n`);

  for (const p of projects) {
    const icon = statusIcon(p.status);
    const name = p.name.padEnd(22);
    const pct = String(p.progress).padStart(3) + "%";
    const bar = miniBar(p.progress);
    const last = p.lastActivity ? chalk.dim(timeSince(p.lastActivity)) : "";
    const miss = p._missing ? chalk.red(" [missing]") : "";
    console.log(`  ${icon} ${name} ${bar} ${pct}  ${last}${miss}`);
  }
  console.log();
}

export function cmdOpen(name: string) {
  const registry = readRegistry();
  const matches = registry.projects.filter(p =>
    p.name.toLowerCase().includes(name.toLowerCase())
  );
  if (matches.length === 0) {
    console.log(chalk.red(`  Проект "${name}" не найден`));
    return;
  }
  if (matches.length === 1) {
    console.log(matches[0].path);
    return;
  }
  console.log(chalk.dim("  Несколько совпадений:"));
  matches.forEach(m => console.log(`  ${m.name.padEnd(20)} ${chalk.dim(m.path)}`));
}

export function cmdScan(dir?: string) {
  const root = resolve(dir ?? CWD);
  const found: { path: string; tags: string[] }[] = [];

  const entries = readdirSync(root, { withFileTypes: true });
  for (const entry of entries) {
    if (!entry.isDirectory() || entry.name.startsWith(".") || entry.name === "node_modules") continue;
    const full = join(root, entry.name);
    const isProject = existsSync(join(full, ".git")) ||
      existsSync(join(full, "package.json")) ||
      existsSync(join(full, "pyproject.toml")) ||
      existsSync(join(full, "Cargo.toml")) ||
      existsSync(join(full, "go.mod"));
    if (isProject && !hasProject(full)) {
      found.push({ path: full, tags: detectTags(full) });
    }
  }

  if (found.length === 0) {
    console.log(chalk.dim(`\n  No new projects found in ${root}\n`));
    return;
  }

  console.log(`\n  ${chalk.cyan.bold("drift scan")} — found ${chalk.bold(String(found.length))} new projects in ${chalk.dim(root.replace(process.env.HOME ?? "", "~"))}\n`);
  for (const f of found) {
    const name = basename(f.path).padEnd(24);
    const tags = f.tags.length ? chalk.dim(f.tags.join(", ")) : chalk.dim("—");
    console.log(`  ${name} ${tags}`);
  }
  console.log(chalk.dim(`\n  Run drift init from each project, or: drift scan --init\n`));

  return found;
}

export function cmdScanInit(dir?: string) {
  const root = resolve(dir ?? CWD);
  const found = cmdScan(dir);
  if (!found || found.length === 0) return;

  console.log(chalk.cyan("  Initializing all...\n"));
  for (const f of found) {
    const project = createProject(f.path);
    project.tags = f.tags;
    project.links.repo = detectRepo(f.path);
    writeProject(f.path, project);
    syncToRegistry(f.path, project);
    addToGitignore(f.path);
    console.log(`  ${chalk.green("✓")} ${project.name}`);
  }
  console.log();
}

function printGoals(p: Project) {
  p.goals.forEach((g, i) => {
    const icon = g.done ? chalk.green("✓") : chalk.dim("○");
    const text = g.done ? chalk.dim.strikethrough(g.text) : g.text;
    console.log(`  ${icon} ${chalk.dim(`${i + 1}.`)} ${text}`);
  });
  console.log(`\n  Progress: ${miniBar(p.progress)} ${p.progress}%`);
}

export function printHelp() {
  console.log(`
${chalk.cyan.bold("  ◆  d r i f t  ◆")}
${chalk.dim("  Vibe-coding project manager")}

${chalk.bold("Commands:")}
  ${chalk.cyan("drift init")}                 Initialize project
  ${chalk.cyan("drift status")}               Show current project
  ${chalk.cyan("drift list")}                 List all projects
  ${chalk.cyan("drift note")} ${chalk.dim('"text"')}           Add a note
  ${chalk.cyan("drift goal")} ${chalk.dim('"text"')}           Add a goal
  ${chalk.cyan("drift goal done")} ${chalk.dim("N")}          Mark goal #N done
  ${chalk.cyan("drift progress")} ${chalk.dim("N")}           Set progress (0-100)
  ${chalk.cyan("drift set-status")} ${chalk.dim("STATUS")}    Change status
  ${chalk.cyan("drift describe")} ${chalk.dim('"text"')}      Set description
  ${chalk.cyan("drift tag")} ${chalk.dim("tag1 tag2")}        Add tags
  ${chalk.cyan("drift link")} ${chalk.dim("type url")}        Set a link
  ${chalk.cyan("drift scan")} ${chalk.dim("[dir]")}           Find untracked projects
  ${chalk.cyan("drift scan --init")}          Find & init all
  ${chalk.cyan("drift open")} ${chalk.dim("name")}            Get project path
  ${chalk.cyan("drift")}                      Interactive TUI

${chalk.dim("Statuses: idea, active, paused, done, abandoned")}
`);
}
