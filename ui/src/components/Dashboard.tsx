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

    if (loading && keys.length === 0) {
        return (
            <div class="flex flex-col items-center justify-center py-24">
                <span class="loading loading-spinner loading-lg text-primary"></span>
                <p class="text-base-content/50 mt-4 text-sm">Loading keys...</p>
            </div>
        );
    }

    if (keys.length === 0) {
        return (
            <div class="flex flex-col items-center justify-center py-20 bg-base-100/30 backdrop-blur-sm rounded-2xl border border-base-content/5 text-center">
                <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-primary/20 to-secondary/20 flex items-center justify-center mb-6 mx-auto">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                    </svg>
                </div>
                <h3 class="text-xl font-semibold mb-2 text-base-content">No API Keys Yet</h3>
                <p class="text-base-content/50 max-w-sm mx-auto mb-8 text-sm">
                    Create your first API key to start managing access and budgets for your AI applications.
                </p>
                <label for="create-key-modal" class="btn btn-primary btn-sm gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    Create Your First Key
                </label>
            </div>
        );
    }

    return (
        <div class="flex flex-col gap-4">
            {keys.map((key) => (
                <KeyCard
                    key={key.id}
                    keyData={key}
                    onEdit={handleEdit}
                    onRevoke={handleRevoke}
                />
            ))}
        </div>
    );
}
