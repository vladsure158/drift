import { readFileSync, writeFileSync, existsSync, mkdirSync } from "node:fs";
import { join, basename } from "node:path";
import { homedir } from "node:os";
import { v4 as uuidv4 } from "uuid";

export interface Goal {
  text: string;
  done: boolean;
}

export interface Note {
  ts: string;
  text: string;
}

export interface Links {
  repo: string | null;
  deploy: string | null;
  design: string | null;
  [key: string]: string | null;
}

export type ProjectStatus = "idea" | "active" | "paused" | "done" | "abandoned";

export interface Project {
  id: string;
  name: string;
  description: string | null;
  status: ProjectStatus;
  progress: number;
  tags: string[];
  created: string;
  lastActivity: string;
  goals: Goal[];
  notes: Note[];
  links: Links;
  [key: string]: unknown; // Preserve unknown fields
}

export interface RegistryEntry {
  id: string;
  path: string;
  name: string;
  status: ProjectStatus;
  lastActivity: string;
}

export interface Registry {
  version: number;
  projects: RegistryEntry[];
}

const REGISTRY_DIR = join(homedir(), ".drift");
const REGISTRY_PATH = join(REGISTRY_DIR, "registry.json");

export function now(): string {
  return new Date().toISOString().replace(/\.\d{3}Z$/, "Z");
}

export function projectPath(projectRoot: string): string {
  return join(projectRoot, ".drift", "project.json");
}

export function hasProject(projectRoot: string): boolean {
  return existsSync(projectPath(projectRoot));
}

export function readProject(projectRoot: string): Project | null {
  const path = projectPath(projectRoot);
  if (!existsSync(path)) return null;
  return JSON.parse(readFileSync(path, "utf-8"));
}

export function writeProject(projectRoot: string, project: Project): void {
  const dir = join(projectRoot, ".drift");
  if (!existsSync(dir)) mkdirSync(dir, { recursive: true });
  writeFileSync(projectPath(projectRoot), JSON.stringify(project, null, 2) + "\n");
}

export function readRegistry(): Registry {
  if (!existsSync(REGISTRY_PATH)) {
    return { version: 1, projects: [] };
  }
  return JSON.parse(readFileSync(REGISTRY_PATH, "utf-8"));
}

export function writeRegistry(registry: Registry): void {
  if (!existsSync(REGISTRY_DIR)) mkdirSync(REGISTRY_DIR, { recursive: true });
  writeFileSync(REGISTRY_PATH, JSON.stringify(registry, null, 2) + "\n");
}

export function syncToRegistry(projectRoot: string, project: Project): void {
  const registry = readRegistry();
  const idx = registry.projects.findIndex((p) => p.id === project.id);
  const entry: RegistryEntry = {
    id: project.id,
    path: projectRoot,
    name: project.name,
    status: project.status,
    lastActivity: project.lastActivity,
  };
  if (idx >= 0) {
    registry.projects[idx] = entry;
  } else {
    registry.projects.push(entry);
  }
  writeRegistry(registry);
}

export function calcProgress(goals: Goal[]): number {
  if (goals.length === 0) return 0;
  const done = goals.filter((g) => g.done).length;
  return Math.round((done / goals.length) * 100);
}

export function createProject(projectRoot: string): Project {
  const project: Project = {
    id: uuidv4(),
    name: basename(projectRoot),
    description: null,
    status: "active",
    progress: 0,
    tags: [],
    created: now(),
    lastActivity: now(),
    goals: [],
    notes: [],
    links: { repo: null, deploy: null, design: null },
  };
  return project;
}

export function loadAllProjects(registry: Registry): (Project & { _path: string; _missing: boolean })[] {
  return registry.projects.map((entry) => {
    const project = readProject(entry.path);
    if (project) {
      return { ...project, _path: entry.path, _missing: false };
    }
    return {
      id: entry.id,
      name: entry.name,
      description: null,
      status: entry.status,
      progress: 0,
      tags: [],
      created: "",
      lastActivity: entry.lastActivity,
      goals: [],
      notes: [],
      links: { repo: null, deploy: null, design: null },
      _path: entry.path,
      _missing: true,
    };
  }).sort((a, b) => b.lastActivity.localeCompare(a.lastActivity));
}
