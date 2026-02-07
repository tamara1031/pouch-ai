import { useState, useRef, useEffect } from "preact/hooks";
import type { MiddlewareInfo } from "../../types";

interface Props {
    modalRef: any;
    onSuccess: (rawKey: string) => void;
}

export default function CreateKeyModal({ modalRef, onSuccess }: Props) {
    const [createMode, setCreateMode] = useState<"prepaid" | "subscription">("prepaid");
    const [createProvider, setCreateProvider] = useState("openai");
    const [availableMiddlewares, setAvailableMiddlewares] = useState<MiddlewareInfo[]>([]);
    const [enabledMiddlewares, setEnabledMiddlewares] = useState<{ id: string, config: Record<string, string> }[]>([]);

    useEffect(() => {
        fetch("/v1/config/plugins/middlewares")
            .then(res => res.json())
            .then(data => {
                const mws: MiddlewareInfo[] = data.middlewares || [];
                setAvailableMiddlewares(mws);
                // Default enabled middlewares
                setEnabledMiddlewares(mws.map(mw => ({
                    id: mw.id,
                    config: Object.keys(mw.schema).reduce((acc, key) => {
                        acc[key] = mw.schema[key].default || "";
                        return acc;
                    }, {} as Record<string, string>)
                })));
            })
            .catch(err => console.error("Failed to fetch middlewares:", err));
    }, []);

    const toggleMiddleware = (mwId: string) => {
        setEnabledMiddlewares(prev => {
            const exists = prev.find(m => m.id === mwId);
            if (exists) return prev.filter(m => m.id !== mwId);
            const mw = availableMiddlewares.find(m => m.id === mwId);
            if (!mw) return prev;
            return [...prev, {
                id: mwId,
                config: Object.keys(mw.schema).reduce((acc, key) => {
                    acc[key] = mw.schema[key].default || "";
                    return acc;
                }, {} as Record<string, string>)
            }];
        });
    };

    const moveMiddleware = (index: number, direction: 'up' | 'down') => {
        setEnabledMiddlewares(prev => {
            const next = [...prev];
            const newIndex = direction === 'up' ? index - 1 : index + 1;
            if (newIndex < 0 || newIndex >= next.length) return prev;
            [next[index], next[newIndex]] = [next[newIndex], next[index]];
            return next;
        });
    };

    const updateMiddlewareConfig = (mwId: string, key: string, value: string) => {
        setEnabledMiddlewares(prev => prev.map(m =>
            m.id === mwId ? { ...m, config: { ...m.config, [key]: value } } : m
        ));
    };

    const handleCreateSubmit = async (e: Event) => {
        e.preventDefault();
        const form = e.target as HTMLFormElement;
        const fd = new FormData(form);

        const mode = fd.get("mode_type") as string;
        const days = parseInt(fd.get("expiration") as string);
        let expires_at: number | null = null;
        let period = "none";

        if (mode === "prepaid") {
            if (days > 0) expires_at = Math.floor(Date.now() / 1000) + days * 86400;
        } else {
            period = fd.get("budget_period") as string;
        }

        const provider = (fd.get("provider") as string) || "openai";

        const payload = {
            name: fd.get("name"),
            provider: provider,
            expires_at,
            budget_limit: parseFloat(fd.get("budget_limit") as string),
            budget_period: period,
            rate_limit: parseInt(fd.get("rate_limit") as string) || 10,
            rate_period: fd.get("rate_period") || "minute",
            middlewares: enabledMiddlewares,
            mock_config: fd.get("mock_config"),
        };

        try {
            const res = await fetch("/v1/config/app-keys", {
                method: "POST",
                body: JSON.stringify(payload),
                headers: { "Content-Type": "application/json" },
            });

            if (res.ok) {
                const data = await res.json();
                onSuccess(data.key);
                if (modalRef.current) modalRef.current.checked = false;
                form.reset();
                setEnabledMiddlewares(availableMiddlewares.map(mw => ({
                    id: mw.id,
                    config: Object.keys(mw.schema).reduce((acc, key) => {
                        acc[key] = mw.schema[key].default || "";
                        return acc;
                    }, {} as Record<string, string>)
                })));
                setCreateProvider("openai");
                setCreateMode("prepaid");
            } else {
                alert("Failed to create key");
            }
        } catch (err) {
            console.error("Create error:", err);
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
                            <p class="text-xs text-white/40 mt-1">Set usage limits and provider for your new key.</p>
                        </div>
                        <label for="create-key-modal" class="btn btn-sm btn-circle btn-ghost">✕</label>
                    </div>

                    <form id="create-key-form" class="p-6 flex flex-col gap-6" onSubmit={handleCreateSubmit}>
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div class="form-control">
                                <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Name</span></label>
                                <input type="text" name="name" placeholder="e.g. My App" class="input input-bordered w-full bg-white/5 border-white/5 focus:border-primary/50 rounded-xl" required />
                            </div>
                            <div class="form-control">
                                <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Provider</span></label>
                                <select name="provider" class="select select-bordered w-full bg-white/5 border-white/5 rounded-xl text-sm" onChange={(e) => setCreateProvider(e.currentTarget.value)}>
                                    <option value="openai" selected>OpenAI</option>
                                    <option value="mock">Mock</option>
                                    <option value="anthropic" disabled>Anthropic (Soon)</option>
                                </select>
                            </div>
                            <div class="form-control">
                                <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Budget Type</span></label>
                                <div class="flex p-1 bg-white/5 rounded-xl border border-white/5">
                                    <button type="button" onClick={() => setCreateMode("prepaid")} class={`flex-1 py-1.5 text-[10px] font-bold tracking-wider rounded-lg transition-all ${createMode === "prepaid" ? 'bg-primary text-white' : 'text-white/40'}`}>ONE-TIME</button>
                                    <button type="button" onClick={() => setCreateMode("subscription")} class={`flex-1 py-1.5 text-[10px] font-bold tracking-wider rounded-lg transition-all ${createMode === "subscription" ? 'bg-primary text-white' : 'text-white/40'}`}>RECURRING</button>
                                    <input type="hidden" name="mode_type" value={createMode} />
                                </div>
                            </div>
                            <div class="form-control">
                                <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Limit ($)</span></label>
                                <input type="number" name="budget_limit" defaultValue="5.00" step="0.01" min="0" class="input input-bordered w-full bg-white/5 border-white/5 rounded-xl font-mono" />
                            </div>
                            {createMode === "prepaid" ? (
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Expiration</span></label>
                                    <select name="expiration" class="select select-bordered w-full bg-white/5 border-white/5 rounded-xl text-sm">
                                        <option value="7">1 Week</option>
                                        <option value="30">1 Month</option>
                                        <option value="90" selected>3 Months</option>
                                        <option value="0">Indefinite</option>
                                    </select>
                                </div>
                            ) : (
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Reset Period</span></label>
                                    <select name="budget_period" class="select select-bordered w-full bg-white/5 border-white/5 rounded-xl text-sm">
                                        <option value="monthly" selected>Monthly</option>
                                        <option value="weekly">Weekly</option>
                                    </select>
                                </div>
                            )}
                            <div class="form-control">
                                <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Rate Limit</span></label>
                                <div class="flex flex-row gap-2">
                                    <input type="number" name="rate_limit" defaultValue="10" min="0" class="input input-bordered flex-1 min-w-0 bg-white/5 border-white/5 rounded-xl font-mono" />
                                    <select name="rate_period" class="select select-bordered bg-white/5 border-white/5 rounded-xl text-[10px] font-bold uppercase w-20">
                                        <option value="minute" selected>RPM</option>
                                        <option value="second">RPS</option>
                                        <option value="none">UNLT</option>
                                    </select>
                                </div>
                            </div>
                        </div>

                        <div class="border-t border-white/5 pt-6 space-y-4">
                            <div class="flex justify-between items-center">
                                <label class="label p-0"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Middleware Composition</span></label>
                                <div class="dropdown dropdown-end">
                                    <div tabindex={0} role="button" class="btn btn-xs btn-primary rounded-lg font-bold">Add Middleware</div>
                                    <ul tabindex={0} class="dropdown-content z-[20] menu p-2 shadow-2xl bg-base-300 border border-white/10 rounded-xl w-52 mt-2">
                                        {availableMiddlewares.filter(mw => !enabledMiddlewares.some(em => em.id === mw.id)).map(mw => (
                                            <li key={mw.id}><a onClick={() => toggleMiddleware(mw.id)} class="text-xs font-bold text-white/60 hover:text-white">{mw.id.replace(/_/g, " ")}</a></li>
                                        ))}
                                        {availableMiddlewares.filter(mw => !enabledMiddlewares.some(em => em.id === mw.id)).length === 0 && (
                                            <li class="disabled"><span class="text-xs italic text-white/20">All plugins added</span></li>
                                        )}
                                    </ul>
                                </div>
                            </div>

                            <div class="grid grid-cols-1 gap-3">
                                {enabledMiddlewares.map((emw, idx) => {
                                    const mwInfo = availableMiddlewares.find(m => m.id === emw.id);
                                    if (!mwInfo) return null;
                                    return (
                                        <div class="p-4 rounded-xl bg-white/5 border border-white/5 space-y-4 relative group" key={emw.id}>
                                            <div class="flex items-center justify-between">
                                                <div class="flex items-center gap-3">
                                                    <div class="flex flex-col gap-1">
                                                        <button type="button" onClick={() => moveMiddleware(idx, 'up')} class={`btn btn-ghost btn-xs p-0 min-h-0 h-4 w-4 ${idx === 0 ? 'invisible' : ''}`}>▲</button>
                                                        <button type="button" onClick={() => moveMiddleware(idx, 'down')} class={`btn btn-ghost btn-xs p-0 min-h-0 h-4 w-4 ${idx === enabledMiddlewares.length - 1 ? 'invisible' : ''}`}>▼</button>
                                                    </div>
                                                    <span class="text-xs font-bold text-white/80">{emw.id.replace(/_/g, " ")}</span>
                                                </div>
                                                <button type="button" onClick={() => toggleMiddleware(emw.id)} class="btn btn-ghost btn-xs h-6 w-6 btn-circle text-white/20 group-hover:text-red-400">✕</button>
                                            </div>

                                            {Object.keys(mwInfo.schema).length > 0 && (
                                                <div class="grid grid-cols-1 md:grid-cols-2 gap-3 pl-6 border-l border-white/10">
                                                    {Object.keys(mwInfo.schema).map(key => {
                                                        const schema = mwInfo.schema[key];
                                                        return (
                                                            <div class="form-control" key={key}>
                                                                <label class="label pt-0"><span class="label-text text-[10px] text-white/40 uppercase font-semibold">{schema.displayName || key.replace(/_/g, " ")}</span></label>
                                                                {schema.type === "select" ? (
                                                                    <select
                                                                        class="select select-bordered select-xs bg-white/5 border-white/5 rounded-lg text-[10px]"
                                                                        value={emw.config[key]}
                                                                        onChange={(e) => updateMiddlewareConfig(emw.id, key, e.currentTarget.value)}
                                                                    >
                                                                        {schema.options?.map(opt => <option value={opt} key={opt}>{opt}</option>)}
                                                                    </select>
                                                                ) : (
                                                                    <input
                                                                        type={schema.type === "number" ? "number" : "text"}
                                                                        placeholder={schema.description}
                                                                        value={emw.config[key]}
                                                                        onInput={(e) => updateMiddlewareConfig(emw.id, key, e.currentTarget.value)}
                                                                        class="input input-bordered input-xs bg-white/5 border-white/5 rounded-lg font-mono text-[10px]"
                                                                    />
                                                                )}
                                                            </div>
                                                        );
                                                    })}
                                                </div>
                                            )}
                                        </div>
                                    )
                                })}
                                {enabledMiddlewares.length === 0 && (
                                    <div class="flex flex-col items-center justify-center p-8 rounded-2xl border-2 border-dashed border-white/5 bg-white/2 cursor-pointer hover:bg-white/5 transition-all">
                                        <p class="text-[10px] text-white/20 font-bold uppercase tracking-widest">No Middlewares Enabled</p>
                                        <p class="text-[10px] text-white/10 mt-1">Add one from the dropdown above</p>
                                    </div>
                                )}
                            </div>
                        </div>

                        {createProvider === "mock" && (
                            <div class="form-control p-4 rounded-xl bg-white/5 border border-white/5 mt-2">
                                <label class="label mb-2">
                                    <span class="font-bold text-sm text-white/80">Mock Configuration</span>
                                </label>
                                <textarea name="mock_config" class="textarea textarea-bordered h-24 font-mono text-[10px] w-full bg-black/20 border-white/5 rounded-xl" defaultValue={`{"choices":[{"message":{"content":"Hello from Mock!", "role":"assistant"}}]}`} />
                            </div>
                        )}

                        <div class="flex justify-end gap-3 pt-6 border-t border-white/5">
                            <label for="create-key-modal" class="btn btn-ghost rounded-xl text-white/40 font-bold uppercase tracking-widest text-[10px]">Cancel</label>
                            <button type="submit" class="btn btn-primary px-10 rounded-xl font-bold uppercase tracking-widest text-[11px] h-11 shadow-lg shadow-primary/20">Generate Key</button>
                        </div>
                    </form>
                </div>
            </div>
        </>
    );
}
