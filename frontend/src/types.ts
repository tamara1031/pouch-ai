export interface PluginConfig {
    id: string;
    config: Record<string, any>;
}

export interface KeyConfiguration {
    provider: PluginConfig;
    middlewares: PluginConfig[];
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

export type MiddlewareSchema = Record<string, FieldSchema>;

export interface MiddlewareInfo {
    id: string;
    schema: MiddlewareSchema;
    is_default?: boolean;
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
