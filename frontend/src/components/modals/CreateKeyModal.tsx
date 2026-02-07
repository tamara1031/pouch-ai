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
    const [middlewares, setMiddlewares] = useState<PluginConfig[]>([]);
    const [expiresAtDays, setExpiresAtDays] = useState("0");
    const [budgetLimit, setBudgetLimit] = useState("5.00");
    const [resetPeriod, setResetPeriod] = useState("2592000"); // 30 days default
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
            <div class="modal">
                <div class="modal-box w-11/12 max-w-2xl bg-base-100 border border-white/5 rounded-2xl shadow-2xl p-0 overflow-visible">
                    <div class="p-6 border-b border-white/5 flex justify-between items-center bg-base-200/50 rounded-t-2xl">
                        <div>
                            <h3 class="font-bold text-xl text-white tracking-tight">Generate API Key</h3>
                            <p class="text-xs text-white/40 mt-1">Configure your key and plugins.</p>
                        </div>
                        <label for="create-key-modal" class="btn btn-sm btn-circle btn-ghost">✕</label>
                    </div>

                    <form class="p-6 flex flex-col gap-6" onSubmit={handleCreate}>
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div class="form-control">
                                <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Name</span></label>
                                <input type="text" value={name} onInput={(e) => setName(e.currentTarget.value)} placeholder="e.g. My App" class="input input-bordered w-full bg-white/5 border-white/5 rounded-xl font-bold" required />
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
                                <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Expiration</span></label>
                                <select value={expiresAtDays} onChange={(e) => setExpiresAtDays(e.currentTarget.value)} class="select select-bordered w-full bg-white/5 border-white/5 rounded-xl text-sm">
                                    <option value="0">Indefinite</option>
                                    <option value="7">1 Week</option>
                                    <option value="30">1 Month</option>
                                    <option value="90">3 Months</option>
                                </select>
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
                            <label for="create-key-modal" class="btn btn-ghost rounded-xl text-white/40 font-bold uppercase tracking-widest text-[10px]">Cancel</label>
                            <button type="submit" class="btn btn-primary px-10 rounded-xl font-bold uppercase tracking-widest text-[11px] h-11 shadow-lg shadow-primary/20" disabled={loading}>
                                {loading ? "Generating..." : "Generate Key"}
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </>
    );
}
