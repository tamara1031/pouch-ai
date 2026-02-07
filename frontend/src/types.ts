export interface PluginConfig {
    id: string;
    config: Record<string, string>;
}

export interface KeyConfiguration {
    provider: PluginConfig;
    middlewares: PluginConfig[];
}

export type FieldType = "string" | "number" | "boolean" | "select";

export interface FieldSchema {
    type: FieldType;
    default?: string;
    description?: string;
    options?: string[];
}

export type MiddlewareSchema = Record<string, FieldSchema>;

export interface MiddlewareInfo {
    id: string;
    schema: MiddlewareSchema;
}

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
    configuration?: KeyConfiguration;
}
