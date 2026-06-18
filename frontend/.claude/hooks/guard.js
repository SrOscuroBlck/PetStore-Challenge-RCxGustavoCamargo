#!/usr/bin/env node
/**
 * PreToolUse guard (Edit|Write|MultiEdit).
 * Mechanically enforces the project's non-negotiables on every file the agent writes.
 * Each rule maps to a line in CLAUDE.md. Exit 2 blocks the edit and feeds the reason
 * back to Claude. Every rule has a documented escape hatch for the rare legitimate case.
 */
const fs = require('fs');

let payload;
try {
  payload = JSON.parse(fs.readFileSync(0, 'utf8') || '{}');
} catch {
  process.exit(0); // never break the workflow on a parse error
}

const input = payload.tool_input || {};
const filePath = input.file_path || '';

// Only scan source we author; skip generated code, deps, and the docs/ source-of-truth.
if (!/\.(ts|tsx|js|jsx|graphql|gql)$/.test(filePath)) process.exit(0);
if (/node_modules|\/dist\/|\.generated\.|\/generated\/|\/__generated__\/|\/gql\/|(^|\/)docs\//.test(filePath)) {
  process.exit(0);
}

// Gather the new content regardless of which edit tool was used.
let text = '';
if (typeof input.content === 'string') text += input.content + '\n';
if (typeof input.new_string === 'string') text += input.new_string + '\n';
if (Array.isArray(input.edits)) {
  for (const e of input.edits) if (e && typeof e.new_string === 'string') text += e.new_string + '\n';
}
if (!text.trim()) process.exit(0);

const lines = text.split('\n');
const problems = [];

function scan(re, msg, escape) {
  lines.forEach((line, i) => {
    if (escape && line.includes(escape)) return;
    if (re.test(line)) problems.push(`  L~${i + 1} — ${msg}\n      > ${line.trim().slice(0, 120)}`);
  });
}

// 1. Credentials/session must never touch localStorage (XSS exfiltration). CLAUDE.md §Security.
scan(
  /\blocalStorage\b/,
  'localStorage is banned — session/credential state uses sessionStorage or in-memory context. Escape (non-sensitive only): add `// safe-localStorage` to the line.',
  '// safe-localStorage'
);

// 2. No raw HTML injection on server/user data (XSS).
scan(
  /dangerouslySetInnerHTML/,
  'dangerouslySetInnerHTML risks XSS — rely on React escaping. Escape (trusted/sanitized only): `// safe-html`.',
  '// safe-html'
);

// 3. Customer-only app: merchant operations return FORBIDDEN and must never be referenced.
scan(
  /\b(createPet|removePet|soldPets|unsoldPets)\b/,
  'Merchant operation referenced — this app is customer-only. Use availablePets / purchasePet / checkout.'
);

// 4. PublicPet has no breeder fields; selecting them is a hard validation error.
scan(
  /\b(breederName|breederEmail)\b/,
  'Breeder PII field referenced — PublicPet omits breeder fields. Never query breeder data.'
);

// 5. Strict typing: no `any` on API data.
scan(
  /:\s*any\b|as\s+any\b|<any>/,
  'Explicit `any` — TS strict mode forbids `any` on API data. Type it properly. Escape: `// allow-any`.',
  '// allow-any'
);

// 6. Relay page cap: first must stay <= 100 or the server returns COMPLEXITY_LIMIT_EXCEEDED.
lines.forEach((line, i) => {
  const m = line.match(/\bfirst:\s*(\d+)/);
  if (m && Number(m[1]) > 100) {
    problems.push(`  L~${i + 1} — first: ${m[1]} exceeds the page cap of 100 (COMPLEXITY_LIMIT_EXCEEDED). Paginate with first<=100 + after.`);
  }
});

if (problems.length) {
  process.stderr.write(
    `\n⛔ Quality guardrail blocked this edit to ${filePath}:\n\n${problems.join('\n\n')}\n\n` +
      `These are project non-negotiables (see CLAUDE.md). Fix them, or use the documented escape hatch if this is a genuine exception.\n`
  );
  process.exit(2);
}
process.exit(0);
