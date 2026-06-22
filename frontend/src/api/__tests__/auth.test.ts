import { describe, it, expect } from 'vitest';
import { decodeToken } from '../auth';

// Helper to create a valid-looking JWT with a given payload
function makeToken(payload: Record<string, unknown>): string {
  const body = btoa(JSON.stringify(payload));
  const bodyUrlSafe = body.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  return `eyJhbGciOiJIUzI1NiJ9.${bodyUrlSafe}.fake-sig`;
}

describe('decodeToken', () => {
  it('decodes a valid JWT token with all fields', () => {
    const token = makeToken({
      user_id: 1,
      username: 'admin',
      role_id: 1,
      role_name: 'super_admin',
    });

    const result = decodeToken(token);
    expect(result).toEqual({
      user_id: 1,
      username: 'admin',
      role_id: 1,
      role_name: 'super_admin',
    });
  });

  it('decodes a token with role_name containing underscore (base64url)', () => {
    const token = makeToken({
      user_id: 2,
      username: 'lab_tech',
      role_id: 2,
      role_name: 'equipment_manager',
    });

    const result = decodeToken(token);
    expect(result).not.toBeNull();
    expect(result!.role_name).toBe('equipment_manager');
    expect(result!.username).toBe('lab_tech');
  });

  it('returns null for an empty string', () => {
    expect(decodeToken('')).toBeNull();
  });

  it('returns null for malformed JWT (no dots)', () => {
    expect(decodeToken('justonestring')).toBeNull();
  });

  it('returns null for a token with invalid base64 payload', () => {
    // The payload segment is not valid base64url
    expect(decodeToken('a.b@d!.c')).toBeNull();
  });

  it('decodes numeric user_id correctly (not as string)', () => {
    const token = makeToken({
      user_id: 42,
      username: 'test',
      role_id: 3,
      role_name: 'member',
    });

    const result = decodeToken(token);
    expect(typeof result!.user_id).toBe('number');
    expect(result!.user_id).toBe(42);
  });
});
