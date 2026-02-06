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
                <div class="modal-box w-11/12 max-w-4xl p-0 bg-[#07070c] border border-white/[0.05] rounded-[2.5rem] shadow-2xl overflow-y-auto max-h-[90vh] transition-all duration-300 translate-y-4 peer-checked:translate-y-0">
                    <div class="p-10 border-b border-white/[0.05] flex justify-between items-center bg-white/[0.01]">
                        <div>
                            <h3 class="font-bold text-3xl text-white tracking-tight">Generate API Key</h3>
                            <p class="text-sm text-white/20 mt-2 font-medium">Configure usage limits and safety guardrails for your new key.</p>
                        </div>
                        <label for="create-key-modal" class="btn btn-sm btn-circle btn-ghost text-white/20 hover:text-white/60 hover:bg-white/5">✕</label>
                    </div>

                    <form id="create-key-form" class="p-10 flex flex-col gap-12" onSubmit={handleCreateSubmit}>
                        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-12">
                            {/* General Settings */}
                            <div class="flex flex-col gap-8">
                                <h4 class="text-[10px] font-bold uppercase tracking-[0.2em] text-white/20">Identity</h4>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/40 text-[9px] uppercase tracking-widest">Key Name</span></label>
                                    <input type="text" name="name" placeholder="e.g. Production Frontend" class="input input-bordered w-full bg-white/5 border-white/[0.05] focus:border-primary/50 rounded-xl transition-all font-medium text-white h-12" required />
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/40 text-[9px] uppercase tracking-widest">Provider</span></label>
                                    <div class="grid grid-cols-1 gap-3">
                                        <label class="flex items-center justify-between p-4 rounded-xl border border-primary/20 bg-primary/5 cursor-pointer transition-all">
                                            <div class="flex items-center gap-3">
                                                <input type="radio" name="provider" value="openai" class="radio radio-primary radio-sm" checked />
                                                <span class="font-bold text-sm text-white">OpenAI</span>
                                            </div>
                                        </label>
                                        <label class="flex items-center gap-3 p-4 rounded-xl border border-white/[0.05] bg-white/[0.02] opacity-30 cursor-not-allowed">
                                            <span class="font-bold text-sm text-white/20">Anthropic (Soon)</span>
                                        </label>
                                    </div>
                                </div>
                            </div>

                            {/* Budget & Limits */}
                            <div class="flex flex-col gap-8">
                                <h4 class="text-[10px] font-bold uppercase tracking-[0.2em] text-white/20">Economics</h4>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/40 text-[9px] uppercase tracking-widest">Billing Mode</span></label>
                                    <div class="flex p-1 bg-white/5 rounded-xl border border-white/[0.05]">
                                        <button type="button" onClick={() => setCreateMode("prepaid")} class={`flex-1 py-2 text-[9px] font-bold tracking-widest rounded-lg transition-all ${createMode === "prepaid" ? 'bg-white/10 text-white' : 'text-white/20'}`}>ONE-TIME</button>
                                        <button type="button" onClick={() => setCreateMode("subscription")} class={`flex-1 py-2 text-[9px] font-bold tracking-widest rounded-lg transition-all ${createMode === "subscription" ? 'bg-white/10 text-white' : 'text-white/20'}`}>RECURRING</button>
                                        <input type="hidden" name="mode_type" value={createMode} />
                                    </div>
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/40 text-[9px] uppercase tracking-widest">Spending Limit ($)</span></label>
                                    <input type="number" name="budget_limit" defaultValue="5.00" step="0.01" min="0" class="input input-bordered w-full bg-white/5 border-white/[0.05] focus:border-primary/50 rounded-xl transition-all font-mono text-xl font-bold text-white h-12" />
                                </div>
                                {createMode === "prepaid" ? (
                                    <div class="form-control">
                                        <label class="label"><span class="label-text font-bold text-white/40 text-[9px] uppercase tracking-widest">Expiration</span></label>
                                        <select name="expiration" class="select select-bordered w-full bg-white/5 border-white/[0.05] rounded-xl text-xs font-bold text-white">
                                            <option value="7">1 Week</option>
                                            <option value="30">1 Month</option>
                                            <option value="90" selected>3 Months</option>
                                            <option value="0">Never</option>
                                        </select>
                                    </div>
                                ) : (
                                    <div class="form-control">
                                        <label class="label"><span class="label-text font-bold text-white/40 text-[9px] uppercase tracking-widest">Reset Period</span></label>
                                        <select name="budget_period" class="select select-bordered w-full bg-white/5 border-white/[0.05] rounded-xl text-xs font-bold text-white">
                                            <option value="monthly" selected>Monthly</option>
                                            <option value="weekly">Weekly</option>
                                        </select>
                                    </div>
                                )}
                            </div>

                            {/* Guardrails */}
                            <div class="flex flex-col gap-8">
                                <h4 class="text-[10px] font-bold uppercase tracking-[0.2em] text-white/20">Guardrails</h4>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/40 text-[9px] uppercase tracking-widest">Rate Limit</span></label>
                                    <div class="flex gap-2 p-1 bg-white/5 rounded-xl border border-white/[0.05]">
                                        <input type="number" name="rate_limit" defaultValue="10" min="0" class="input input-ghost flex-1 bg-transparent font-mono text-center font-bold text-white h-10 border-none" />
                                        <select name="rate_period" class="bg-white/10 rounded-lg px-3 py-1 text-[9px] font-bold uppercase text-white/60">
                                            <option value="minute" selected>RPM</option>
                                            <option value="second">RPS</option>
                                        </select>
                                    </div>
                                </div>
                                <div class="form-control flex-1 flex flex-col">
                                    <div class="flex justify-between items-center mb-4">
                                        <label class="label py-0"><span class="label-text font-bold text-white/40 text-[9px] uppercase tracking-widest">Simulation Mode</span></label>
                                        <input type="checkbox" class="toggle toggle-primary toggle-sm" checked={showMock} onChange={() => setShowMock(!showMock)} />
                                    </div>
                                    <div class="flex-1 flex flex-col">
                                        {showMock ? (
                                            <textarea name="mock_config" class="textarea textarea-bordered font-mono text-[10px] leading-tight flex-1 w-full bg-black/40 border-white/[0.05] rounded-xl p-4 text-white/40 resize-none" defaultValue={`{"choices":[{"message":{"content":"Hello from Mock!", "role":"assistant"}}]}`} />
                                        ) : (
                                            <div class="flex-1 flex flex-col items-center justify-center p-6 border border-dashed border-white/[0.05] rounded-xl opacity-20">
                                                <span class="text-[9px] font-bold tracking-widest">OFF</span>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div class="flex justify-end gap-3 pt-10 border-t border-white/[0.05]">
                            <label for="create-key-modal" class="btn btn-ghost rounded-xl text-white/20 hover:text-white uppercase tracking-widest font-bold text-[10px]">Close</label>
                            <button type="submit" class="btn btn-primary px-12 rounded-xl font-bold uppercase tracking-widest text-[10px] h-12">Generate Key</button>
                        </div>
                    </form>
                </div>
            </div>

            {/* Edit Key Modal */}
            <input type="checkbox" id="edit-key-modal" class="modal-toggle peer" ref={editModalRef} />
            <div class="modal backdrop-blur-md transition-all duration-500 peer-checked:modal-open">
                <div class="modal-box w-11/12 max-w-2xl bg-[#07070c] border border-white/[0.05] rounded-[2.5rem] shadow-2xl p-0 transition-all duration-300 translate-y-4 peer-checked:translate-y-0">
                    <div class="p-10 border-b border-white/[0.05] bg-white/[0.01] flex justify-between items-center">
                        <h3 class="font-bold text-2xl text-white tracking-tight">Edit API Key</h3>
                        <label for="edit-key-modal" class="btn btn-sm btn-circle btn-ghost text-white/20">✕</label>
                    </div>
                    {editKey && (
                        <form id="edit-key-form" class="p-10 flex flex-col gap-8" onSubmit={handleEditSubmit}>
                            <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/40 text-[9px] uppercase tracking-widest">Identity</span></label>
                                    <input type="text" name="name" defaultValue={editKey.name} class="input input-bordered w-full bg-white/5 border-white/[0.05] rounded-xl font-bold text-white h-12" required />
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/40 text-[9px] uppercase tracking-widest">Provider</span></label>
                                    <select name="provider" defaultValue={editKey.provider} class="select select-bordered w-full bg-white/5 border-white/[0.05] rounded-xl font-bold text-white h-12">
                                        <option value="openai">OpenAI</option>
                                    </select>
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/40 text-[9px] uppercase tracking-widest">Spending Limit ($)</span></label>
                                    <input type="number" name="budget_limit" defaultValue={editKey.budget_limit} step="0.01" min="0" class="input input-bordered w-full bg-white/5 border-white/[0.05] rounded-xl font-mono text-xl font-bold text-white h-12" />
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-white/40 text-[9px] uppercase tracking-widest">Rate Limit</span></label>
                                    <div class="flex gap-2">
                                        <input type="number" name="rate_limit" defaultValue={editKey.rate_limit} min="0" class="input input-bordered flex-1 bg-white/5 border-white/[0.05] rounded-xl font-mono font-bold text-white h-12" />
                                        <select name="rate_period" defaultValue={editKey.rate_period} class="select select-bordered bg-white/5 border-white/[0.05] rounded-xl text-[10px] font-black uppercase h-12">
                                            <option value="minute">RPM</option>
                                            <option value="second">RPS</option>
                                        </select>
                                    </div>
                                </div>
                            </div>
                            <div class="form-control p-8 rounded-2xl bg-white/[0.02] border border-white/[0.05]">
                                <label class="label cursor-pointer justify-between p-0 mb-6">
                                    <div class="flex flex-col">
                                        <span class="font-bold text-sm text-white/80">Simulation Mode</span>
                                        <span class="text-[9px] uppercase font-bold text-white/20 tracking-widest mt-1">Return static JSON instead of live calls</span>
                                    </div>
                                    <input type="checkbox" class="toggle toggle-primary toggle-sm" checked={showEditMock} onChange={() => setShowEditMock(!showEditMock)} />
                                </label>
                                {showEditMock && (
                                    <div class="animate-in fade-in slide-in-from-top-2 duration-300">
                                        <textarea name="mock_config" defaultValue={editKey.mock_config} class="textarea textarea-bordered w-full h-40 font-mono text-[10px] leading-tight bg-black/40 border-white/[0.05] rounded-xl text-white/40"></textarea>
                                    </div>
                                )}
                            </div>
                            <div class="flex justify-end gap-3 pt-6 border-t border-white/[0.05]">
                                <label for="edit-key-modal" class="btn btn-ghost rounded-xl text-white/20 font-bold uppercase tracking-widest text-[9px]">Cancel</label>
                                <button type="submit" class="btn btn-primary px-10 rounded-xl font-bold uppercase tracking-widest text-[9px] h-12">Save Changes</button>
                            </div>
                        </form>
                    )}
                </div>
            </div>

            {/* New Key Display Modal */}
            <input type="checkbox" id="new-key-display-modal" class="modal-toggle peer" ref={newKeyModalRef} />
            <div class="modal backdrop-blur-md transition-all duration-500 peer-checked:modal-open">
                <div class="modal-box relative max-w-lg w-11/12 p-12 bg-[#07070c] border border-white/[0.05] rounded-[2.5rem] shadow-2xl text-center transition-all duration-300 translate-y-4 peer-checked:translate-y-0">
                    <div class="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center mx-auto mb-8">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                    </div>
                    <h3 class="font-bold text-3xl text-white tracking-tight mb-4">Key Generated</h3>
                    <p class="text-white/30 font-medium mb-12 leading-relaxed text-sm">Make sure to copy your secret key now. You won't be able to see it again.</p>

                    <div class="relative mb-12">
                        <div class="relative bg-white/[0.02] p-8 rounded-2xl border border-white/[0.05] flex flex-col gap-6">
                            <code class="break-all font-mono font-bold text-2xl text-primary tracking-tight" onClick={() => {
                                if (newKeyRaw) navigator.clipboard.writeText(newKeyRaw);
                            }}>{newKeyRaw || "sk-xxxxxxxx"}</code>
                            <button class="text-[9px] font-bold uppercase tracking-widest text-white/20 hover:text-white transition-colors" onClick={() => {
                                if (newKeyRaw) navigator.clipboard.writeText(newKeyRaw);
                            }}>
                                Copy to Clipboard
                            </button>
                        </div>
                    </div>

                    <button class="w-full btn btn-primary h-14 rounded-2xl font-bold uppercase tracking-widest text-xs" onClick={() => window.location.reload()}>
                        Done
                    </button>
                </div>
            </div>
        </>

    );
}
