let API_BASE: string | null = null;

async function getConfig() {
  if (!API_BASE) {
    const { apiUrl } = await fetch('/api/config').then(res => res.json());
    API_BASE = apiUrl;
  }
  return API_BASE;
}

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

interface ExtendedRequestInit extends RequestInit {
  responseType?: 'json' | 'blob';
}

export async function apiRequest<T>(
  endpoint: string,
  options: ExtendedRequestInit = {}
): Promise<T> {
  const apiUrl = await getConfig();
  const token = localStorage.getItem('token');
  const headers = {
    'Content-Type': 'application/json',
    ...(token && { Authorization: `Bearer ${token}` }),
    ...options.headers,
  };

  const response = await fetch(`${apiUrl}${endpoint}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    if (response.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    const error = await response.text();
    throw new ApiError(response.status, error || `HTTP error! status: ${response.status}`);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  if (options.responseType === 'blob') {
    return response.blob() as T;
  }

  return response.json();
}
