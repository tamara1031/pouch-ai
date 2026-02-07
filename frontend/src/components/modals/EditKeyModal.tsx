import { useState, useEffect } from "preact/hooks";
import type { Key, MiddlewareInfo, PluginConfig, ProviderInfo } from "../../types";
import MiddlewareComposition from "./MiddlewareComposition";
import ProviderConfigSection from "./ProviderConfigSection";

interface Props {
    modalRef: any;
    editKey: Key | null;
    middlewareInfos: MiddlewareInfo[];
    providerInfos: ProviderInfo[];
}

export default function EditKeyModal({ modalRef, editKey, middlewareInfos, providerInfos }: Props) {
    const [id, setId] = useState<number>(0);
    const [name, setName] = useState("");
    const [providerId, setProviderId] = useState("openai");
    const [providerConfig, setProviderConfig] = useState<Record<string, any>>({});
    const [autoRenew, setAutoRenew] = useState(false);
    const [middlewares, setMiddlewares] = useState<PluginConfig[]>([]);
    const [expiresAt, setExpiresAt] = useState<number | null>(null);
    const [budgetLimit, setBudgetLimit] = useState("0");
    const [resetPeriod, setResetPeriod] = useState("0");
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        if (editKey) {
            setId(editKey.id);
            setName(editKey.name);
            setProviderId(editKey.configuration?.provider.id || "openai");
            setProviderConfig(editKey.configuration?.provider.config || {});
            setAutoRenew(editKey.auto_renew || false);
            setMiddlewares(editKey.configuration?.middlewares || []);
            setExpiresAt(editKey.expires_at);
            setBudgetLimit((editKey.configuration?.budget_limit || 0).toString());
            setResetPeriod((editKey.configuration?.reset_period || 0).toString());
        }
    }, [editKey]);

    // Reset provider config when provider changes (but not when loading from editKey)
    const handleProviderChange = (newProviderId: string) => {
        setProviderId(newProviderId);
        const info = providerInfos.find(p => p.id === newProviderId);
        if (info?.schema) {
            const defaults = Object.keys(info.schema).reduce((acc, key) => {
                acc[key] = info.schema[key].default ?? "";
                return acc;
            }, {} as Record<string, any>);
            setProviderConfig(defaults);
        } else {
            setProviderConfig({});
        }
    };

    const handleSave = async (e: Event) => {
        e.preventDefault();
        setLoading(true);
        try {
            const res = await fetch(`/v1/config/app-keys/${id}`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    name,
                    provider: { id: providerId, config: providerConfig },
                    middlewares: middlewares,
                    auto_renew: autoRenew,
                    budget_limit: parseFloat(budgetLimit) || 0,
                    reset_period: parseInt(resetPeriod) || 0,
                    expires_at: expiresAt,
                }),
            });
            if (res.ok) {
                window.location.reload();
            } else {
                alert("Failed to update key");
            }
        } catch (err) {
            console.error("Update error:", err);
        } finally {
            setLoading(false);
        }
    };

    const formatDateForInput = (timestamp: number | null) => {
        if (!timestamp) return "";
        const date = new Date(timestamp * 1000);
        const year = date.getFullYear();
        const month = (date.getMonth() + 1).toString().padStart(2, '0');
        const day = date.getDate().toString().padStart(2, '0');
        const hours = date.getHours().toString().padStart(2, '0');
        const minutes = date.getMinutes().toString().padStart(2, '0');
        return `${year}-${month}-${day}T${hours}:${minutes}`;
    };

    const handleExpiryChange = (e: Event) => {
        const value = (e.target as HTMLInputElement).value;
        if (value === "") {
            setExpiresAt(null);
        } else {
            setExpiresAt(Math.floor(new Date(value).getTime() / 1000));
        }
    };

    return (
        <>
            <input type="checkbox" id="edit-key-modal" class="modal-toggle" ref={modalRef} />
            <div class="modal modal-bottom sm:modal-middle">
                <div class="modal-box w-full max-w-3xl bg-base-100 border border-white/10 rounded-2xl shadow-2xl p-0 max-h-[90vh] flex flex-col">
                    {/* Header - Fixed */}
                    <div class="p-6 border-b border-white/10 flex justify-between items-center bg-base-200/30 rounded-t-2xl shrink-0">
                        <div>
                            <h3 class="font-bold text-lg text-white">Edit API Key</h3>
                            <p class="text-sm text-white/40 mt-0.5">Modify key settings and middleware configuration.</p>
                        </div>
                        <label for="edit-key-modal" class="btn btn-sm btn-circle btn-ghost text-white/40 hover:text-white">✕</label>
                    </div>

                    {editKey && (
                        <form class="flex-1 overflow-y-auto p-6 space-y-6" onSubmit={handleSave}>
                            {/* Basic Settings */}
                            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                                <div class="form-control">
                                    <label class="label pb-1"><span class="label-text text-xs text-white/50 font-medium">Name</span></label>
                                    <input type="text" value={name} onInput={(e) => setName(e.currentTarget.value)} class="input input-bordered w-full bg-base-200/50 border-white/10 rounded-lg h-10" required />
                                </div>
                                <div class="form-control">
                                    <label class="label pb-1"><span class="label-text text-xs text-white/50 font-medium">Provider</span></label>
                                    <select value={providerId} onChange={(e) => handleProviderChange(e.currentTarget.value)} class="select select-bordered w-full bg-base-200/50 border-white/10 rounded-lg h-10">
                                        {providerInfos.map(p => (
                                            <option key={p.id} value={p.id}>{p.id.charAt(0).toUpperCase() + p.id.slice(1)}</option>
                                        ))}
                                    </select>
                                </div>
                                <div class="form-control">
                                    <label class="label pb-1"><span class="label-text text-xs text-white/50 font-medium">Expires At</span></label>
                                    <input type="datetime-local" value={formatDateForInput(expiresAt)} onChange={handleExpiryChange} class="input input-bordered w-full bg-base-200/50 border-white/10 rounded-lg h-10 text-sm" />
                                </div>
                                <div class="form-control">
                                    <label class="label pb-1"><span class="label-text text-xs text-white/50 font-medium">Budget Limit (USD)</span></label>
                                    <input type="number" step="0.01" value={budgetLimit} onInput={(e) => setBudgetLimit(e.currentTarget.value)} class="input input-bordered w-full bg-base-200/50 border-white/10 rounded-lg h-10" />
                                </div>
                                <div class="form-control">
                                    <label class="label pb-1 cursor-pointer flex justify-start gap-3">
                                        <input type="checkbox" checked={autoRenew} onChange={(e) => setAutoRenew(e.currentTarget.checked)} class="checkbox checkbox-primary checkbox-sm rounded-md" />
                                        <span class="label-text text-sm font-medium text-white/70">Auto-Renew</span>
                                    </label>
                                    <div class="text-[10px] text-white/30 pl-8">Automatically reset budget and extend expiration</div>
                                </div>
                                <div class="form-control sm:col-span-2">
                                    <label class="label pb-1"><span class="label-text text-xs text-white/50 font-medium">Reset Period (Seconds)</span></label>
                                    <input type="number" value={resetPeriod} onInput={(e) => setResetPeriod(e.currentTarget.value)} placeholder="2592000 = 30 days" class="input input-bordered w-full bg-base-200/50 border-white/10 rounded-lg h-10" />
                                </div>
                            </div>

                            {/* Provider Config */}
                            <ProviderConfigSection
                                providerId={providerId}
                                providerInfos={providerInfos}
                                config={providerConfig}
                                onConfigUpdate={(key, val) => setProviderConfig(prev => ({ ...prev, [key]: val }))}
                            />

                            {/* Middlewares */}
                            <MiddlewareComposition
                                middlewares={middlewares}
                                middlewareInfos={middlewareInfos}
                                setMiddlewares={setMiddlewares}
                            />

                            {/* Footer - Fixed */}
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
