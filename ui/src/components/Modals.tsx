import { useState, useEffect, useRef } from "preact/hooks";
import type { Key } from "../types";

export default function Modals() {
    const [editKey, setEditKey] = useState<Key | null>(null);
    const [newKeyRaw, setNewKeyRaw] = useState<string | null>(null);
    const [createMode, setCreateMode] = useState<"prepaid" | "subscription">("prepaid");
    const [createProvider, setCreateProvider] = useState("openai");
    const [editProvider, setEditProvider] = useState("openai");
    const [newKeyCopied, setNewKeyCopied] = useState(false);

    const createModalRef = useRef<HTMLInputElement>(null);
    const editModalRef = useRef<HTMLInputElement>(null);
    const newKeyModalRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        const handleOpenEdit = (e: any) => {
            setEditKey(e.detail);
            setEditProvider(e.detail.provider);
            if (editModalRef.current) editModalRef.current.checked = true;
        };

        window.addEventListener('open-edit-modal', handleOpenEdit);
        return () => window.removeEventListener('open-edit-modal', handleOpenEdit);
    }, []);

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

        const provider = fd.get("provider") || "openai";
        const payload = {
            name: fd.get("name"),
            provider: provider,
            expires_at,
            budget_limit: parseFloat(fd.get("budget_limit") as string),
            budget_period: period,
            is_mock: provider === "mock",
            mock_config: fd.get("mock_config"),
            rate_limit: parseInt(fd.get("rate_limit") as string) || 10,
            rate_period: fd.get("rate_period") || "minute",
        };

        try {
            const res = await fetch("/v1/config/app-keys", {
                method: "POST",
                body: JSON.stringify(payload),
                headers: { "Content-Type": "application/json" },
            });

            if (res.ok) {
                const data = await res.json();
                setNewKeyRaw(data.key);
                if (createModalRef.current) createModalRef.current.checked = false;
                if (newKeyModalRef.current) newKeyModalRef.current.checked = true;
                form.reset();
                setCreateProvider("openai");
                setCreateMode("prepaid");
            } else {
                alert("Failed to create key");
            }
        } catch (err) {
            console.error("Create error:", err);
        }
    };

    const handleEditSubmit = async (e: Event) => {
        e.preventDefault();
        if (!editKey) return;

        const form = e.target as HTMLFormElement;
        const fd = new FormData(form);
        const provider = fd.get("provider");

        const payload = {
            name: fd.get("name"),
            provider: provider,
            budget_limit: parseFloat(fd.get("budget_limit") as string),
            is_mock: provider === "mock",
            mock_config: fd.get("mock_config"),
            rate_limit: parseInt(fd.get("rate_limit") as string) || 10,
            rate_period: fd.get("rate_period") || "minute",
        };

        try {
            const res = await fetch(`/v1/config/app-keys/${editKey.id}`, {
                method: "PUT",
                body: JSON.stringify(payload),
                headers: { "Content-Type": "application/json" },
            });

            if (res.ok) {
                if (editModalRef.current) editModalRef.current.checked = false;
                window.location.reload(); // Quick way to refresh Dashboard island if separate
            } else {
                alert("Failed to update key");
            }
        } catch (err) {
            console.error("Edit error:", err);
        }
    };

    return (
        <>
            {/* Create Key Modal */}
            <input type="checkbox" id="create-key-modal" class="modal-toggle" ref={createModalRef} />
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

            {/* Edit Key Modal */}
            <input type="checkbox" id="edit-key-modal" class="modal-toggle" ref={editModalRef} />
            <div class="modal">
                <div class="modal-box w-11/12 max-w-xl bg-base-100 border border-white/5 rounded-2xl shadow-2xl p-0 overflow-visible">
                    <div class="p-6 border-b border-white/5 bg-base-200/50 rounded-t-2xl flex justify-between items-center">
                        <h3 class="font-bold text-xl text-white tracking-tight">Edit API Key</h3>
                        <label for="edit-key-modal" class="btn btn-sm btn-circle btn-ghost">✕</label>
                    </div>
                    {editKey && (
                        <form id="edit-key-form" class="p-6 flex flex-col gap-6" onSubmit={handleEditSubmit}>
                            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Name</span></label>
                                    <input type="text" name="name" defaultValue={editKey.name} class="input input-bordered w-full bg-white/5 border-white/5 rounded-xl font-bold" required />
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Provider</span></label>
                                    <select name="provider" defaultValue={editKey.provider} class="select select-bordered w-full bg-white/5 border-white/5 rounded-xl text-sm" onChange={(e) => setEditProvider(e.currentTarget.value)}>
                                        <option value="openai">OpenAI</option>
                                        <option value="mock">Mock</option>
                                        <option value="anthropic" disabled>Anthropic (Soon)</option>
                                    </select>
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">New Limit ($)</span></label>
                                    <input type="number" name="budget_limit" defaultValue={editKey.budget_limit} step="0.01" min="0" class="input input-bordered w-full bg-white/5 border-white/5 rounded-xl font-mono" />
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Rate Limit</span></label>
                                    <div class="flex flex-row gap-2">
                                        <input type="number" name="rate_limit" defaultValue={editKey.rate_limit} min="0" class="input input-bordered flex-1 min-w-0 bg-white/5 border-white/5 rounded-xl font-mono" />
                                        <select name="rate_period" defaultValue={editKey.rate_period} class="select select-bordered bg-white/5 border-white/5 rounded-xl text-[10px] font-bold uppercase w-20">
                                            <option value="minute">RPM</option>
                                            <option value="second">RPS</option>
                                            <option value="none">UNLT</option>
                                        </select>
                                    </div>
                                </div>
                            </div>
                            {editProvider === "mock" && (
                                <div class="form-control p-4 rounded-xl bg-white/5 border border-white/5">
                                    <label class="label mb-2">
                                        <span class="font-bold text-sm text-white/80">Mock Configuration</span>
                                    </label>
                                    <textarea name="mock_config" defaultValue={editKey.mock_config} class="textarea textarea-bordered w-full h-32 font-mono text-[10px] bg-black/20 border-white/5 rounded-xl"></textarea>
                                </div>
                            )}
                            <div class="flex justify-end gap-3 pt-6 border-t border-white/5">
                                <label for="edit-key-modal" class="btn btn-ghost rounded-xl text-white/40 font-bold uppercase tracking-widest text-[10px]">Cancel</label>
                                <button type="submit" class="btn btn-primary px-10 rounded-xl font-bold uppercase tracking-widest text-[11px] h-11 shadow-lg shadow-primary/20">Save Changes</button>
                            </div>
                        </form>
                    )}
                </div>
            </div>

            {/* New Key Display Modal */}
            <input type="checkbox" id="new-key-display-modal" class="modal-toggle" ref={newKeyModalRef} />
            <div class="modal">
                <div class="modal-box max-w-md w-11/12 p-8 bg-base-100 border border-white/5 rounded-2xl shadow-2xl text-center">
                    <div class="w-16 h-16 rounded-full bg-success/10 flex items-center justify-center mx-auto mb-4">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                    </div>
                    <h3 class="font-bold text-2xl text-white tracking-tight mb-2">API Key Generated</h3>
                    <p class="text-white/40 text-sm mb-8">Save this key now. It <span class="text-error/80 font-bold">won't be shown again</span>.</p>

                    <div class="bg-black/20 p-4 rounded-xl border border-white/5 flex flex-col gap-3 mb-8">
                        <code class="break-all font-mono font-bold text-xl text-primary tracking-tight">{newKeyRaw || "pk-xxxxxxxx"}</code>
                        <button class={`btn btn-sm btn-ghost bg-white/5 hover:bg-white/10 rounded-lg text-[10px] font-bold uppercase tracking-widest h-9 transition-all ${newKeyCopied ? 'text-success bg-success/10' : ''}`} onClick={() => {
                            if (newKeyRaw) {
                                navigator.clipboard.writeText(newKeyRaw)
                                    .then(() => {
                                        setNewKeyCopied(true);
                                        setTimeout(() => setNewKeyCopied(false), 2000);
                                    })
                                    .catch(err => console.error("Failed to copy:", err));
                            }
                        }}>
                            {newKeyCopied ? (
                                <span class="flex items-center gap-2">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    Copied!
                                </span>
                            ) : (
                                "Copy to Clipboard"
                            )}
                        </button>
                    </div>

                    <button class="w-full btn btn-primary rounded-xl font-bold uppercase tracking-widest text-xs h-12 shadow-lg shadow-primary/20" onClick={() => window.location.reload()}>
                        Done
                    </button>
                </div>
            </div>
        </>
    );
}
