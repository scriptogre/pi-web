import { describe, expect, it } from 'vitest';
import { formatShellCommand } from './shell-format.js';

describe('formatShellCommand', () => {
  it('breaks a long one-liner at top-level operators', () => {
    expect(formatShellCommand('cd /tmp/some/long/project/path && git status --short | head -3; ls -la /tmp')).toBe(
      'cd /tmp/some/long/project/path\n  && git status --short\n  | head -3\nls -la /tmp',
    );
  });

  it('keeps operators inside quotes', () => {
    const command = `git commit -qm "Fix a; b && c" && echo 'x | y' && echo done-with-a-long-tail`;
    expect(formatShellCommand(command)).toBe(`git commit -qm "Fix a; b && c"\n  && echo 'x | y'\n  && echo done-with-a-long-tail`);
  });

  it('leaves short and multi-line commands alone', () => {
    expect(formatShellCommand('ls; pwd')).toBe('ls; pwd');
    const multi = 'cat <<EOF\na; b && c that is long enough to pass the length check\nEOF';
    expect(formatShellCommand(multi)).toBe(multi);
  });
});
