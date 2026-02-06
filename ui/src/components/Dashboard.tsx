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
                <span class="loading loading-spinner loading-lg text-primary"></span>
                <p class="text-base-content/50 mt-4 text-sm animate-pulse">Initializing Dashboard...</p>
            </div>
        );
    }

    if (keys.length === 0) {
        return (
            <div class="flex flex-col items-center justify-center py-20 px-6 bg-white/5 backdrop-blur-md rounded-[2rem] border border-white/5 text-center shadow-2xl">
                <div class="w-20 h-20 rounded-3xl bg-gradient-to-br from-primary/20 to-secondary/20 flex items-center justify-center mb-8 mx-auto shadow-inner">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                    </svg>
                </div>
                <h3 class="text-2xl font-bold mb-3 text-white">Secure Your AI Fleet</h3>
                <p class="text-base-content/60 max-w-sm mx-auto mb-10 text-base leading-relaxed">
                    Create your first programmable API key to impose budgets, rate limits, and monitoring on your LLM usage.
                </p>
                <label for="create-key-modal" class="btn btn-primary btn-lg rounded-2xl gap-3 shadow-lg shadow-primary/25 hover:scale-105 transition-transform">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    Provision First Key
                </label>
            </div>
        );
    }

    return (
        <div class="flex flex-col gap-10 animate-in fade-in slide-in-from-bottom-4 duration-1000">
            {/* Stats Overview */}
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div class="group relative overflow-hidden bg-white/5 backdrop-blur-2xl border border-white/5 p-8 rounded-[2rem] flex flex-col gap-1 transition-all duration-500 hover:bg-white/[0.08] hover:border-white/10 hover:shadow-2xl hover:shadow-primary/10">
                    <div class="absolute top-0 right-0 w-32 h-32 bg-primary/10 rounded-full blur-3xl -mr-16 -mt-16 group-hover:bg-primary/20 transition-all duration-700"></div>
                    <span class="text-[10px] font-black uppercase tracking-[0.2em] text-white/30 mb-2">Registry Ledger</span>
                    <div class="flex items-baseline gap-3">
                        <span class="text-4xl font-black text-white tracking-tighter">{keys.length}</span>
                        <span class="text-[10px] font-bold text-white/20 uppercase tracking-widest">Provisioned</span>
                    </div>
                </div>
                <div class="group relative overflow-hidden bg-white/5 backdrop-blur-2xl border border-white/5 p-8 rounded-[2rem] flex flex-col gap-1 transition-all duration-500 hover:bg-white/[0.08] hover:border-white/10 hover:shadow-2xl hover:shadow-success/10">
                    <div class="absolute top-0 right-0 w-32 h-32 bg-success/10 rounded-full blur-3xl -mr-16 -mt-16 group-hover:bg-success/20 transition-all duration-700"></div>
                    <span class="text-[10px] font-black uppercase tracking-[0.2em] text-white/30 mb-2">Fleet Health</span>
                    <div class="flex items-baseline gap-3">
                        <span class="text-4xl font-black text-success tracking-tighter">{activeKeys}</span>
                        <span class="text-[10px] font-bold text-white/20 uppercase tracking-widest">Active</span>
                    </div>
                </div>
                <div class="group relative overflow-hidden bg-white/5 backdrop-blur-2xl border border-white/5 p-8 rounded-[2rem] flex flex-col gap-1 transition-all duration-500 hover:bg-white/[0.08] hover:border-white/10 hover:shadow-2xl hover:shadow-secondary/10">
                    <div class="absolute top-0 right-0 w-32 h-32 bg-secondary/10 rounded-full blur-3xl -mr-16 -mt-16 group-hover:bg-secondary/20 transition-all duration-700"></div>
                    <span class="text-[10px] font-black uppercase tracking-[0.2em] text-white/30 mb-2">Monetary Velocity</span>
                    <div class="flex items-baseline gap-3">
                        <span class="text-4xl font-black text-primary tracking-tighter">${totalUsage.toFixed(2)}</span>
                        <span class="text-[10px] font-bold text-white/20 uppercase tracking-widest font-mono">/ {totalBudget > 0 ? "$" + totalBudget.toFixed(0) : "∞"} Cap</span>
                    </div>
                </div>
            </div>

            <div class="flex flex-col gap-5">
                <div class="flex items-center justify-between px-2">
                    <h2 class="text-[10px] font-black uppercase tracking-[0.3em] text-white/20">Managed Access Points</h2>
                    <div class="h-px flex-1 bg-white/5 ml-6"></div>
                </div>
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
    );
}
