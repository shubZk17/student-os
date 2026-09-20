const BASE_URL = `${import.meta.env.VITE_API_URL ?? ''}/api/v1`;

const ACCESS_KEY = 'studentos_access_token';
const REFRESH_KEY = 'studentos_refresh_token';
const USER_KEY = 'studentos_user';

export function saveSession(user: unknown, tokens: { access_token: string; refresh_token: string }) {
  localStorage.setItem(USER_KEY, JSON.stringify(user));
  localStorage.setItem(ACCESS_KEY, tokens.access_token);
  localStorage.setItem(REFRESH_KEY, tokens.refresh_token);
}

export function clearSession() {
  localStorage.removeItem(USER_KEY);
  localStorage.removeItem(ACCESS_KEY);
  localStorage.removeItem(REFRESH_KEY);
}

// Refresh tokens are single-use, so concurrent 401s must share one refresh call.
let refreshing: Promise<boolean> | null = null;

function refreshTokens(): Promise<boolean> {
  const refreshToken = localStorage.getItem(REFRESH_KEY);
  if (!refreshToken) return Promise.resolve(false);
  refreshing ??= fetch(`${BASE_URL}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken }),
  })
    .then(async (res) => {
      if (!res.ok) return false;
      const data = await res.json();
      localStorage.setItem(ACCESS_KEY, data.tokens.access_token);
      localStorage.setItem(REFRESH_KEY, data.tokens.refresh_token);
      return true;
    })
    .catch(() => false)
    .finally(() => {
      refreshing = null;
    });
  return refreshing;
}

function send(endpoint: string, options: RequestInit) {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };
  const token = localStorage.getItem(ACCESS_KEY);
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return fetch(`${BASE_URL}${endpoint}`, { ...options, headers });
}

export async function apiRequest<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  let response = await send(endpoint, options);

  if (response.status === 401 && !endpoint.startsWith('/auth/')) {
    if (await refreshTokens()) {
      response = await send(endpoint, options);
    } else {
      clearSession();
      window.location.assign('/login');
      throw new Error('Session expired. Please sign in again.');
    }
  }

  if (!response.ok) {
    let errMsg = `Request failed with status ${response.status}`;
    try {
      const errData = await response.json();
      if (errData.error) errMsg = errData.error;
    } catch (_) {}
    throw new Error(errMsg);
  }

  return response.json();
}
