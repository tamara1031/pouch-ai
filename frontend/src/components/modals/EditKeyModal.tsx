import { useState, useEffect } from "preact/hooks";
import type { Key, MiddlewareInfo, PluginConfig } from "../../types";

interface Props {
    modalRef: any;
    editKey: Key | null;
    middlewareInfos: MiddlewareInfo[];
}

export default function EditKeyModal({ editKey, middlewareInfos }: Props) {
    const [id, setId] = useState<number>(0);
    const [name, setName] = useState("");
    const [providerId, setProviderId] = useState("openai");
    const [providerConfig, setProviderConfig] = useState<Record<string, any>>({});
    const [middlewares, setMiddlewares] = useState<PluginConfig[]>([]);
    const [expiresAt, setExpiresAt] = useState<number | null>(null);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        if (editKey) {
            setId(editKey.id);
            setName(editKey.name);
            setProviderId(editKey.configuration?.provider.id || "openai");
            setProviderConfig(editKey.configuration?.provider.config || {});
            setMiddlewares(editKey.configuration?.middlewares || []);
            setExpiresAt(editKey.expires_at);
        }
    }, [editKey]);

    const handleSave = async (e: Event) => {
        e.preventDefault();
        setLoading(true);
        try {
            const res = await fetch(`/v1/config/app-keys/${id}`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    name,
                    configuration: {
                        provider: { id: providerId, config: providerConfig },
                        middlewares: middlewares,
                    },
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

    const toggleMiddleware = (mwId: string) => {
        setMiddlewares(prev => {
            const exists = prev.find(m => m.id === mwId);
            if (exists) return prev.filter(m => m.id !== mwId);
            const mw = middlewareInfos.find(m => m.id === mwId);
            if (!mw) return prev;
            return [...prev, {
                id: mwId,
                config: Object.keys(mw.schema).reduce((acc, key) => {
                    acc[key] = mw.schema[key].default !== undefined ? mw.schema[key].default : "";
                    return acc;
                }, {} as Record<string, any>)
            }];
        });
    };

    const moveMiddleware = (index: number, direction: 'up' | 'down') => {
        setMiddlewares(prev => {
            const next = [...prev];
            const newIndex = direction === 'up' ? index - 1 : index + 1;
            if (newIndex < 0 || newIndex >= next.length) return prev;
            [next[index], next[newIndex]] = [next[newIndex], next[index]];
            return next;
        });
    };

    const updateMiddlewareConfig = (mwId: string, key: string, value: string, type: string) => {
        setMiddlewares(prev => prev.map(m => {
            if (m.id !== mwId) return m;
            let val: any = value;
            if (type === "number") {
                val = value === "" ? 0 : parseFloat(value);
            } else if (type === "boolean") {
                val = value === "true";
            }
            return { ...m, config: { ...m.config, [key]: val } };
        }));
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
                                    <select value={providerId} onChange={(e) => setProviderId(e.currentTarget.value)} class="select select-bordered w-full bg-white/5 border-white/5 rounded-xl text-sm">
                                        <option value="openai">OpenAI</option>
                                        <option value="mock">Mock</option>
                                    </select>
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Expires At</span></label>
                                    <input type="datetime-local" value={formatDateForInput(expiresAt)} onChange={handleExpiryChange} class="input input-bordered w-full bg-white/5 border-white/5 rounded-xl text-xs" />
                                </div>
                            </div>

                            <div class="border-t border-white/5 pt-6 space-y-4">
                                <div class="flex justify-between items-center">
                                    <label class="label p-0"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Middleware Composition</span></label>
                                    <div class="dropdown dropdown-end">
                                        <div tabindex={0} role="button" class="btn btn-xs btn-primary rounded-lg font-bold">Add Middleware</div>
                                        <ul tabindex={0} class="dropdown-content z-[20] menu p-2 shadow-2xl bg-base-300 border border-white/10 rounded-xl w-52 mt-2">
                                            {middlewareInfos.filter(mw => !middlewares.some(em => em.id === mw.id)).map(mw => (
                                                <li key={mw.id}><a onClick={() => toggleMiddleware(mw.id)} class="text-xs font-bold text-white/60 hover:text-white">{mw.id.replace(/_/g, " ")}</a></li>
                                            ))}
                                        </ul>
                                    </div>
                                </div>

                                <div class="grid grid-cols-1 gap-3">
                                    {middlewares.map((emw, idx) => {
                                        const mwInfo = middlewareInfos.find(m => m.id === emw.id);
                                        if (!mwInfo) return null;
                                        return (
                                            <div class="p-4 rounded-xl bg-white/5 border border-white/5 space-y-4 relative group" key={emw.id}>
                                                <div class="flex items-center justify-between">
                                                    <div class="flex items-center gap-3">
                                                        <div class="flex flex-col gap-1">
                                                            <button type="button" onClick={() => moveMiddleware(idx, 'up')} class={`btn btn-ghost btn-xs p-0 min-h-0 h-4 w-4 ${idx === 0 ? 'invisible' : ''}`}>▲</button>
                                                            <button type="button" onClick={() => moveMiddleware(idx, 'down')} class={`btn btn-ghost btn-xs p-0 min-h-0 h-4 w-4 ${idx === middlewares.length - 1 ? 'invisible' : ''}`}>▼</button>
                                                        </div>
                                                        <span class="text-xs font-bold text-white/80">{emw.id.replace(/_/g, " ")}</span>
                                                    </div>
                                                    <button type="button" onClick={() => toggleMiddleware(emw.id)} class="btn btn-ghost btn-xs h-6 w-6 btn-circle text-white/20 group-hover:text-red-400">✕</button>
                                                </div>

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
                                                                        onChange={(e) => updateMiddlewareConfig(emw.id, key, e.currentTarget.value, schema.type)}
                                                                    >
                                                                        {schema.options?.map(opt => <option value={opt} key={opt}>{opt}</option>)}
                                                                    </select>
                                                                ) : (
                                                                    <input
                                                                        type={schema.type === "number" ? "number" : "text"}
                                                                        placeholder={schema.description}
                                                                        value={emw.config[key]}
                                                                        onInput={(e) => updateMiddlewareConfig(emw.id, key, e.currentTarget.value, schema.type)}
                                                                        class="input input-bordered input-xs bg-white/5 border-white/5 rounded-lg font-mono text-[10px]"
                                                                    />
                                                                )}
                                                            </div>
                                                        );
                                                    })}
                                                </div>
                                            </div>
                                        )
                                    })}
                                </div>
                            </div>

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
