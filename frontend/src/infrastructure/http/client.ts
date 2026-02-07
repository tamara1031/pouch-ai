export interface RequestOptions extends RequestInit {
    headers?: Record<string, string>;
}

export class HttpClient {
    constructor(private baseURL: string = '/v1') { }

    async request<T>(endpoint: string, options?: RequestOptions): Promise<T> {
        const url = `${this.baseURL}${endpoint}`;
        const headers = {
            'Content-Type': 'application/json',
            ...(options?.headers || {}),
        };

        const config: RequestInit = {
            ...options,
            headers,
        };

        try {
            const response = await fetch(url, config);

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || `API request failed: ${response.statusText}`);
            }

            // Handle 204 No Content
            if (response.status === 204) {
                return {} as T;
            }

            return response.json();
        } catch (error) {
            console.error(`HTTP Request failed for ${url}:`, error);
            throw error;
        }
    }

    get<T>(endpoint: string, options?: RequestOptions): Promise<T> {
        return this.request<T>(endpoint, { ...options, method: 'GET' });
    }

    post<T>(endpoint: string, body: any, options?: RequestOptions): Promise<T> {
        return this.request<T>(endpoint, { ...options, method: 'POST', body: JSON.stringify(body) });
    }

    put<T>(endpoint: string, body: any, options?: RequestOptions): Promise<T> {
        return this.request<T>(endpoint, { ...options, method: 'PUT', body: JSON.stringify(body) });
    }

    delete<T>(endpoint: string, options?: RequestOptions): Promise<T> {
        return this.request<T>(endpoint, { ...options, method: 'DELETE' });
    }
}

export const httpClient = new HttpClient();
