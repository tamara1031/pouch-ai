import { useState, useEffect } from "preact/hooks";
import type { MiddlewareInfo, PluginConfig, ProviderInfo } from "../../types";
import MiddlewareComposition from "./MiddlewareComposition";
import ProviderConfigSection from "./ProviderConfigSection";

interface Props {
    modalRef: any;
    onSuccess: (rawKey: string) => void;
    middlewareInfos: MiddlewareInfo[];
    providerInfos: ProviderInfo[];
}

export default function CreateKeyModal({ modalRef, onSuccess, middlewareInfos, providerInfos }: Props) {
    const [name, setName] = useState("");
    const [providerId, setProviderId] = useState("openai");
    const [providerConfig, setProviderConfig] = useState<Record<string, any>>({});
    const [autoRenew, setAutoRenew] = useState(false);
    const [middlewares, setMiddlewares] = useState<PluginConfig[]>([]);
    const [expiresAtDays, setExpiresAtDays] = useState("0");
    const [budgetLimit, setBudgetLimit] = useState("5.00");
    const [resetPeriod, setResetPeriod] = useState("2592000");
    const [loading, setLoading] = useState(false);

    // Initialize provider config when provider changes
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

    // Default middlewares based on is_default flag in infos
    useEffect(() => {
        if (middlewareInfos.length > 0 && middlewares.length === 0) {
            const initial = middlewareInfos
                .filter(mw => mw.is_default)
                .map(mw => ({
                    id: mw.id,
                    config: Object.keys(mw.schema).reduce((acc, key) => {
                        acc[key] = mw.schema[key].default !== undefined ? mw.schema[key].default : "";
                        return acc;
                    }, {} as Record<string, any>)
                }));
            setMiddlewares(initial);
        }
    }, [middlewareInfos]);

    const handleCreate = async (e: Event) => {
        e.preventDefault();
        setLoading(true);

        let expires_at: number | null = null;
        const days = parseInt(expiresAtDays);
        if (days > 0) {
            expires_at = Math.floor(Date.now() / 1000) + days * 86400;
        }

        try {
            const res = await fetch("/v1/config/app-keys", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    name,
                    provider: { id: providerId, config: providerConfig },
                    middlewares: middlewares,
                    auto_renew: autoRenew,
                    budget_limit: parseFloat(budgetLimit) || 0,
                    reset_period: parseInt(resetPeriod) || 0,
                    expires_at,
                }),
            });

            if (res.ok) {
                const data = await res.json();
                onSuccess(data.key);
                setName("");
                setExpiresAtDays("0");
                setAutoRenew(false);
                setProviderConfig({});
                // Reset middlewares to defaults
                setMiddlewares(middlewareInfos
                    .filter(mw => mw.is_default)
                    .map(mw => ({
                        id: mw.id,
                        config: Object.keys(mw.schema).reduce((acc, key) => {
                            acc[key] = mw.schema[key].default !== undefined ? mw.schema[key].default : "";
                            return acc;
                        }, {} as Record<string, any>)
                    })));
            } else {
                alert("Failed to create key");
            }
        } catch (err) {
            console.error("Create error:", err);
        } finally {
            setLoading(false);
        }
    };

    return (
        <>
            <input type="checkbox" id="create-key-modal" class="modal-toggle" ref={modalRef} />
            <div class="modal modal-bottom sm:modal-middle">
                <div class="modal-box w-full max-w-3xl bg-base-100 border border-white/10 rounded-2xl shadow-2xl p-0 max-h-[90vh] flex flex-col">
                    {/* Header - Fixed */}
                    <div class="p-6 border-b border-white/10 flex justify-between items-center bg-base-200/30 rounded-t-2xl shrink-0">
                        <div>
                            <h3 class="font-bold text-lg text-white">Create API Key</h3>
                            <p class="text-sm text-white/40 mt-0.5">Configure your key settings and middleware.</p>
                        </div>
                        <label for="create-key-modal" class="btn btn-sm btn-circle btn-ghost text-white/40 hover:text-white">✕</label>
                    </div>

                    {/* Scrollable Content */}
                    <form class="flex-1 overflow-y-auto p-6 space-y-6" onSubmit={handleCreate}>
                        {/* Basic Settings */}
                        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                            <div class="form-control">
                                <label class="label pb-1"><span class="label-text text-xs text-white/50 font-medium">Name</span></label>
                                <input type="text" value={name} onInput={(e) => setName(e.currentTarget.value)} placeholder="e.g. My App" class="input input-bordered w-full bg-base-200/50 border-white/10 rounded-lg h-10" required />
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
                                <label class="label pb-1"><span class="label-text text-xs text-white/50 font-medium">Expiration</span></label>
                                <select value={expiresAtDays} onChange={(e) => setExpiresAtDays(e.currentTarget.value)} class="select select-bordered w-full bg-base-200/50 border-white/10 rounded-lg h-10">
                                    <option value="0">No Expiration</option>
                                    <option value="7">1 Week</option>
                                    <option value="30">1 Month</option>
                                    <option value="90">3 Months</option>
                                </select>
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
