#!/usr/bin/env node
/**
 * PostToolUse (Edit|Write|MultiEdit) on .ts/.tsx.
 * Auto-formats/fixes the changed file with the project's ESLint and surfaces anything
 * that --fix could not resolve, so problems are caught at write-time, not at the gate.
 * No-ops silently until the project is scaffolded (package.json + local eslint present).
 */
const fs = require('fs');
const path = require('path');
const { execFileSync } = require('child_process');

let payload;
try {
  payload = JSON.parse(fs.readFileSync(0, 'utf8') || '{}');
} catch {
  process.exit(0);
}

const filePath = (payload.tool_input && payload.tool_input.file_path) || '';
if (!/\.(ts|tsx)$/.test(filePath)) process.exit(0);
if (/node_modules|\/dist\/|\.generated\.|\/__generated__\//.test(filePath)) process.exit(0);

const root = process.env.CLAUDE_PROJECT_DIR || process.cwd();
if (!fs.existsSync(path.join(root, 'package.json'))) process.exit(0); // pre-scaffold

const eslintBin = path.join(root, 'node_modules', '.bin', 'eslint');
if (!fs.existsSync(eslintBin)) process.exit(0);

try {
  execFileSync(eslintBin, ['--fix', filePath], { cwd: root, stdio: 'pipe' });
  process.exit(0);
} catch (e) {
  const out = `${e.stdout ? e.stdout.toString() : ''}${e.stderr ? e.stderr.toString() : ''}`.trim();
  process.stderr.write(`ESLint issues remain in ${path.basename(filePath)} after auto-fix:\n${out.slice(0, 2000)}\n`);
  process.exit(2); // surface to Claude (the file is already written; this prompts a fix)
}
