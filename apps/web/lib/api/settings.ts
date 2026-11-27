import { ApiError, apiRequest } from "@/lib/api-client";
import { UpdateSettingsRequest, GetSettingsResponse, UpdateSettingsResponse } from "@/types/settings";

export async function getUserSettings(): Promise<GetSettingsResponse> {
  return apiRequest<GetSettingsResponse>('/api/settings');
}

export async function updateUserSettings(settings: UpdateSettingsRequest): Promise<UpdateSettingsResponse> {
  return apiRequest<UpdateSettingsResponse>('/api/settings', {
    method: 'PUT',
    body: JSON.stringify(settings),
  });
}

export async function testStorageConnection(settings: UpdateSettingsRequest): Promise<void> {
  try {
    await apiRequest('/api/settings/storage/test', {
      method: 'POST',
      body: JSON.stringify(settings),
    });
  } catch (error) {
    if (!(error instanceof ApiError)) throw error;
    let message: string | undefined;
    try {
      message = JSON.parse(error.message)?.message;
    } catch {
      // The server may return a non-JSON error response.
    }
    throw new Error(message || 'Unable to test the storage connection.');
  }
}
