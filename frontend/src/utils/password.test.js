import { describe, it, expect } from 'vitest';
import { passwordProblems, isPasswordValid, PASSWORD_RULES } from './password.js';

describe('password complexity (frontend mirror of backend)', () => {
  it('accepts a compliant password', () => {
    expect(isPasswordValid('Abcdefgh1j!k')).toBe(true);
    expect(passwordProblems('Abcdefgh1j!k')).toEqual([]);
  });

  it('flags each unmet rule', () => {
    expect(passwordProblems('abcdefgh1j!k')).toContain('upper');
    expect(passwordProblems('ABCDEFGH1J!K')).toContain('lower');
    expect(passwordProblems('Abcdefghij!k')).toContain('digit');
    expect(passwordProblems('Abcdefghij1k')).toContain('symbol');
    expect(passwordProblems('Abc1!')).toContain('length');
  });

  it('rejects passwords over 72 bytes', () => {
    expect(passwordProblems('Aa1!' + 'a'.repeat(70))).toContain('length');
  });

  it('exposes rules with labels for UI rendering', () => {
    expect(PASSWORD_RULES.map((r) => r.id).sort()).toEqual(['digit', 'length', 'lower', 'symbol', 'upper']);
  });
});
