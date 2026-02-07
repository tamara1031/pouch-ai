export interface KeyProps {
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

export class Key {
    constructor(private props: KeyProps) { }

    get id(): number { return this.props.id; }
    get name(): string { return this.props.name; }
    get provider(): string { return this.props.provider; }
    get keyHash(): string { return this.props.key_hash; }
    get prefix(): string { return this.props.prefix; }
    get budget(): KeyProps['budget'] { return this.props.budget; }
    get rateLimit(): KeyProps['rate_limit'] { return this.props.rate_limit; }
    get isMock(): boolean { return this.props.is_mock; }
    get mockConfig(): string { return this.props.mock_config; }
    get lastResetAt(): Date { return new Date(this.props.last_reset_at); }
    get createdAt(): Date { return new Date(this.props.created_at); }
    get expiresAt(): Date | undefined {
        return this.props.expires_at ? new Date(this.props.expires_at) : undefined;
    }

    get isActive(): boolean {
        if (!this.expiresAt) return true;
        return this.expiresAt > new Date();
    }

    get usagePercentage(): number {
        if (this.props.budget.limit === 0) return 0;
        return (this.props.budget.usage / this.props.budget.limit) * 100;
    }

    toJSON(): KeyProps {
        return { ...this.props };
    }
}
