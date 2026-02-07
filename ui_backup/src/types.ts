export interface Key {
    id: number;
    name: string;
    provider: string;
    prefix: string;
    expires_at: number | null;
    budget_limit: number;
    budget_usage: number;
    budget_period: string;
    is_mock: boolean;
    mock_config: string;
    rate_limit: number;
    rate_period: string;
    created_at: number;
}
