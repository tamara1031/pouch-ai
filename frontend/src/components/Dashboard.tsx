import { useState, useEffect } from "preact/hooks";
import type { Key, MiddlewareInfo } from "../types";
import KeyCard from "./KeyCard";

export default function Dashboard() {
    const [keys, setKeys] = useState<Key[]>([]);
    const [providerUsage, setProviderUsage] = useState<Record<string, number>>({});
    const [middlewareInfo, setMiddlewareInfo] = useState<MiddlewareInfo[]>([]);
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

    const loadProviderUsage = async () => {
        try {
            const res = await fetch("/v1/config/providers/usage", { cache: "no-store" });
            const data = await res.json();
            setProviderUsage(data || {});
        } catch (err) {
            console.error("Failed to load provider usage:", err);
        }
    };

    const loadMiddlewareInfo = async () => {
        try {
            const res = await fetch("/v1/config/middlewares", { cache: "no-store" });
            const data = await res.json();
            setMiddlewareInfo(data || []);
            (window as any).middlewareInfos = data || []; // Expose globally for KeyCardParts to use roles
        } catch (err) {
            console.error("Failed to load middleware info:", err);
        }
    };

    useEffect(() => {
        loadKeys();
        loadProviderUsage();
        loadMiddlewareInfo();
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
        const event = new CustomEvent('open-edit-modal', { detail: key });
        window.dispatchEvent(event);
    };

    // Helper to get budget limit from a key
    const getBudgetLimit = (k: Key) => k.configuration?.budget_limit || 0;

    const totalBudget = keys.reduce((acc, k) => acc + getBudgetLimit(k), 0);
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
            <div class="flex flex-col items-center justify-center py-20 px-6 bg-base-200/50 rounded-2xl border border-white/5 text-center shadow-lg">
                <div class="w-16 h-16 rounded-2xl bg-primary/10 flex items-center justify-center mb-6 mx-auto">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-primary/60" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                    </svg>
                </div>
                <h3 class="text-xl font-bold mb-2 text-white">No API Keys Found</h3>
                <p class="text-base-content/60 max-w-sm mx-auto mb-8 text-sm">
                    Generate your first API key to start using LLM providers with Pouch AI.
                </p>
                <label for="create-key-modal" class="btn btn-primary rounded-xl gap-2 shadow-lg shadow-primary/20 hover:scale-[1.02] transition-all">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    Generate API Key
                </label>
            </div>
        );
    }

    return (
        <div class="flex flex-col gap-10 animate-in fade-in slide-in-from-bottom-2 duration-700">
            {/* Stats Overview */}
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div class="bg-base-200/50 border border-white/5 p-6 rounded-2xl flex flex-col gap-1 transition-all hover:bg-base-200/80">
                    <span class="text-[10px] font-bold uppercase tracking-widest text-white/30 mb-1">Total Keys</span>
                    <div class="flex items-baseline gap-2">
                        <span class="text-3xl font-bold text-white tracking-tight">{keys.length}</span>
                        <span class="text-[10px] font-medium text-white/20 uppercase tracking-widest">Created</span>
                    </div>
                </div>
                <div class="bg-base-200/50 border border-white/5 p-6 rounded-2xl flex flex-col gap-1 transition-all hover:bg-base-200/80">
                    <span class="text-[10px] font-bold uppercase tracking-widest text-white/30 mb-1">Active Status</span>
                    <div class="flex items-baseline gap-2">
                        <span class="text-3xl font-bold text-success tracking-tight">{activeKeys}</span>
                        <span class="text-[10px] font-medium text-white/20 uppercase tracking-widest">Online</span>
                    </div>
                </div>
                <div class="bg-base-200/50 border border-white/5 p-6 rounded-2xl flex flex-col gap-1 transition-all hover:bg-base-200/80">
                    <span class="text-[10px] font-bold uppercase tracking-widest text-white/30 mb-1">Usage Cost</span>
                    <div class="flex items-baseline gap-2">
                        <span class="text-3xl font-bold text-primary tracking-tight">${totalUsage.toFixed(2)}</span>
                        <span class="text-[10px] font-medium text-white/20 font-mono">/ {totalBudget > 0 ? "$" + totalBudget.toFixed(0) : "∞"}</span>
                    </div>
                    {providerUsage["openai"] !== undefined && (
                        <div class="mt-2 pt-2 border-t border-white/5 flex items-center justify-between">
                            <span class="text-[9px] font-bold uppercase tracking-wider text-white/20">OpenAI Actual</span>
                            <span class="text-[10px] font-bold text-white/40">${providerUsage["openai"].toFixed(2)}</span>
                        </div>
                    )}
                </div>
            </div>

            <div class="flex flex-col gap-4">
                <div class="flex items-center justify-between px-2">
                    <h2 class="text-[10px] font-bold uppercase tracking-widest text-white/20">Active API Keys</h2>
                    <div class="h-px flex-1 bg-white/5 ml-4"></div>
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
