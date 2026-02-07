import { Key, type KeyProps } from '../../domain/key/Key';
import type { KeyRepository, CreateKeyDTO, UpdateKeyDTO } from '../../domain/key/KeyRepository';
import { httpClient } from '../http/client';

export class RestKeyRepository implements KeyRepository {
    async getAll(): Promise<Key[]> {
        const keysProps = await httpClient.get<KeyProps[]>('/config/app-keys');
        return keysProps.map(props => new Key(props));
    }

    async create(data: CreateKeyDTO): Promise<{ key: string }> {
        return httpClient.post<{ key: string }>('/config/app-keys', data);
    }

    async update(id: number, data: UpdateKeyDTO): Promise<void> {
        await httpClient.put(`/config/app-keys/${id}`, data);
    }

    async delete(id: number): Promise<void> {
        await httpClient.delete(`/config/app-keys/${id}`);
    }

    async getProviders(): Promise<string[]> {
        const res = await httpClient.get<{ providers: string[] }>('/config/providers');
        return res.providers;
    }

    async getProviderUsage(): Promise<Record<string, number>> {
        return httpClient.get<Record<string, number>>('/config/providers/usage');
    }
}
