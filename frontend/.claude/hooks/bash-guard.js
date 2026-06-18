#!/usr/bin/env node
/**
 * PreToolUse guard (Bash). Stops secrets from being committed.
 * The challenge is confidential and credentials must never enter version control.
 */
const fs = require('fs');

let payload;
try {
  payload = JSON.parse(fs.readFileSync(0, 'utf8') || '{}');
} catch {
  process.exit(0);
}

const cmd = (payload.tool_input && payload.tool_input.command) || '';

// Block staging real env files; .env.example (value-less, documented) is allowed.
if (/git\s+add\b/.test(cmd) && /(^|[\s'"/])\.env(\.[\w.]+)?(\s|$|['"])/.test(cmd) && !/\.env\.example/.test(cmd)) {
  process.stderr.write(
    '⛔ Refusing to stage a .env file — secrets must never be committed. ' +
      'Document variables (without values) in .env.example instead.\n'
  );
  process.exit(2);
}

process.exit(0);
