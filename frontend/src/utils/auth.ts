/**
 * Authentication Utilities
 * Handles JWT token management and authenticated fetch requests
 */

const TOKEN_KEY = 'ares_jwt_token';
const USER_KEY = 'ares_user';

/**
 * Get the stored JWT token from localStorage
 */
export function getToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(TOKEN_KEY);
}

/**
 * Store a JWT token in localStorage
 */
export function setToken(token: string): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem(TOKEN_KEY, token);
}

/**
 * Remove the JWT token from localStorage
 */
export function clearToken(): void {
  if (typeof window === 'undefined') return;
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
}

/**
 * Check if user is authenticated
 */
export function isAuthenticated(): boolean {
  return getToken() !== null;
}

/**
 * Store user information
 */
export function setUser(user: any): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}

/**
 * Get stored user information
 */
export function getUser(): any | null {
  if (typeof window === 'undefined') return null;
  const userStr = localStorage.getItem(USER_KEY);
  if (!userStr) return null;
  try {
    return JSON.parse(userStr);
  } catch {
    return null;
  }
}

/**
 * Perform an authenticated fetch request
 * Automatically adds Authorization header if token exists
 */
export async function fetchWithAuth(
  url: string,
  options: RequestInit = {}
): Promise<Response> {
  const token = getToken();
  
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  // Add Authorization header if token exists
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(url, {
    ...options,
    headers,
  });

  // Handle 401 Unauthorized - token expired or invalid
  if (response.status === 401) {
    clearToken();
    // Optionally redirect to login
    if (typeof window !== 'undefined') {
      window.location.href = '/login.html';
    }
  }

  return response;
}

/**
 * Login and store the token
 */
export async function login(username: string, password: string): Promise<{ success: boolean; error?: string }> {
  try {
    const response = await fetch('http://localhost:8080/api/v1/users/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    });

    if (!response.ok) {
      const error = await response.text();
      return { success: false, error };
    }

    const data = await response.json();
    
    if (data.token) {
      setToken(data.token);
      if (data.user) {
        setUser(data.user);
      }
      return { success: true };
    }

    return { success: false, error: 'No token received' };
  } catch (error) {
    return { success: false, error: String(error) };
  }
}

/**
 * Logout and clear stored credentials
 */
export function logout(): void {
  clearToken();
  if (typeof window !== 'undefined') {
    window.location.href = '/login.html';
  }
}
