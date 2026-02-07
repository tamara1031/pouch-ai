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
    budgetLimit: "0",
    resetPeriod: "0",
};

export default function EditKeyModal() {
    const id = useSignal(0);
    const formData = useSignal(INITIAL_FORM_DATA);
    const loading = useSignal(false);

    const isOpen = modalStore.isEditOpen.value;
    const editKey = modalStore.editKeyData.value;
    const middlewareInfos = pluginStore.middlewares.value;
    const providerInfos = pluginStore.providers.value;

    useEffect(() => {
        if (editKey) {
            id.value = editKey.id;
            formData.value = {
                name: editKey.name,
                providerId: editKey.configuration?.provider.id || "openai",
                providerConfig: editKey.configuration?.provider.config || {},
                autoRenew: editKey.auto_renew || false,
                middlewares: editKey.configuration?.middlewares || [],
                expiresAt: editKey.expires_at,
                budgetLimit: (editKey.configuration?.budget_limit || 0).toString(),
                resetPeriod: (editKey.configuration?.reset_period || 0).toString(),
            };
        }
    }, [editKey]);

    const onClose = () => modalStore.closeEdit();

    const handleSave = async (e: Event) => {
        e.preventDefault();
        loading.value = true;
        try {
            await apiClient.keys.update(id.value, {
                name: formData.value.name,
                auto_renew: formData.value.autoRenew,
                expires_at: formData.value.expiresAt,
                provider: { id: formData.value.providerId, config: formData.value.providerConfig },
                middlewares: formData.value.middlewares,
                budget_limit: parseFloat(formData.value.budgetLimit) || 0,
                reset_period: parseInt(formData.value.resetPeriod) || 0,
            });
            window.dispatchEvent(new CustomEvent('refresh-keys'));
            onClose();
        } catch (err: any) {
            console.error("Update error:", err);
            alert(err.message || "Failed to update key");
        } finally {
            loading.value = false;
        }
    };

    const setFormData = (update: any) => {
        if (typeof update === 'function') {
            formData.value = update(formData.value);
        } else {
            formData.value = update;
        }
    };

    if (!isOpen) return null;

    return (
        <div class="modal modal-open modal-bottom sm:modal-middle backdrop-blur-sm transition-all duration-300">
            <div class="modal-box w-full max-w-3xl bg-base-100 border border-white/10 rounded-2xl shadow-2xl p-0 max-h-[90vh] flex flex-col scale-100 opacity-100 animate-in fade-in zoom-in duration-200">
                <div class="p-6 border-b border-white/10 flex justify-between items-center bg-base-200/30 rounded-t-2xl shrink-0">
                    <div>
                        <h3 class="font-bold text-lg text-white">Edit API Key</h3>
                        <p class="text-sm text-white/40 mt-0.5">Modify key settings and middleware configuration.</p>
                    </div>
                    <button onClick={onClose} class="btn btn-sm btn-circle btn-ghost text-white/40 hover:text-white">✕</button>
                </div>

                {editKey && (
                    <form class="flex-1 overflow-y-auto p-6 space-y-6" onSubmit={handleSave}>
                        <KeyForm
                            formData={formData.value}
                            setFormData={setFormData}
                            middlewareInfos={middlewareInfos}
                            providerInfos={providerInfos}
                            showExpirationField={true}
                        />

                        <div class="flex justify-end gap-3 pt-4 border-t border-white/10">
                            <button type="button" onClick={onClose} class="btn btn-ghost rounded-lg text-white/50 hover:text-white">Cancel</button>
                            <button type="submit" class="btn btn-primary px-8 rounded-lg font-bold" disabled={loading.value}>
                                {loading.value ? <span class="loading loading-spinner loading-xs"></span> : "Save Changes"}
                            </button>
                        </div>
                    </form>
                )}
            </div>
            <div class="modal-backdrop bg-black/40" onClick={onClose}>Close</div>
        </div>
    );
}
