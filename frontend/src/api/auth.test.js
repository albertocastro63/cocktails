import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { getToken, setToken, clearToken, isLoggedIn, isAdmin, getUserID } from './auth.js';

function makeToken(payload) {
  return `header.${btoa(JSON.stringify(payload))}.signature`;
}

describe('auth module', () => {
  beforeEach(() => sessionStorage.clear());
  afterEach(() => sessionStorage.clear());

  describe('getToken / setToken / clearToken', () => {
    it('getToken returns null when nothing is stored', () => {
      expect(getToken()).toBeNull();
    });

    it('setToken stores a value and getToken retrieves it', () => {
      setToken('my-jwt');
      expect(getToken()).toBe('my-jwt');
    });

    it('clearToken removes the stored token', () => {
      setToken('my-jwt');
      clearToken();
      expect(getToken()).toBeNull();
    });
  });

  describe('isLoggedIn', () => {
    it('returns false when no token is stored', () => {
      expect(isLoggedIn()).toBe(false);
    });

    it('returns true when a token is stored', () => {
      setToken('some-token');
      expect(isLoggedIn()).toBe(true);
    });
  });

  describe('isAdmin', () => {
    it('returns false when no token is stored', () => {
      expect(isAdmin()).toBe(false);
    });

    it('returns false for a malformed token (not 3 segments)', () => {
      setToken('not.valid');
      expect(isAdmin()).toBe(false);
    });

    it('returns false for a token whose payload is not valid base64', () => {
      setToken('header.!!!.sig');
      expect(isAdmin()).toBe(false);
    });

    it('returns true when payload has is_admin: true', () => {
      setToken(makeToken({ is_admin: true, user_id: 'u1' }));
      expect(isAdmin()).toBe(true);
    });

    it('returns false when payload has is_admin: false', () => {
      setToken(makeToken({ is_admin: false, user_id: 'u1' }));
      expect(isAdmin()).toBe(false);
    });

    it('returns false when payload lacks is_admin', () => {
      setToken(makeToken({ user_id: 'u1' }));
      expect(isAdmin()).toBe(false);
    });
  });

  describe('getUserID', () => {
    it('returns null when no token is stored', () => {
      expect(getUserID()).toBeNull();
    });

    it('returns null for a malformed token', () => {
      setToken('not.valid');
      expect(getUserID()).toBeNull();
    });

    it('returns the user_id from the payload', () => {
      setToken(makeToken({ user_id: 'abc123', is_admin: false }));
      expect(getUserID()).toBe('abc123');
    });

    it('returns null when payload lacks user_id', () => {
      setToken(makeToken({ is_admin: true }));
      expect(getUserID()).toBeNull();
    });
  });
});
