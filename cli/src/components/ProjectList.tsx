import React from "react";
import { Text, Box } from "ink";
import type { Project } from "../lib/protocol.js";
import { STATUS_ICONS } from "../theme.js";

interface ProjectListProps {
  projects: (Project & { _path: string; _missing: boolean })[];
  selectedIndex: number;
  focused: boolean;
}

function statusColor(status: string): string {
  switch (status) {
    case "active": return "cyan";
    case "done": return "green";
    case "idea": return "yellow";
    case "paused": return "gray";
    case "abandoned": return "gray";
    default: return "white";
  }
}

function miniBar(pct: number): string {
  const w = 5;
  const filled = Math.round((pct / 100) * w);
  return "█".repeat(filled) + "░".repeat(w - filled);
}

export function ProjectList({ projects, selectedIndex, focused }: ProjectListProps) {
  if (projects.length === 0) {
    return (
      <Box flexDirection="column" paddingY={1}>
        <Text dimColor>No projects yet.</Text>
        <Text dimColor>Run `drift init` in a project.</Text>
      </Box>
    );
  }

  return (
    <Box flexDirection="column">
      {projects.map((p, i) => {
        const isSelected = i === selectedIndex;
        const icon = STATUS_ICONS[p.status] ?? "?";
        const color = statusColor(p.status);
        const bar = miniBar(p.progress);
        const nameWidth = 18;
        const name = p.name.length > nameWidth
          ? p.name.slice(0, nameWidth - 1) + "…"
          : p.name.padEnd(nameWidth);

        return (
          <Box key={p.id}>
            <Text
              color={isSelected && focused ? "cyan" : undefined}
              bold={isSelected}
              inverse={isSelected && focused}
            >
              {isSelected ? " ▸ " : "   "}
              <Text color={color}>{icon}</Text>
              {" "}
              {name}
              {" "}
              <Text color={color} dimColor={!isSelected}>
                {bar}
              </Text>
              <Text dimColor>
                {" " + String(p.progress).padStart(3) + "%"}
              </Text>
              {p._missing && <Text color="red"> !</Text>}
            </Text>
          </Box>
        );
      })}
    </Box>
  );
}
