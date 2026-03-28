// ComplianceForge — Auth Utilities
// Standalone token management and authentication helpers.
// These complement the ApiClient's built-in token handling and the Zustand auth store.

const TOKEN_KEY = 'cf_token';

/**
 * Retrieve the stored JWT token from localStorage.
 */
export function getToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(TOKEN_KEY);
}

/**
 * Persist a JWT token to localStorage.
 */
export function setToken(token: string): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem(TOKEN_KEY, token);
}

/**
 * Remove the JWT token from localStorage.
 */
export function clearToken(): void {
  if (typeof window === 'undefined') return;
  localStorage.removeItem(TOKEN_KEY);
}

/**
 * Check whether the user currently has a stored token.
 * Note: This does NOT verify the token is valid or unexpired.
 */
export function isAuthenticated(): boolean {
  return getToken() !== null;
}

/**
 * Build an Authorization header object for use with fetch or other HTTP clients.
 * Returns an empty object if no token is present.
 */
export function getAuthHeader(): Record<string, string> {
  const token = getToken();
  if (!token) return {};
  return { Authorization: `Bearer ${token}` };
}

/**
 * Full logout: clears the token, wipes any cached data from sessionStorage,
 * and redirects to the login page.
 */
export function logout(): void {
  clearToken();

  // Clear any session-level caches
  if (typeof window !== 'undefined') {
    try {
      sessionStorage.clear();
    } catch {
      // sessionStorage may not be available in some environments
    }
    window.location.href = '/auth/login';
  }
}
