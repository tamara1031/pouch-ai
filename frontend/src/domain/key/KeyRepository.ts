import type { Key } from './Key';

export interface CreateKeyDTO {
    name: string;
    provider: string;
    budget_limit: number;
    budget_period: string;
    rate_limit: number;
    rate_period: string;
    is_mock: boolean;
    mock_config: string;
    expires_at?: number;
}

export interface UpdateKeyDTO {
    name: string;
    provider: string;
    budget_limit: number;
    rate_limit: number;
    rate_period: string;
    is_mock: boolean;
    mock_config: string;
    expires_at?: number;
}

export interface KeyRepository {
    getAll(): Promise<Key[]>;
    create(data: CreateKeyDTO): Promise<{ key: string }>;
    update(id: number, data: UpdateKeyDTO): Promise<void>;
    delete(id: number): Promise<void>;
    getProviders(): Promise<string[]>;
    getProviderUsage(): Promise<Record<string, number>>;
}
