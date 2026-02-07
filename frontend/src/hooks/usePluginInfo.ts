import { useState, useEffect } from "preact/hooks";
import { api } from "../api/api";
import type { MiddlewareInfo, ProviderInfo } from "../types";

export function usePluginInfo() {
    const [middlewareInfo, setMiddlewareInfo] = useState<MiddlewareInfo[]>([]);
    const [providerInfo, setProviderInfo] = useState<ProviderInfo[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<Error | null>(null);

    useEffect(() => {
        async function loadInfo() {
            try {
                const [mwData, pData] = await Promise.all([
                    api.plugins.middlewares(),
                    api.plugins.providers()
                ]);
                setMiddlewareInfo(mwData.middlewares || []);
                setProviderInfo(pData.providers || []);
            } catch (err) {
                console.error("Failed to load plugin info:", err);
                setError(err as Error);
            } finally {
                setLoading(false);
            }
        }
        loadInfo();
    }, []);

    return { middlewareInfo, providerInfo, loading, error };
}
