import { useEffect } from "preact/hooks";
import { keyStore, modalStore } from "../../../lib/store";
import { apiClient } from "../../../lib/api-client";
import KeyItem from "./KeyItem";

export default function KeyList() {
    const keys = keyStore.keys.value;
    const loading = keyStore.loading.value;

    const fetchKeys = async () => {
        keyStore.loading.value = true;
        try {
            const data = await apiClient.keys.list();
            keyStore.keys.value = data || [];
        } catch (err) {
            console.error("Failed to load keys:", err);
            keyStore.error.value = err as Error;
        } finally {
            keyStore.loading.value = false;
        }
    };

    useEffect(() => {
        fetchKeys();
        const handleRefresh = () => fetchKeys();
        window.addEventListener('refresh-keys', handleRefresh);
        return () => window.removeEventListener('refresh-keys', handleRefresh);
    }, []);

    const handleRevoke = async (id: number) => {
        if (!confirm("Are you sure? This cannot be undone.")) return;
        try {
            await apiClient.keys.delete(id);
            fetchKeys();
        } catch (err) {
            console.error("Revoke error:", err);
            alert("Failed to revoke key");
        }
    };

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
                    onClick={() => modalStore.openCreate()}
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
        <div class="flex flex-col gap-4">
             <div class="flex items-center justify-between px-2">
                <h2 class="text-[10px] font-bold uppercase tracking-widest text-white/20">Active API Keys</h2>
                <div class="h-px flex-1 bg-white/5 ml-4"></div>
            </div>
            {keys.map(key => <KeyItem key={key.id} keyData={key} onRevoke={handleRevoke} />)}
        </div>
    );
}
