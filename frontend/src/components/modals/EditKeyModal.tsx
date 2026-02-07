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

export default function EditKeyModal({ editKey, middlewareInfos, providerInfos }: Props) {
    const [id, setId] = useState<number>(0);
    const [name, setName] = useState("");
    const [providerId, setProviderId] = useState("openai");
    const [providerConfig, setProviderConfig] = useState<Record<string, any>>({});
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
            <input type="checkbox" id="edit-key-modal" class="modal-toggle" />
            <div class="modal">
                <div class="modal-box w-11/12 max-w-xl bg-base-100 border border-white/5 rounded-2xl shadow-2xl p-0 overflow-visible">
                    <div class="p-6 border-b border-white/5 bg-base-200/50 rounded-t-2xl flex justify-between items-center">
                        <h3 class="font-bold text-xl text-white tracking-tight">Edit API Key</h3>
                        <label for="edit-key-modal" class="btn btn-sm btn-circle btn-ghost">✕</label>
                    </div>
                    {editKey && (
                        <form class="p-6 flex flex-col gap-6" onSubmit={handleSave}>
                            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Name</span></label>
                                    <input type="text" value={name} onInput={(e) => setName(e.currentTarget.value)} class="input input-bordered w-full bg-white/5 border-white/5 rounded-xl font-bold" required />
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Provider</span></label>
                                    <select value={providerId} onChange={(e) => handleProviderChange(e.currentTarget.value)} class="select select-bordered w-full bg-white/5 border-white/5 rounded-xl text-sm">
                                        {providerInfos.map(p => (
                                            <option key={p.id} value={p.id}>{p.id.charAt(0).toUpperCase() + p.id.slice(1)}</option>
                                        ))}
                                    </select>
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Expires At</span></label>
                                    <input type="datetime-local" value={formatDateForInput(expiresAt)} onChange={handleExpiryChange} class="input input-bordered w-full bg-white/5 border-white/5 rounded-xl text-xs" />
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Budget Limit (USD)</span></label>
                                    <input type="number" step="0.01" value={budgetLimit} onInput={(e) => setBudgetLimit(e.currentTarget.value)} class="input input-bordered w-full bg-white/5 border-white/5 rounded-xl font-bold" />
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Reset Period (Seconds)</span></label>
                                    <input type="number" value={resetPeriod} onInput={(e) => setResetPeriod(e.currentTarget.value)} class="input input-bordered w-full bg-white/5 border-white/5 rounded-xl font-bold" />
                                </div>
                            </div>

                            <ProviderConfigSection
                                providerId={providerId}
                                providerInfos={providerInfos}
                                config={providerConfig}
                                onConfigUpdate={(key, val) => setProviderConfig(prev => ({ ...prev, [key]: val }))}
                            />

                            <MiddlewareComposition
                                middlewares={middlewares}
                                middlewareInfos={middlewareInfos}
                                setMiddlewares={setMiddlewares}
                            />

                            <div class="flex justify-end gap-3 pt-6 border-t border-white/5">
                                <label for="edit-key-modal" class="btn btn-ghost rounded-xl text-white/40 font-bold uppercase tracking-widest text-[10px]">Cancel</label>
                                <button type="submit" class="btn btn-primary px-10 rounded-xl font-bold uppercase tracking-widest text-[11px] h-11 shadow-lg shadow-primary/20" disabled={loading}>
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
