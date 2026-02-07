import { useState, useEffect } from "preact/hooks";
import type { Key, MiddlewareInfo, ProviderInfo } from "../../types";
import { api } from "../../api/api";
import KeyForm from "./KeyForm";

interface Props {
    modalRef: any;
    editKey: Key | null;
    middlewareInfos: MiddlewareInfo[];
    providerInfos: ProviderInfo[];
}

const INITIAL_FORM_DATA = {
    name: "",
    providerId: "openai",
    providerConfig: {},
    autoRenew: false,
    middlewares: [],
    expiresAt: null,
    budgetLimit: "0",
    resetPeriod: "0",
};

export default function EditKeyModal({ modalRef, editKey, middlewareInfos, providerInfos }: Props) {
    const [id, setId] = useState<number>(0);
    const [formData, setFormData] = useState(INITIAL_FORM_DATA);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        if (editKey) {
            setId(editKey.id);
            setFormData({
                name: editKey.name,
                providerId: editKey.configuration?.provider.id || "openai",
                providerConfig: editKey.configuration?.provider.config || {},
                autoRenew: editKey.auto_renew || false,
                middlewares: editKey.configuration?.middlewares || [],
                expiresAt: editKey.expires_at,
                budgetLimit: (editKey.configuration?.budget_limit || 0).toString(),
                resetPeriod: (editKey.configuration?.reset_period || 0).toString(),
            });
        }
    }, [editKey]);

    const handleSave = async (e: Event) => {
        e.preventDefault();
        setLoading(true);
        try {
            await api.keys.update(id, {
                name: formData.name,
                provider: { id: formData.providerId, config: formData.providerConfig },
                middlewares: formData.middlewares,
                auto_renew: formData.autoRenew,
                budget_limit: parseFloat(formData.budgetLimit) || 0,
                reset_period: parseInt(formData.resetPeriod) || 0,
                expires_at: formData.expiresAt,
            });
            window.location.reload();
        } catch (err: any) {
            console.error("Update error:", err);
            alert(err.message || "Failed to update key");
        } finally {
            setLoading(false);
        }
    };

    return (
        <>
            <input type="checkbox" id="edit-key-modal" class="modal-toggle" ref={modalRef} />
            <div class="modal modal-bottom sm:modal-middle">
                <div class="modal-box w-full max-w-3xl bg-base-100 border border-white/10 rounded-2xl shadow-2xl p-0 max-h-[90vh] flex flex-col">
                    <div class="p-6 border-b border-white/10 flex justify-between items-center bg-base-200/30 rounded-t-2xl shrink-0">
                        <div>
                            <h3 class="font-bold text-lg text-white">Edit API Key</h3>
                            <p class="text-sm text-white/40 mt-0.5">Modify key settings and middleware configuration.</p>
                        </div>
                        <label for="edit-key-modal" class="btn btn-sm btn-circle btn-ghost text-white/40 hover:text-white">✕</label>
                    </div>

                    {editKey && (
                        <form class="flex-1 overflow-y-auto p-6 space-y-6" onSubmit={handleSave}>
                            <KeyForm
                                formData={formData}
                                setFormData={setFormData}
                                middlewareInfos={middlewareInfos}
                                providerInfos={providerInfos}
                                showExpirationField={true}
                            />

                            <div class="flex justify-end gap-3 pt-4 border-t border-white/10">
                                <label for="edit-key-modal" class="btn btn-ghost rounded-lg text-white/50 hover:text-white">Cancel</label>
                                <button type="submit" class="btn btn-primary px-8 rounded-lg font-medium" disabled={loading}>
                                    {loading ? "Saving..." : "Save Changes"}
                                </button>
                            </div>
                        </form>
                    )}
                </div>
                <label class="modal-backdrop" for="edit-key-modal">Close</label>
            </div>
        </>
    );
}
