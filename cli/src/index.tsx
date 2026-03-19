#!/usr/bin/env node
import React from "react";
import { render } from "ink";
import { App } from "./app.js";
import {
  cmdInit, cmdStatus, cmdList, cmdNote, cmdGoal, cmdGoalDone,
  cmdProgress, cmdSetStatus, cmdDescribe, cmdTag, cmdLink,
  cmdOpen, cmdScan, cmdScanInit, printHelp,
} from "./lib/commands.js";

const args = process.argv.slice(2);
const cmd = args[0];
const rest = args.slice(1).join(" ").replace(/^["']|["']$/g, "");

switch (cmd) {
  case "init":
    cmdInit(args[1]);
    break;

  case "status":
  case "s":
    cmdStatus();
    break;

  case "list":
  case "ls": {
    const sortFlag = args.find(a => a.startsWith("--sort="))?.split("=")[1] as any;
    cmdList(sortFlag ?? "activity");
    break;
  }

  case "note":
  case "n":
    if (!rest) { console.log("  Usage: drift note \"text\""); break; }
    cmdNote(rest);
    break;

  case "goal":
  case "g":
    if (args[1] === "done") {
      const n = parseInt(args[2], 10);
      if (isNaN(n)) { console.log("  Usage: drift goal done N"); break; }
      cmdGoalDone(n);
    } else {
      if (!rest) { console.log("  Usage: drift goal \"text\""); break; }
      cmdGoal(rest);
    }
    break;

  case "progress":
  case "p": {
    const n = parseInt(args[1], 10);
    if (isNaN(n)) { console.log("  Usage: drift progress N"); break; }
    cmdProgress(n);
    break;
  }

  case "set-status":
    if (!args[1]) { console.log("  Usage: drift set-status STATUS"); break; }
    cmdSetStatus(args[1]);
    break;

  case "describe":
  case "desc":
    if (!rest) { console.log("  Usage: drift describe \"text\""); break; }
    cmdDescribe(rest);
    break;

  case "tag":
    if (args.length < 2) { console.log("  Usage: drift tag tag1 tag2"); break; }
    cmdTag(args.slice(1));
    break;

  case "link":
    if (args.length < 3) { console.log("  Usage: drift link type url"); break; }
    cmdLink(args[1], args[2]);
    break;

  case "open":
  case "o":
    if (!args[1]) { console.log("  Usage: drift open name"); break; }
    cmdOpen(args[1]);
    break;

  case "scan":
    if (args.includes("--init")) {
      cmdScanInit(args.find(a => a !== "scan" && a !== "--init"));
    } else {
      cmdScan(args[1]);
    }
    break;

  case "help":
  case "-h":
  case "--help":
    printHelp();
    break;

  case "tui":
  case undefined:
    // No args: TUI or status
    if (process.stdin.isTTY) {
      process.on("SIGINT", () => process.exit(0));
      // Alternate screen = fullscreen, clean exit
      process.stdout.write("\x1b[?1049h"); // enter alternate screen
      process.stdout.write("\x1b[?25l");   // hide cursor
      const cleanup = () => {
        process.stdout.write("\x1b[?25h");   // show cursor
        process.stdout.write("\x1b[?1049l"); // leave alternate screen
      };
      process.on("exit", cleanup);
      render(<App />);
    } else {
      cmdList();
    }
    break;

  default:
    console.log(`  Unknown command: ${cmd}`);
    printHelp();
    break;
}
