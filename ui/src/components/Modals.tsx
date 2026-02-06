import { useState, useEffect, useRef } from "preact/hooks";
import type { Key } from "../types";

export default function Modals() {
    const [editKey, setEditKey] = useState<Key | null>(null);
    const [newKeyRaw, setNewKeyRaw] = useState<string | null>(null);
    const [createMode, setCreateMode] = useState<"prepaid" | "subscription">("prepaid");
    const [showMock, setShowMock] = useState(false);
    const [showEditMock, setShowEditMock] = useState(false);

    const createModalRef = useRef<HTMLInputElement>(null);
    const editModalRef = useRef<HTMLInputElement>(null);
    const newKeyModalRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        const handleOpenEdit = (e: any) => {
            setEditKey(e.detail);
            setShowEditMock(e.detail.is_mock);
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

        const payload = {
            name: fd.get("name"),
            provider: fd.get("provider") || "openai",
            expires_at,
            budget_limit: parseFloat(fd.get("budget_limit") as string),
            budget_period: period,
            is_mock: showMock,
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
                setShowMock(false);
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

        const payload = {
            name: fd.get("name"),
            provider: fd.get("provider"),
            budget_limit: parseFloat(fd.get("budget_limit") as string),
            is_mock: showEditMock,
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
            <input type="checkbox" id="create-key-modal" class="modal-toggle peer" ref={createModalRef} />
            <div class="modal backdrop-blur-md transition-all duration-500 peer-checked:modal-open">
                <div class="modal-box w-11/12 max-w-4xl p-0 bg-[#0f0f1a]/95 backdrop-blur-2xl border border-white/5 rounded-[2rem] shadow-[0_0_50px_rgba(0,0,0,0.5)] overflow-y-auto max-h-[90vh] scale-95 opacity-0 transition-all duration-500 peer-checked:scale-100 peer-checked:opacity-100">
                    <div class="p-8 border-b border-white/5 flex justify-between items-center bg-white/[0.02]">
                        <div>
                            <h3 class="font-bold text-3xl text-white tracking-tight">Provision Access Key</h3>
                            <p class="text-sm text-white/40 mt-1.5 font-medium">Fine-tune your application's AI usage boundaries</p>
                        </div>
                        <label for="create-key-modal" class="btn btn-sm btn-circle btn-ghost text-white/20 hover:text-white/60 hover:bg-white/5">✕</label>
                    </div>

                    <form id="create-key-form" class="p-8 md:p-10 flex flex-col gap-10" onSubmit={handleCreateSubmit}>
                        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-10">
                            {/* General Settings */}
                            <div class="flex flex-col gap-8">
                                <div class="flex items-center gap-2">
                                    <div class="w-1 h-4 bg-primary rounded-full shadow-[0_0_8px_rgba(var(--p-rgb),0.6)]"></div>
                                    <h4 class="text-xs font-bold uppercase tracking-[0.2em] text-white/30">Registry</h4>
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Key Identity</span></label>
                                    <input type="text" name="name" placeholder="Deployment Identifier" class="input input-bordered w-full bg-white/5 focus:bg-white/[0.08] border-white/5 focus:border-primary/50 rounded-xl transition-all font-medium text-white placeholder:text-white/10 h-12" required />
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Target Engine</span></label>
                                    <div class="grid grid-cols-1 gap-2">
                                        <label class="flex items-center justify-between p-4 rounded-xl border border-primary/30 bg-primary/5 hover:bg-primary/10 cursor-pointer transition-all group/radio">
                                            <div class="flex items-center gap-3">
                                                <input type="radio" name="provider" value="openai" class="radio radio-primary radio-sm shadow-[0_0_10px_rgba(var(--p-rgb),0.3)]" checked />
                                                <span class="font-bold text-sm text-white">OpenAI</span>
                                            </div>
                                            <div class="w-2 h-2 rounded-full bg-primary/40 group-hover/radio:scale-150 transition-transform"></div>
                                        </label>
                                        <label class="flex items-center gap-3 p-4 rounded-xl border border-white/5 bg-white/[0.02] opacity-30 cursor-not-allowed">
                                            <input type="radio" name="provider" value="anthropic" class="radio radio-primary radio-sm" disabled />
                                            <span class="font-bold text-sm text-white/40">Anthropic (Soon)</span>
                                        </label>
                                    </div>
                                </div>
                            </div>

                            {/* Budget & Limits */}
                            <div class="flex flex-col gap-8">
                                <div class="flex items-center gap-2">
                                    <div class="w-1 h-4 bg-secondary rounded-full shadow-[0_0_8px_rgba(var(--s-rgb),0.6)]"></div>
                                    <h4 class="text-xs font-bold uppercase tracking-[0.2em] text-white/30">Economics</h4>
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Lifecycle Model</span></label>
                                    <div class="flex p-1 bg-white/5 rounded-xl border border-white/5">
                                        <button type="button" onClick={() => setCreateMode("prepaid")} class={`flex-1 py-2 text-[10px] font-black tracking-widest rounded-lg transition-all ${createMode === "prepaid" ? 'bg-white/10 text-white shadow-lg shadow-white/5' : 'text-white/20 hover:text-white/40'}`}>ONE-TIME</button>
                                        <button type="button" onClick={() => setCreateMode("subscription")} class={`flex-1 py-2 text-[10px] font-black tracking-widest rounded-lg transition-all ${createMode === "subscription" ? 'bg-white/10 text-white shadow-lg shadow-white/5' : 'text-white/20 hover:text-white/40'}`}>RECURRING</button>
                                        <input type="hidden" name="mode_type" value={createMode} />
                                    </div>
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Monetary Limit ($)</span></label>
                                    <div class="relative group">
                                        <div class="absolute inset-0 bg-primary/20 blur-xl opacity-0 group-focus-within:opacity-100 transition-opacity"></div>
                                        <span class="absolute left-4 top-1/2 -translate-y-1/2 text-primary font-bold opacity-50">$</span>
                                        <input type="number" name="budget_limit" defaultValue="5.00" step="0.01" min="0" class="relative input input-bordered w-full pl-8 bg-white/5 focus:bg-white/[0.08] border-white/5 focus:border-primary/50 rounded-xl transition-all font-mono text-2xl font-black text-white" />
                                    </div>
                                </div>
                                {createMode === "prepaid" ? (
                                    <div class="form-control">
                                        <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Expiration</span></label>
                                        <select name="expiration" class="select select-bordered w-full bg-white/5 border-white/5 rounded-xl text-sm font-bold text-white focus:border-primary/50">
                                            <option value="7">1 Week</option>
                                            <option value="30">1 Month</option>
                                            <option value="90" selected>3 Months</option>
                                            <option value="365">1 Year</option>
                                            <option value="0">Indefinite</option>
                                        </select>
                                    </div>
                                ) : (
                                    <div class="form-control">
                                        <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Reset Orbit</span></label>
                                        <select name="budget_period" class="select select-bordered w-full bg-white/5 border-white/5 rounded-xl text-sm font-bold text-white focus:border-primary/50">
                                            <option value="monthly" selected>Monthly Cycle</option>
                                            <option value="weekly">Weekly Cycle</option>
                                        </select>
                                    </div>
                                )}
                            </div>

                            {/* Safety & Mocking */}
                            <div class="flex flex-col gap-8">
                                <div class="flex items-center gap-2">
                                    <div class="w-1 h-4 bg-accent rounded-full shadow-[0_0_8px_rgba(var(--a-rgb),0.6)]"></div>
                                    <h4 class="text-xs font-bold uppercase tracking-[0.2em] text-white/30">Guardrails</h4>
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Rate Governor</span></label>
                                    <div class="flex gap-2 p-1 bg-white/5 rounded-xl border border-white/5">
                                        <input type="number" name="rate_limit" defaultValue="10" min="0" class="input input-ghost flex-1 bg-transparent font-mono text-center font-bold text-white h-10 border-none focus:outline-none" />
                                        <select name="rate_period" class="bg-white/10 rounded-lg px-3 py-1 text-[10px] font-black uppercase tracking-widest text-white/60 border-none outline-none focus:text-white transition-colors">
                                            <option value="minute" selected>RPM</option>
                                            <option value="second">RPS</option>
                                            <option value="none">UNLT</option>
                                        </select>
                                    </div>
                                </div>
                                <div class="form-control flex-1 flex flex-col">
                                    <div class="flex justify-between items-center mb-2">
                                        <label class="label py-0"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Simulation</span></label>
                                        <input type="checkbox" class="toggle toggle-primary toggle-xs" checked={showMock} onChange={() => setShowMock(!showMock)} />
                                    </div>
                                    <div class="flex-1 flex flex-col relative">
                                        {showMock ? (
                                            <textarea name="mock_config" class="textarea textarea-bordered font-mono text-[10px] leading-tight flex-1 w-full bg-black/40 border-white/5 rounded-xl p-4 text-white/60 resize-none focus:border-primary/30 transition-all" defaultValue={`{"choices":[{"message":{"content":"Hello from Mock!", "role":"assistant"}}]}`} />
                                        ) : (
                                            <div class="flex-1 flex flex-col items-center justify-center p-6 border-2 border-dashed border-white/5 rounded-2xl transition-all duration-700 opacity-20 hover:opacity-30">
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 mb-2 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" /></svg>
                                                <span class="text-[9px] font-black tracking-widest">SANDBOX OFF</span>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div class="flex justify-end gap-3 pt-10 border-t border-white/5">
                            <label for="create-key-modal" class="btn btn-ghost rounded-xl text-white/40 hover:text-white uppercase tracking-widest font-bold text-[10px]">Decline</label>
                            <button type="submit" class="btn btn-primary px-16 rounded-xl shadow-[0_0_25px_rgba(var(--p-rgb),0.3)] transform active:scale-95 transition-all font-bold uppercase tracking-widest text-[11px] h-12">Provision Key</button>
                        </div>
                    </form>
                </div>
            </div>

            {/* Edit Key Modal */}
            <input type="checkbox" id="edit-key-modal" class="modal-toggle peer" ref={editModalRef} />
            <div class="modal backdrop-blur-md transition-all duration-500 peer-checked:modal-open">
                <div class="modal-box w-11/12 max-w-2xl bg-[#0f0f1a]/95 backdrop-blur-2xl border border-white/5 rounded-[2rem] shadow-2xl p-0 scale-95 opacity-0 transition-all duration-500 peer-checked:scale-100 peer-checked:opacity-100">
                    <div class="p-8 border-b border-white/5 bg-white/[0.02] flex justify-between items-center">
                        <h3 class="font-bold text-2xl text-white tracking-tight">Reconfigure Access</h3>
                        <label for="edit-key-modal" class="btn btn-sm btn-circle btn-ghost text-white/20">✕</label>
                    </div>
                    {editKey && (
                        <form id="edit-key-form" class="p-8 flex flex-col gap-8" onSubmit={handleEditSubmit}>
                            <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Identity</span></label>
                                    <input type="text" name="name" defaultValue={editKey.name} class="input input-bordered w-full bg-white/5 border-white/5 focus:border-primary/50 rounded-xl font-bold text-white h-12" required />
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Provider</span></label>
                                    <select name="provider" defaultValue={editKey.provider} class="select select-bordered w-full bg-white/5 border-white/5 rounded-xl font-bold text-white h-12">
                                        <option value="openai">OpenAI</option>
                                        <option value="anthropic" disabled>Anthropic (Soon)</option>
                                    </select>
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">New Limit ($)</span></label>
                                    <input type="number" name="budget_limit" defaultValue={editKey.budget_limit} step="0.01" min="0" class="input input-bordered w-full bg-white/5 border-white/5 focus:border-primary/50 rounded-xl font-mono text-xl font-bold text-white h-12" />
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Rate Governor</span></label>
                                    <div class="flex gap-2">
                                        <input type="number" name="rate_limit" defaultValue={editKey.rate_limit} min="0" class="input input-bordered flex-1 bg-white/5 border-white/5 rounded-xl font-mono font-bold text-white h-12" />
                                        <select name="rate_period" defaultValue={editKey.rate_period} class="select select-bordered bg-white/5 border-white/5 rounded-xl text-[10px] font-black uppercase h-12">
                                            <option value="minute">RPM</option>
                                            <option value="second">RPS</option>
                                            <option value="none">UNLT</option>
                                        </select>
                                    </div>
                                </div>
                            </div>
                            <div class="form-control p-6 rounded-2xl bg-white/5 border border-white/5">
                                <label class="label cursor-pointer justify-between p-0 mb-4">
                                    <div class="flex flex-col">
                                        <span class="font-bold text-sm text-white/80 tracking-tight">Active Sandbox Mode</span>
                                        <span class="text-[10px] uppercase font-bold text-white/20 tracking-[0.1em] mt-0.5 whitespace-nowrap overflow-hidden text-ellipsis mr-4">Override live calls with static JSON</span>
                                    </div>
                                    <input type="checkbox" class="toggle toggle-primary toggle-sm shrink-0" checked={showEditMock} onChange={() => setShowEditMock(!showEditMock)} />
                                </label>
                                {showEditMock && (
                                    <div class="animate-in fade-in slide-in-from-top-2 duration-500">
                                        <textarea name="mock_config" defaultValue={editKey.mock_config} class="textarea textarea-bordered w-full h-40 font-mono text-[10px] leading-tight bg-black/40 border-white/5 rounded-xl text-white/60 focus:border-primary/30 transition-all"></textarea>
                                    </div>
                                )}
                            </div>
                            <div class="flex justify-end gap-3 pt-6 border-t border-white/5">
                                <label for="edit-key-modal" class="btn btn-ghost rounded-xl text-white/40 font-bold uppercase tracking-widest text-[10px]">Cancel</label>
                                <button type="submit" class="btn btn-primary px-12 rounded-xl font-bold uppercase tracking-widest text-[11px] h-12 shadow-[0_0_20px_rgba(var(--p-rgb),0.2)]">Commit Changes</button>
                            </div>
                        </form>
                    )}
                </div>
            </div>

            {/* New Key Display Modal */}
            <input type="checkbox" id="new-key-display-modal" class="modal-toggle peer" ref={newKeyModalRef} />
            <div class="modal backdrop-blur-md transition-all duration-500 peer-checked:modal-open">
                <div class="modal-box relative max-w-lg w-11/12 p-10 bg-[#0f0f1a]/95 backdrop-blur-2xl border border-white/5 rounded-[2.5rem] shadow-[0_0_100px_rgba(var(--p-rgb),0.2)] text-center scale-95 opacity-0 transition-all duration-500 peer-checked:scale-100 peer-checked:opacity-100">
                    <div class="w-20 h-20 rounded-full bg-success/20 flex items-center justify-center mx-auto mb-6 shadow-[0_0_30px_rgba(var(--s-rgb),0.2)] animate-pulse">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                    </div>
                    <h3 class="font-bold text-3xl text-white tracking-tight mb-2">Secret Key Created</h3>
                    <p class="text-white/40 font-medium mb-10 leading-relaxed text-sm">Persist this key securely. Due to security protocols, <span class="text-error/80 font-bold">it cannot be retrieved again</span> once dismissed.</p>

                    <div class="relative group mb-10">
                        <div class="absolute -inset-1 bg-gradient-to-r from-primary to-secondary rounded-2xl blur opacity-25 group-hover:opacity-40 transition-opacity"></div>
                        <div class="relative bg-black/60 backdrop-blur-xl p-6 rounded-2xl border border-white/10 flex flex-col gap-4">
                            <code class="break-all font-mono font-black text-2xl text-primary tracking-tight selection:bg-primary selection:text-white px-2 cursor-pointer" onClick={() => {
                                if (newKeyRaw) navigator.clipboard.writeText(newKeyRaw);
                            }}>{newKeyRaw || "pk-xxxxxxxx"}</code>
                            <button class="btn btn-sm btn-ghost bg-white/5 hover:bg-white/10 rounded-xl text-[10px] font-black uppercase tracking-[0.2em] text-white/40 h-10 transition-all" onClick={() => {
                                if (newKeyRaw) navigator.clipboard.writeText(newKeyRaw);
                            }}>
                                Copy to Clipboard
                            </button>
                        </div>
                    </div>

                    <button class="w-full btn btn-primary btn-lg rounded-2xl font-black uppercase tracking-[0.2em] shadow-[0_0_30px_rgba(var(--p-rgb),0.3)] hover:scale-[1.02] active:scale-95 transition-all text-sm h-14" onClick={() => window.location.reload()}>
                        Acknowledge & Finalize
                    </button>
                </div>
            </div>
        </>
    );
}
