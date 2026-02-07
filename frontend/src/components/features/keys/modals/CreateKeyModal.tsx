import { useEffect } from "preact/hooks";
import { useSignal } from "@preact/signals";
import { modalStore, pluginStore } from "../../../../lib/store";
import { apiClient } from "../../../../lib/api-client";
import KeyForm from "../form/KeyForm";
import type { PluginConfig } from "../../../../types";

const INITIAL_FORM_DATA = {
    name: "",
    providerId: "openai",
    providerConfig: {} as Record<string, any>,
    autoRenew: false,
    middlewares: [] as PluginConfig[],
    expiresAt: null as number | null,
    budgetLimit: "5.00",
    resetPeriod: "2592000",
};

export default function CreateKeyModal() {
    const formData = useSignal(INITIAL_FORM_DATA);
    const expirationDays = useSignal("0");
    const loading = useSignal(false);

    const isOpen = modalStore.isCreateOpen.value;
    const middlewareInfos = pluginStore.middlewares.value;
    const providerInfos = pluginStore.providers.value;

    useEffect(() => {
        if (isOpen && middlewareInfos.length > 0 && formData.value.middlewares.length === 0) {
            const initial = middlewareInfos
                .filter(mw => mw.is_default)
                .map(mw => ({
                    id: mw.id,
                    config: Object.keys(mw.schema).reduce((acc, key) => {
                        acc[key] = mw.schema[key].default !== undefined ? mw.schema[key].default : "";
                        return acc;
                    }, {} as Record<string, any>)
                }));
            formData.value = { ...formData.value, middlewares: initial };
        }
    }, [isOpen, middlewareInfos]);

    const onClose = () => modalStore.closeCreate();

    const handleCreate = async (e: Event) => {
        e.preventDefault();
        loading.value = true;

        let expires_at: number | null = null;
        const days = parseInt(expirationDays.value);
        if (days > 0) {
            expires_at = Math.floor(Date.now() / 1000) + days * 86400;
        }

        try {
            const data = await apiClient.keys.create({
                name: formData.value.name,
                auto_renew: formData.value.autoRenew,
                expires_at,
                provider: { id: formData.value.providerId, config: formData.value.providerConfig },
                middlewares: formData.value.middlewares,
                budget_limit: parseFloat(formData.value.budgetLimit) || 0,
                reset_period: parseInt(formData.value.resetPeriod) || 0,
            });

            onClose();
            modalStore.openNewKey(data.key);
            window.dispatchEvent(new CustomEvent('refresh-keys'));

            // Reset form
            formData.value = {
                ...INITIAL_FORM_DATA,
                middlewares: middlewareInfos
                    .filter(mw => mw.is_default)
                    .map(mw => ({
                        id: mw.id,
                        config: Object.keys(mw.schema || {}).reduce((acc, key) => {
                            acc[key] = mw.schema[key].default !== undefined ? mw.schema[key].default : "";
                            return acc;
                        }, {} as Record<string, any>)
                    }))
            };
            expirationDays.value = "0";
        } catch (err: any) {
            console.error("Create error:", err);
            alert(err.message || "Failed to create key");
        } finally {
            loading.value = false;
        }
    };

    if (!isOpen) return null;

    const setFormData = (update: any) => {
        if (typeof update === 'function') {
            formData.value = update(formData.value);
        } else {
            formData.value = update;
        }
    };

    return (
        <div class="modal modal-open modal-bottom sm:modal-middle backdrop-blur-sm transition-all duration-300">
            <div class="modal-box w-full max-w-3xl bg-base-100 border border-white/10 rounded-2xl shadow-2xl p-0 max-h-[90vh] flex flex-col scale-100 opacity-100 animate-in fade-in zoom-in duration-200">
                <div class="p-6 border-b border-white/10 flex justify-between items-center bg-base-200/30 rounded-t-2xl shrink-0">
                    <div>
                        <h3 class="font-bold text-lg text-white">Create API Key</h3>
                        <p class="text-sm text-white/40 mt-0.5">Configure your key settings and middleware.</p>
                    </div>
                    <button onClick={onClose} class="btn btn-sm btn-circle btn-ghost text-white/40 hover:text-white">✕</button>
                </div>

                <form class="flex-1 overflow-y-auto p-6 space-y-6" onSubmit={handleCreate}>
                    <KeyForm
                        formData={formData.value}
                        setFormData={setFormData}
                        middlewareInfos={middlewareInfos}
                        providerInfos={providerInfos}
                        expirationDays={expirationDays.value}
                        setExpirationDays={(val) => expirationDays.value = val}
                    />

                    <div class="flex justify-end gap-3 pt-4 border-t border-white/10">
                        <button type="button" onClick={onClose} class="btn btn-ghost rounded-lg text-white/50 hover:text-white">Cancel</button>
                        <button type="submit" class="btn btn-primary px-8 rounded-lg font-bold" disabled={loading.value}>
                            {loading.value ? <span class="loading loading-spinner loading-xs"></span> : "Create Key"}
                        </button>
                    </div>
                </form>
            </div>
            <div class="modal-backdrop bg-black/40" onClick={onClose}>Close</div>
        </div>
    );
}
