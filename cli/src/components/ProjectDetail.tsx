import React from "react";
import { Text, Box } from "ink";
import { formatDistanceToNow } from "date-fns";
import type { Project } from "../lib/protocol.js";

interface ProjectDetailProps {
  project: (Project & { _path: string; _missing: boolean }) | null;
}

function progressBar(pct: number): string {
  const w = 20;
  const filled = Math.round((pct / 100) * w);
  return "█".repeat(filled) + "░".repeat(w - filled);
}

function relativeTime(iso: string): string {
  if (!iso) return "—";
  try {
    return formatDistanceToNow(new Date(iso), { addSuffix: true });
  } catch {
    return iso;
  }
}

export function ProjectDetail({ project }: ProjectDetailProps) {
  if (!project) {
    return (
      <Box flexDirection="column" paddingY={1}>
        <Text dimColor>Select a project</Text>
      </Box>
    );
  }

  const p = project;
  const doneGoals = p.goals.filter((g) => g.done).length;
  const totalGoals = p.goals.length;
  const recentNotes = p.notes.slice(-5).reverse();

  return (
    <Box flexDirection="column">
      {/* Name & Description */}
      <Text bold color="white">{p.name}</Text>
      {p.description ? (
        <Text dimColor>{p.description}</Text>
      ) : (
        <Text color="gray" italic>no description</Text>
      )}

      <Text>{""}</Text>

      {/* Progress */}
      <Box>
        <Text color={p.progress >= 100 ? "green" : "cyan"}>
          {progressBar(p.progress)}
        </Text>
        <Text bold> {p.progress}%</Text>
      </Box>

      <Text>{""}</Text>

      {/* Tags */}
      {p.tags.length > 0 && (
        <Box gap={1} flexWrap="wrap">
          {p.tags.map((tag) => (
            <Text key={tag} color="cyan" dimColor>{tag}</Text>
          ))}
        </Box>
      )}

      {/* Last activity */}
      <Box marginTop={1}>
        <Text dimColor>Last: </Text>
        <Text>{relativeTime(p.lastActivity)}</Text>
      </Box>

      {/* Path */}
      <Box>
        <Text dimColor>Path: </Text>
        <Text dimColor>{p._path.replace(process.env.HOME ?? "", "~")}</Text>
      </Box>

      {/* Goals */}
      {totalGoals > 0 && (
        <Box flexDirection="column" marginTop={1}>
          <Box>
            <Text bold>Goals</Text>
            <Text dimColor> {doneGoals}/{totalGoals}</Text>
          </Box>
          {p.goals.map((g, i) => (
            <Box key={i}>
              <Text color={g.done ? "green" : "gray"}>
                {g.done ? "  ✓ " : "  ○ "}
              </Text>
              <Text dimColor={g.done} strikethrough={g.done}>
                {g.text}
              </Text>
            </Box>
          ))}
        </Box>
      )}

      {/* Notes */}
      {recentNotes.length > 0 && (
        <Box flexDirection="column" marginTop={1}>
          <Text bold>Notes</Text>
          {recentNotes.map((n, i) => {
            const time = n.ts.slice(11, 16); // HH:MM
            return (
              <Box key={i} gap={1}>
                <Text dimColor>{time}</Text>
                <Text>{n.text}</Text>
              </Box>
            );
          })}
          {p.notes.length > 5 && (
            <Text dimColor>  +{p.notes.length - 5} more</Text>
          )}
        </Box>
      )}

      {/* Links */}
      {Object.entries(p.links).some(([, v]) => v) && (
        <Box flexDirection="column" marginTop={1}>
          <Text bold>Links</Text>
          {Object.entries(p.links).map(([key, url]) =>
            url ? (
              <Box key={key} gap={1}>
                <Text dimColor>{key}:</Text>
                <Text color="cyan" underline>{url}</Text>
              </Box>
            ) : null
          )}
        </Box>
      )}
    </Box>
  );
}
