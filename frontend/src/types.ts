export interface PluginConfig {
    id: string;
    config: Record<string, any>;
}

export interface KeyConfiguration {
    provider: PluginConfig;
    middlewares: PluginConfig[];
    budget_limit: number;
    reset_period: number;
}

export type FieldType = "string" | "number" | "boolean" | "select";

export type FieldRole = "limit" | "period";

export interface FieldSchema {
    type: FieldType;
    displayName?: string;
    default?: any;
    description?: string;
    options?: string[];
    role?: FieldRole;
}

export type PluginSchema = Record<string, FieldSchema>;

export interface MiddlewareInfo {
    id: string;
    schema: PluginSchema;
    is_default?: boolean;
}

export interface ProviderInfo {
    id: string;
    schema: PluginSchema;
}

export interface Key {
    id: number;
    name: string;
    prefix: string;
    expires_at: number | null;
    budget_usage: number;
    created_at: number;
    configuration: KeyConfiguration;
}
