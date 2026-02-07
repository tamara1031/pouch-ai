import { useState, useEffect } from "preact/hooks";
import type { Key } from "../../../types";
import { api } from "../../../api/api";
import { usePluginInfo } from "../../../hooks/usePluginInfo";
import KeyCard from "./KeyCard";
import { openEditModal, openCreateModal } from "../../../hooks/useModalState";

export default function Dashboard() {
    const [keys, setKeys] = useState<Key[]>([]);
    const { middlewareInfo, loading: infoLoading } = usePluginInfo();
    const [loading, setLoading] = useState(true);

    const loadKeys = async () => {
        setLoading(true);
        try {
            const data = await api.keys.list();
            setKeys(data || []);
        } catch (err) {
            console.error("Failed to load keys:", err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadKeys();

        // Listen for refresh events if any
        const handleRefresh = () => loadKeys();
        window.addEventListener('refresh-keys', handleRefresh);
        return () => window.removeEventListener('refresh-keys', handleRefresh);
    }, []);

    const handleRevoke = async (id: number) => {
        if (!confirm("Are you sure? This cannot be undone.")) return;
        try {
            await api.keys.delete(id);
            loadKeys();
        } catch (err) {
            console.error("Revoke error:", err);
            alert("Failed to revoke key");
        }
    };

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
                <button
                    onClick={openCreateModal}
                    class="btn btn-primary rounded-xl gap-2 shadow-lg shadow-primary/20 hover:scale-[1.02] transition-all"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    Generate API Key
                </button>
            </div>
        );
    }

    return (
        <div class="flex flex-col gap-10 animate-in fade-in slide-in-from-bottom-2 duration-700">
            {/* Stats Overview */}
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
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
                        middlewareInfos={middlewareInfo}
                        onEdit={openEditModal}
                        onRevoke={handleRevoke}
                    />
                ))}
            </div>
        </div>
    );
}
