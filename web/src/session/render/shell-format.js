const breaks = [' && ', ' || ', ' | ', '; '];

// Breaks a long one-line shell command at top-level ;, &&, || and | for display.
// Quoted text stays intact. Multi-line and short commands are returned unchanged.
export function formatShellCommand(command) {
  if (command.length <= 60 || command.includes('\n')) return command;
  let out = '';
  let quote = null;
  for (let i = 0; i < command.length; i++) {
    const ch = command[i];
    if (quote) {
      if (ch === quote) quote = null;
    } else if (ch === "'" || ch === '"') {
      quote = ch;
    } else {
      const op = breaks.find((b) => command.startsWith(b, i));
      if (op) {
        out += op === '; ' ? '\n' : `\n  ${op.trim()} `;
        i += op.length - 1;
        continue;
      }
    }
    out += ch;
  }
  return out;
}
