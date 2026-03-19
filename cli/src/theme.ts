import chalk from "chalk";

// drift color palette — minimal, modern, dark-native
export const t = {
  // Primary
  accent: chalk.cyan,
  accentBold: chalk.cyan.bold,
  accentDim: chalk.dim.cyan,

  // Status colors
  active: chalk.cyan,
  done: chalk.green,
  idea: chalk.yellow,
  paused: chalk.dim,
  abandoned: chalk.dim.strikethrough,

  // Semantic
  bold: chalk.bold,
  dim: chalk.dim,
  muted: chalk.gray,
  bright: chalk.white,
  error: chalk.red,
  warn: chalk.yellow,

  // UI elements
  border: chalk.dim,
  selected: chalk.bgGray.white,
  label: chalk.dim,
  value: chalk.white,
  tag: chalk.cyan.dim,
  timestamp: chalk.dim,

  // Box drawing — rounded corners
  box: {
    tl: "╭",
    tr: "╮",
    bl: "╰",
    br: "╯",
    h: "─",
    v: "│",
  },
} as const;

export const STATUS_ICONS: Record<string, string> = {
  active: "●",
  done: "✓",
  idea: "○",
  paused: "◊",
  abandoned: "✗",
};

export const STATUS_COLORS: Record<string, (s: string) => string> = {
  active: t.active,
  done: t.done,
  idea: t.idea,
  paused: t.paused,
  abandoned: t.abandoned,
};

export function statusIcon(status: string): string {
  const icon = STATUS_ICONS[status] ?? "?";
  const color = STATUS_COLORS[status] ?? t.dim;
  return color(icon);
}

export function progressBar(pct: number, width = 20): string {
  const filled = Math.round((pct / 100) * width);
  const empty = width - filled;
  const bar = "█".repeat(filled) + "░".repeat(empty);

  if (pct >= 100) return t.done(bar);
  if (pct >= 60) return t.accent(bar);
  if (pct >= 30) return t.idea(bar);
  return t.dim(bar);
}
