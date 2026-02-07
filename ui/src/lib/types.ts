export interface Key {
    id: number;
    name: string;
    provider: string;
    key_hash: string;
    prefix: string;
    budget: {
        limit: number;
        period: string;
        usage: number;
    };
    rate_limit: {
        limit: number;
        period: string;
    };
    is_mock: boolean;
    mock_config: string;
    last_reset_at: string;
    created_at: string;
    expires_at?: string;
}

export interface CreateKeyInput {
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

export interface UpdateKeyInput {
    name: string;
    provider: string;
    budget_limit: number;
    rate_limit: number;
    rate_period: string;
    is_mock: boolean;
    mock_config: string;
    expires_at?: number;
}

export interface ProviderUsage {
    [provider: string]: number;
}
