import type { Key, MiddlewareInfo, ProviderInfo, PluginConfig } from "../types";

const BASE_URL = "/v1";

async function request<T>(path: string, options?: RequestInit): Promise<T> {
    const res = await fetch(`${BASE_URL}${path}`, {
        ...options,
        headers: {
            "Content-Type": "application/json",
            ...options?.headers,
        },
    });

    if (!res.ok) {
        let errorMsg = `API error: ${res.status}`;
        try {
            const errorData = await res.json();
            errorMsg = errorData.error || errorMsg;
        } catch (_) {
            // Ignore parse error
        }
        throw new Error(errorMsg);
    }

    if (res.status === 204) return {} as T;
    return res.json();
}

export const api = {
    keys: {
        list: () => request<Key[]>("/config/app-keys", { cache: "no-store" }),
        create: (data: any) => request<{ key: string }>("/config/app-keys", {
            method: "POST",
            body: JSON.stringify(data),
        }),
        update: (id: number, data: any) => request<void>(`/config/app-keys/${id}`, {
            method: "PUT",
            body: JSON.stringify(data),
        }),
        delete: (id: number) => request<void>(`/config/app-keys/${id}`, {
            method: "DELETE",
        }),
    },
    plugins: {
        middlewares: () => request<{ middlewares: MiddlewareInfo[] }>("/config/middlewares", { cache: "no-store" }),
        providers: () => request<{ providers: ProviderInfo[] }>("/config/providers", { cache: "no-store" }),
    },
};
