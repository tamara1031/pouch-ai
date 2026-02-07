import { useState } from "preact/hooks";
import { signal } from "@preact/signals";
import type { Key } from "../types";

export const editKeySignal = signal<Key | null>(null);
export const isEditModalOpen = signal(false);

export const isCreateModalOpen = signal(false);

export const newKeyRawSignal = signal<string | null>(null);
export const isNewKeyModalOpen = signal(false);

export function openCreateModal() {
    isCreateModalOpen.value = true;
}

export function openEditModal(key: Key) {
    editKeySignal.value = key;
    isEditModalOpen.value = true;
}

export function openNewKeyModal(rawKey: string) {
    newKeyRawSignal.value = rawKey;
    isNewKeyModalOpen.value = true;
}
