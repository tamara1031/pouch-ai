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
            <input type="checkbox" id="create-key-modal" class="modal-toggle" ref={createModalRef} />
            <div class="modal backdrop-blur-md">
                <div class="modal-box w-11/12 max-w-4xl p-0 bg-base-100/95 backdrop-blur-xl border border-base-content/5 rounded-2xl shadow-2xl overflow-y-auto max-h-[90vh]">
                    <div class="p-6 border-b border-base-content/5 flex justify-between items-center bg-base-200/50">
                        <div>
                            <h3 class="font-bold text-2xl">Create API Key</h3>
                            <p class="text-sm text-base-content/50 mt-1">Configure model access and usage policies</p>
                        </div>
                        <label for="create-key-modal" class="btn btn-sm btn-circle btn-ghost text-base-content/50">✕</label>
                    </div>

                    <form id="create-key-form" class="p-6 md:p-8 flex flex-col gap-8" onSubmit={handleCreateSubmit}>
                        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 md:gap-10">
                            {/* General Settings */}
                            <div class="flex flex-col gap-6">
                                <h4 class="text-xs font-bold uppercase tracking-wider text-primary/70">General Settings</h4>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-base-content/80">Key Name</span></label>
                                    <input type="text" name="name" placeholder="e.g. Mobile App v2" class="input input-bordered w-full bg-base-200/50 focus:bg-base-100 border-base-content/10 transition-all font-medium" required />
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-base-content/80">LLM Provider</span></label>
                                    <div class="grid grid-cols-1 gap-2">
                                        <label class="flex items-center gap-3 p-3 rounded-xl border border-primary bg-primary/5 cursor-pointer transition-all">
                                            <input type="radio" name="provider" value="openai" class="radio radio-primary radio-sm" checked />
                                            <div class="flex items-center gap-2">
                                                <img src="https://simpleicons.org/icons/openai.svg" class="w-4 h-4 dark:invert" />
                                                <span class="font-semibold text-sm">OpenAI</span>
                                            </div>
                                        </label>
                                        <label class="flex items-center gap-3 p-3 rounded-xl border border-base-content/10 hover:border-primary/50 opacity-50 cursor-not-allowed">
                                            <input type="radio" name="provider" value="anthropic" class="radio radio-primary radio-sm" disabled />
                                            <div class="flex items-center gap-2"><span class="font-semibold text-sm">Anthropic (Soon)</span></div>
                                        </label>
                                    </div>
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-base-content/80">Mock Mode</span></label>
                                    <label class="flex items-center justify-between p-3 rounded-xl bg-base-200/50 border border-base-content/5 cursor-pointer">
                                        <div>
                                            <span class="text-sm font-medium">Simulated Response</span>
                                            <p class="text-[10px] opacity-50">Free & fast. No external calls.</p>
                                        </div>
                                        <input type="checkbox" class="toggle toggle-primary toggle-sm" checked={showMock} onChange={() => setShowMock(!showMock)} />
                                    </label>
                                </div>
                            </div>

                            {/* Budget & Limits */}
                            <div class="flex flex-col gap-6">
                                <h4 class="text-xs font-bold uppercase tracking-wider text-primary/70">Budget & Limits</h4>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-base-content/80">Budget Type</span></label>
                                    <div class="join w-full">
                                        <input class="join-item btn btn-sm flex-1" type="radio" name="mode_type" value="prepaid" aria-label="One-time" checked={createMode === "prepaid"} onChange={() => setCreateMode("prepaid")} />
                                        <input class="join-item btn btn-sm flex-1" type="radio" name="mode_type" value="subscription" aria-label="Recurring" checked={createMode === "subscription"} onChange={() => setCreateMode("subscription")} />
                                    </div>
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-base-content/80">Amount ($ USD)</span></label>
                                    <input type="number" name="budget_limit" defaultValue="5.00" step="0.01" min="0" class="input input-bordered w-full font-mono text-center bg-base-200/50 text-xl font-bold" />
                                </div>
                                {createMode === "prepaid" ? (
                                    <div class="form-control">
                                        <label class="label"><span class="label-text font-bold text-base-content/80">Auto Expire</span></label>
                                        <select name="expiration" class="select select-bordered w-full bg-base-200/50 text-sm">
                                            <option value="7">7 Days</option>
                                            <option value="30">30 Days</option>
                                            <option value="90" selected>90 Days</option>
                                            <option value="365">1 Year</option>
                                            <option value="0">Never</option>
                                        </select>
                                    </div>
                                ) : (
                                    <div class="form-control">
                                        <label class="label"><span class="label-text font-bold text-base-content/80">Reset Frequency</span></label>
                                        <select name="budget_period" class="select select-bordered w-full bg-base-200/50 text-sm">
                                            <option value="monthly" selected>Every Month</option>
                                            <option value="weekly">Every Week</option>
                                        </select>
                                    </div>
                                )}
                            </div>

                            {/* Safety & Mocking */}
                            <div class="flex flex-col gap-6">
                                <h4 class="text-xs font-bold uppercase tracking-wider text-primary/70">Safety & Mocking</h4>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-base-content/80">🚦 Rate Limit</span></label>
                                    <div class="flex gap-2">
                                        <input type="number" name="rate_limit" defaultValue="10" min="0" class="input input-bordered flex-1 bg-base-200/50 font-mono text-center" />
                                        <select name="rate_period" class="select select-bordered bg-base-200/50 text-xs">
                                            <option value="minute" selected>/ min</option>
                                            <option value="second">/ sec</option>
                                            <option value="none">None</option>
                                        </select>
                                    </div>
                                </div>
                                {showMock ? (
                                    <div class="flex-1 flex flex-col">
                                        <label class="label"><span class="label-text font-bold text-base-content/80">Mock Response JSON</span></label>
                                        <textarea name="mock_config" class="textarea textarea-bordered font-mono text-[10px] leading-tight flex-1 w-full bg-base-300/50 border-base-content/5 min-h-[160px] resize-none" defaultValue={`{"choices":[{"message":{"content":"Hello from Mock!", "role":"assistant"}}]}`} />
                                    </div>
                                ) : (
                                    <div class="flex-1 flex flex-col items-center justify-center p-6 border-2 border-dashed border-base-content/10 rounded-2xl opacity-40">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" /></svg>
                                        <span class="text-[10px] font-bold">MOCK CONFIG INACTIVE</span>
                                    </div>
                                )}
                            </div>
                        </div>

                        <div class="flex justify-end gap-3 pt-6 border-t border-base-content/5">
                            <label for="create-key-modal" class="btn btn-ghost">Cancel</label>
                            <button type="submit" class="btn btn-primary px-12">Generate Access Key</button>
                        </div>
                    </form>
                </div>
            </div>

            {/* Edit Key Modal */}
            <input type="checkbox" id="edit-key-modal" class="modal-toggle" ref={editModalRef} />
            <div class="modal backdrop-blur-md">
                <div class="modal-box w-11/12 max-w-2xl bg-base-100/95 backdrop-blur-xl border border-base-content/5 rounded-2xl shadow-2xl p-0 overflow-y-auto max-h-[90vh]">
                    <div class="p-6 border-b border-base-content/5 bg-base-200/50 flex justify-between items-center">
                        <h3 class="font-bold text-xl">Edit Access Key</h3>
                        <label for="edit-key-modal" class="btn btn-sm btn-circle btn-ghost text-base-content/50">✕</label>
                    </div>
                    {editKey && (
                        <form id="edit-key-form" class="p-6 flex flex-col gap-6" onSubmit={handleEditSubmit}>
                            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-base-content/80">Key Name</span></label>
                                    <input type="text" name="name" defaultValue={editKey.name} class="input input-bordered w-full bg-base-200/50" required />
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-base-content/80">LLM Provider</span></label>
                                    <select name="provider" defaultValue={editKey.provider} class="select select-bordered w-full bg-base-200/50">
                                        <option value="openai">OpenAI</option>
                                        <option value="anthropic" disabled>Anthropic (Soon)</option>
                                    </select>
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-base-content/80">Budget Limit ($)</span></label>
                                    <input type="number" name="budget_limit" defaultValue={editKey.budget_limit} step="0.01" min="0" class="input input-bordered w-full bg-base-200/50 font-mono" />
                                </div>
                                <div class="form-control">
                                    <label class="label"><span class="label-text font-bold text-base-content/80">🚦 Rate Limit</span></label>
                                    <div class="flex gap-2">
                                        <input type="number" name="rate_limit" defaultValue={editKey.rate_limit} min="0" class="input input-bordered flex-1 bg-base-200/50 font-mono" />
                                        <select name="rate_period" defaultValue={editKey.rate_period} class="select select-bordered bg-base-200/50 text-xs">
                                            <option value="minute">/ min</option>
                                            <option value="second">/ sec</option>
                                            <option value="none">None</option>
                                        </select>
                                    </div>
                                </div>
                            </div>
                            <div class="form-control p-4 rounded-xl bg-base-200/50 border border-base-content/5">
                                <label class="label cursor-pointer justify-between p-0">
                                    <div class="flex flex-col">
                                        <span class="font-bold text-sm">Active Mock Mode</span>
                                        <span class="text-[10px] opacity-50">Returns static JSON instead of calling API</span>
                                    </div>
                                    <input type="checkbox" class="toggle toggle-primary toggle-sm" checked={showEditMock} onChange={() => setShowEditMock(!showEditMock)} />
                                </label>
                                {showEditMock && (
                                    <div class="mt-4 animate-in fade-in slide-in-from-top-2">
                                        <textarea name="mock_config" defaultValue={editKey.mock_config} class="textarea textarea-bordered w-full h-32 font-mono text-[10px] bg-base-300/50"></textarea>
                                    </div>
                                )}
                            </div>
                            <div class="flex justify-end gap-3 pt-4 border-t border-base-content/5">
                                <label for="edit-key-modal" class="btn btn-ghost">Cancel</label>
                                <button type="submit" class="btn btn-primary px-10">Save Changes</button>
                            </div>
                        </form>
                    )}
                </div>
            </div>

            {/* New Key Display Modal */}
            <input type="checkbox" id="new-key-display-modal" class="modal-toggle" ref={newKeyModalRef} />
            <div class="modal backdrop-blur-md">
                <div class="modal-box relative max-w-lg w-11/12 p-8 bg-base-100/95 backdrop-blur-xl border border-base-content/5 rounded-2xl shadow-2xl overflow-y-auto max-h-[90vh]">
                    <h3 class="font-bold text-xl text-success flex items-center gap-2 mb-4">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                        Key Generated Successfully!
                    </h3>
                    <p class="py-4">Please copy your API key now.<br /><span class="font-bold text-error">It will not be shown again.</span></p>
                    <div class="bg-base-200 p-4 rounded-lg flex justify-between items-center">
                        <code class="break-all font-mono font-bold text-lg">{newKeyRaw || "pk-xxxxxxxx"}</code>
                        <button class="btn btn-sm btn-ghost" onClick={() => {
                            if (newKeyRaw) navigator.clipboard.writeText(newKeyRaw);
                        }}>Copy</button>
                    </div>
                    <div class="modal-action">
                        <label for="new-key-display-modal" class="btn" onClick={() => window.location.reload()}>Done</label>
                    </div>
                </div>
            </div>
        </>
    );
}
