import { useState, useEffect } from "preact/hooks";
import type { Key } from "../types";
import KeyCard from "./KeyCard";

export default function Dashboard() {
    const [keys, setKeys] = useState<Key[]>([]);
    const [loading, setLoading] = useState(true);

    const loadKeys = async () => {
        setLoading(true);
        try {
            const res = await fetch("/v1/config/app-keys", { cache: "no-store" });
            const data = await res.json();
            setKeys(data || []);
        } catch (err) {
            console.error("Failed to load keys:", err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadKeys();
    }, []);

    const handleRevoke = async (id: number) => {
        if (!confirm("Are you sure? This cannot be undone.")) return;
        try {
            const res = await fetch(`/v1/config/app-keys/${id}`, { method: "DELETE" });
            if (res.ok) {
                loadKeys();
            } else {
                alert("Failed to revoke key");
            }
        } catch (err) {
            console.error("Revoke error:", err);
        }
    };

    const handleEdit = (key: Key) => {
        // We'll implement Edit Modal coordination here
        console.log("Edit key:", key);
        // This will trigger the Edit modal state
        const event = new CustomEvent('open-edit-modal', { detail: key });
        window.dispatchEvent(event);
    };

    const totalBudget = keys.reduce((acc, k) => acc + (k.budget_limit > 0 ? k.budget_limit : 0), 0);
    const totalUsage = keys.reduce((acc, k) => acc + k.budget_usage, 0);
    const activeKeys = keys.filter(k => !k.expires_at || new Date(k.expires_at * 1000) > new Date()).length;

    if (loading && keys.length === 0) {
        return (
            <div class="flex flex-col items-center justify-center py-24">
                <span class="loading loading-spinner loading-lg text-primary opacity-20"></span>
                <p class="text-white/20 mt-6 text-xs font-bold uppercase tracking-[0.2em] animate-pulse">Initializing Dashboard...</p>
            </div>
        );
    }

    if (keys.length === 0) {
        return (
            <div class="flex flex-col items-center justify-center py-24 px-12 bg-white/[0.02] border border-white/[0.05] rounded-[2.5rem] text-center">
                <div class="w-16 h-16 rounded-2xl bg-white/5 flex items-center justify-center mb-8 mx-auto">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white/20" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                    </svg>
                </div>
                <h3 class="text-2xl font-bold mb-4 text-white">Generate Your First Key</h3>
                <p class="text-white/30 max-w-sm mx-auto mb-10 text-sm leading-relaxed">
                    Start by generating a secure API key to manage quotas and monitor usage for your applications.
                </p>
                <label for="create-key-modal" class="btn btn-primary px-8 rounded-xl font-bold uppercase tracking-widest text-[10px] h-12 shadow-xl shadow-primary/10">
                    Generate Key
                </label>
            </div>
        );
    }

    return (
        <div class="flex flex-col gap-20 animate-in fade-in slide-in-from-bottom-4 duration-1000">
            {/* Stats Overview - Simple & Informative */}
            <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
                <div class="group relative bg-white/[0.02] border border-white/[0.05] p-10 rounded-[2.5rem] transition-all hover:bg-white/[0.04]">
                    <span class="text-[10px] font-bold uppercase tracking-[0.2em] text-white/20 mb-4 block">Total Keys</span>
                    <div class="flex items-baseline gap-3">
                        <span class="text-5xl font-black text-white tracking-tighter">{keys.length}</span>
                    </div>
                </div>
                <div class="group relative bg-white/[0.02] border border-white/[0.05] p-10 rounded-[2.5rem] transition-all hover:bg-white/[0.04]">
                    <span class="text-[10px] font-bold uppercase tracking-[0.2em] text-white/20 mb-4 block">Active Keys</span>
                    <div class="flex items-baseline gap-3">
                        <span class="text-5xl font-black text-white tracking-tighter">{activeKeys}</span>
                    </div>
                </div>
                <div class="group relative bg-white/[0.02] border border-white/[0.05] p-10 rounded-[2.5rem] transition-all hover:bg-white/[0.04]">
                    <span class="text-[10px] font-bold uppercase tracking-[0.2em] text-white/20 mb-4 block">Total Usage</span>
                    <div class="flex items-baseline gap-3">
                        <span class="text-5xl font-black text-white tracking-tighter">${totalUsage.toFixed(2)}</span>
                        <span class="text-[10px] font-bold text-white/20 uppercase tracking-widest">/ {totalBudget > 0 ? "$" + totalBudget.toFixed(0) : "∞"}</span>
                    </div>
                </div>
            </div>

            <div class="flex flex-col gap-6">
                <div class="flex items-center justify-between px-2 mb-4">
                    <h2 class="text-[10px] font-bold uppercase tracking-[0.3em] text-white/20">Generated Keys</h2>
                    <div class="h-px flex-1 bg-white/[0.05] ml-8"></div>
                </div>
                <div class="grid grid-cols-1 gap-4">
                    {keys.map((key) => (
                        <KeyCard
                            key={key.id}
                            keyData={key}
                            onEdit={handleEdit}
                            onRevoke={handleRevoke}
                        />
                    ))}
                </div>
            </div>
        </div>
    );
}
