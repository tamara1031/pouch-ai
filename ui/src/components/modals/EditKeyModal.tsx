import { useState, useEffect } from "preact/hooks";
import type { Key } from "../../types";

interface Props {
    modalRef: any;
    editKey: Key | null;
}

export default function EditKeyModal({ modalRef, editKey }: Props) {
    const [editProvider, setEditProvider] = useState("openai");

    useEffect(() => {
        if (editKey) {
            setEditProvider(editKey.provider);
        }
    }, [editKey]);

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
                if (modalRef.current) modalRef.current.checked = false;
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
            <input type="checkbox" id="edit-key-modal" class="modal-toggle" ref={modalRef} />
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
        </>
    );
}
