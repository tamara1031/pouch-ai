import { writable, type Writable } from 'svelte/store';
import type { Key } from '../../domain/key/Key';
import type { KeyRepository, CreateKeyDTO } from '../../domain/key/KeyRepository';
import { RestKeyRepository } from '../../infrastructure/repositories/RestKeyRepository';

interface KeyState {
    keys: Key[];
    providers: string[];
    loading: boolean;
    error: string | null;
}

const initialState: KeyState = {
    keys: [],
    providers: [],
    loading: false,
    error: null,
};

export class KeyStore {
    private store: Writable<KeyState>;
    private repository: KeyRepository;

    constructor(repository: KeyRepository = new RestKeyRepository()) {
        this.store = writable(initialState);
        this.repository = repository;
    }

    subscribe(run: (value: KeyState) => void) {
        return this.store.subscribe(run);
    }

    private setLoading(loading: boolean) {
        this.store.update(s => ({ ...s, loading }));
    }

    private setError(error: string | null) {
        this.store.update(s => ({ ...s, error }));
    }

    async loadKeys() {
        this.setLoading(true);
        this.setError(null);
        try {
            const keys = await this.repository.getAll();
            this.store.update(s => ({ ...s, keys }));
        } catch (e: any) {
            this.setError(e.message || 'Failed to load keys');
        } finally {
            this.setLoading(false);
        }
    }

    async loadProviders() {
        try {
            const providers = await this.repository.getProviders();
            this.store.update(s => ({ ...s, providers }));
        } catch (e: any) {
            console.error('Failed to load providers', e);
            // Don't set global error for this, maybe just log or retry?
        }
    }

    async createKey(data: CreateKeyDTO): Promise<string> {
        this.setLoading(true);
        this.setError(null);
        try {
            const result = await this.repository.create(data);
            await this.loadKeys(); // Refresh list
            return result.key;
        } catch (e: any) {
            const msg = e.message || 'Failed to create key';
            this.setError(msg);
            throw e;
        } finally {
            this.setLoading(false);
        }
    }

    async deleteKey(id: number) {
        // Optimistic update? Or just reload?
        // Let's do optimistic for better UX
        this.store.update(s => ({
            ...s,
            keys: s.keys.filter(k => k.id !== id)
        }));

        try {
            await this.repository.delete(id);
        } catch (e: any) {
            this.setError(e.message || 'Failed to delete key');
            // Revert on failure
            await this.loadKeys();
        }
    }
}

export const keyStore = new KeyStore();
