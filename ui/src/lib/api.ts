import type { Key, CreateKeyInput, UpdateKeyInput, ProviderUsage } from './types';

const API_BASE = '/v1';

async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const res = await fetch(`${API_BASE}${endpoint}`, options);
    if (!res.ok) {
        const errorText = await res.text();
        throw new Error(errorText || `API check failed: ${res.statusText}`);
    }
    // For 204 No Content
    if (res.status === 204) {
        return {} as T;
    }
    return res.json();
}

export const api = {
    getKeys: () => request<Key[]>('/config/app-keys'),

    createKey: (data: CreateKeyInput) => request<{ key: string }>('/config/app-keys', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
    }),

    updateKey: (id: number, data: UpdateKeyInput) => request<void>(`/config/app-keys/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
    }),

    deleteKey: (id: number) => request<void>(`/config/app-keys/${id}`, {
        method: 'DELETE',
    }),

    getProviders: async () => {
        const res = await request<{ providers: string[] }>('/config/providers');
        return res.providers;
    },

    getProviderUsage: () => request<ProviderUsage>('/config/providers/usage'),
};
