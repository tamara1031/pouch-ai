import { useState, useEffect } from "preact/hooks";
import type { MiddlewareInfo, ProviderInfo } from "../../types";
import { api } from "../../api/api";
import KeyForm from "./KeyForm";

interface Props {
    modalRef: any;
    onSuccess: (rawKey: string) => void;
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
    budgetLimit: "5.00",
    resetPeriod: "2592000",
};

export default function CreateKeyModal({ modalRef, onSuccess, middlewareInfos, providerInfos }: Props) {
    const [formData, setFormData] = useState(INITIAL_FORM_DATA);
    const [expirationDays, setExpirationDays] = useState("0");
    const [loading, setLoading] = useState(false);

    // Initial default middlewares
    useEffect(() => {
        if (middlewareInfos.length > 0 && formData.middlewares.length === 0) {
            const initial = middlewareInfos
                .filter(mw => mw.is_default)
                .map(mw => ({
                    id: mw.id,
                    config: Object.keys(mw.schema).reduce((acc, key) => {
                        acc[key] = mw.schema[key].default !== undefined ? mw.schema[key].default : "";
                        return acc;
                    }, {} as Record<string, any>)
                }));
            setFormData(prev => ({ ...prev, middlewares: initial }));
        }
    }, [middlewareInfos]);

    const handleCreate = async (e: Event) => {
        e.preventDefault();
        setLoading(true);

        let expires_at: number | null = null;
        const days = parseInt(expirationDays);
        if (days > 0) {
            expires_at = Math.floor(Date.now() / 1000) + days * 86400;
        }

        try {
            const data = await api.keys.create({
                name: formData.name,
                provider: { id: formData.providerId, config: formData.providerConfig },
                middlewares: formData.middlewares,
                auto_renew: formData.autoRenew,
                budget_limit: parseFloat(formData.budgetLimit) || 0,
                reset_period: parseInt(formData.resetPeriod) || 0,
                expires_at,
            });

            onSuccess(data.key);
            setFormData({
                ...INITIAL_FORM_DATA,
                middlewares: middlewareInfos
                    .filter(mw => mw.is_default)
                    .map(mw => ({
                        id: mw.id,
                        config: Object.keys(mw.schema).reduce((acc, key) => {
                            acc[key] = mw.schema[key].default !== undefined ? mw.schema[key].default : "";
                            return acc;
                        }, {} as Record<string, any>)
                    }))
            });
            setExpirationDays("0");
        } catch (err: any) {
            console.error("Create error:", err);
            alert(err.message || "Failed to create key");
        } finally {
            setLoading(false);
        }
    };

    return (
        <>
            <input type="checkbox" id="create-key-modal" class="modal-toggle" ref={modalRef} />
            <div class="modal modal-bottom sm:modal-middle">
                <div class="modal-box w-full max-w-3xl bg-base-100 border border-white/10 rounded-2xl shadow-2xl p-0 max-h-[90vh] flex flex-col">
                    <div class="p-6 border-b border-white/10 flex justify-between items-center bg-base-200/30 rounded-t-2xl shrink-0">
                        <div>
                            <h3 class="font-bold text-lg text-white">Create API Key</h3>
                            <p class="text-sm text-white/40 mt-0.5">Configure your key settings and middleware.</p>
                        </div>
                        <label for="create-key-modal" class="btn btn-sm btn-circle btn-ghost text-white/40 hover:text-white">✕</label>
                    </div>

                    <form class="flex-1 overflow-y-auto p-6 space-y-6" onSubmit={handleCreate}>
                        <KeyForm
                            formData={formData}
                            setFormData={setFormData}
                            middlewareInfos={middlewareInfos}
                            providerInfos={providerInfos}
                            expirationDays={expirationDays}
                            setExpirationDays={setExpirationDays}
                        />

                        <div class="flex justify-end gap-3 pt-4 border-t border-white/10">
                            <label for="create-key-modal" class="btn btn-ghost rounded-lg text-white/50 hover:text-white">Cancel</label>
                            <button type="submit" class="btn btn-primary px-8 rounded-lg font-medium" disabled={loading}>
                                {loading ? "Creating..." : "Create Key"}
                            </button>
                        </div>
                    </form>
                </div>
                <label class="modal-backdrop" for="create-key-modal">Close</label>
            </div>
        </>
    );
}
