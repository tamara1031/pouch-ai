import { signal } from "@preact/signals";
import type { MiddlewareInfo, ProviderInfo, Key } from "../types";

export const modalStore = {
    isCreateOpen: signal(false),
    isEditOpen: signal(false),
    isNewKeyOpen: signal(false),

    // Data associated with modals
    editKeyData: signal<Key | null>(null),
    newKeyRaw: signal<string | null>(null),

    openCreate: () => {
        modalStore.isCreateOpen.value = true;
    },
    closeCreate: () => {
        modalStore.isCreateOpen.value = false;
    },

    openEdit: (key: Key) => {
        modalStore.editKeyData.value = key;
        modalStore.isEditOpen.value = true;
    },
    closeEdit: () => {
        modalStore.isEditOpen.value = false;
        modalStore.editKeyData.value = null;
    },

    openNewKey: (rawKey: string) => {
        modalStore.newKeyRaw.value = rawKey;
        modalStore.isNewKeyOpen.value = true;
    },
    closeNewKey: () => {
        modalStore.isNewKeyOpen.value = false;
        modalStore.newKeyRaw.value = null;
    }
};

export const pluginStore = {
    middlewares: signal<MiddlewareInfo[]>([]),
    providers: signal<ProviderInfo[]>([]),
    loading: signal(true),
    error: signal<Error | null>(null),
};

export const keyStore = {
    keys: signal<Key[]>([]),
    loading: signal(false),
    error: signal<Error | null>(null),
};
