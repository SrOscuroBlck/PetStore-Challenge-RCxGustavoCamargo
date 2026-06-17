#!/usr/bin/env node
/**
 * Stop hook. Refuses to let a turn finish with a broken type check when TS files were
 * touched. This is the backstop: nothing "done" ships with type errors.
 * Self-gating — no TS changes, not scaffolded, or no typecheck available => no-op.
 */
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

let payload;
try {
  payload = JSON.parse(fs.readFileSync(0, 'utf8') || '{}');
} catch {
  process.exit(0);
}

if (payload.stop_hook_active) process.exit(0); // prevent re-trigger loops

const root = process.env.CLAUDE_PROJECT_DIR || process.cwd();
const pkgPath = path.join(root, 'package.json');
if (!fs.existsSync(pkgPath)) process.exit(0);

// Only gate when there are uncommitted TS/TSX changes worth checking.
let changed = '';
try {
  changed = execSync('git status --porcelain', { cwd: root, stdio: ['ignore', 'pipe', 'ignore'] }).toString();
} catch {
  process.exit(0);
}
if (!/\.(ts|tsx)(\s|$)/m.test(changed)) process.exit(0);

let pkg;
try {
  pkg = JSON.parse(fs.readFileSync(pkgPath, 'utf8'));
} catch {
  process.exit(0);
}
const hasTypecheck = pkg.scripts && pkg.scripts.typecheck;
const tscBin = path.join(root, 'node_modules', '.bin', 'tsc');
if (!hasTypecheck && !fs.existsSync(tscBin)) process.exit(0);

try {
  execSync(hasTypecheck ? 'npm run -s typecheck' : `"${tscBin}" --noEmit`, {
    cwd: root,
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  process.exit(0);
} catch (e) {
  const out = `${e.stdout ? e.stdout.toString() : ''}${e.stderr ? e.stderr.toString() : ''}`.trim();
  process.stderr.write(`\n⛔ Type check failed — resolve before finishing:\n\n${out.slice(-3000)}\n`);
  process.exit(2);
}
