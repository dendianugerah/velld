import { Auth, AuthResponse, Profile } from "@/types/auth";
import { apiRequest } from "@/lib/api-client";

export async function login(auth: Auth): Promise<AuthResponse> {
  const response = await apiRequest<AuthResponse>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify(auth),
  });
  localStorage.setItem('token', response.data.token);
  return response;
}

export async function getProfile(): Promise<Profile> {
  return apiRequest<Profile>('/api/auth/profile');
}