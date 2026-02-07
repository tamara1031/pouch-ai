import { useEffect } from "preact/hooks";
import { pluginStore, keyStore } from "../../../lib/store";
import { apiClient } from "../../../lib/api-client";
import KeyList from "./KeyList";
import CreateKeyModal from "./modals/CreateKeyModal";
import EditKeyModal from "./modals/EditKeyModal";
import NewKeyDisplayModal from "./modals/NewKeyDisplayModal";

export default function KeyManagement() {
    const keys = keyStore.keys.value;
    const activeKeys = keys.filter(k => !k.expires_at || new Date(k.expires_at * 1000) > new Date()).length;

    useEffect(() => {
        async function loadInfo() {
            try {
                const [mwData, pData] = await Promise.all([
                    apiClient.plugins.middlewares(),
                    apiClient.plugins.providers()
                ]);
                pluginStore.middlewares.value = mwData.middlewares || [];
                pluginStore.providers.value = pData.providers || [];
            } catch (err) {
                console.error(err);
            }
        }
        loadInfo();
    }, []);

    return (
        <div class="flex flex-col gap-10 animate-in fade-in slide-in-from-bottom-2 duration-700">
            {/* Stats Overview */}
            {keys.length > 0 && (
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
            )}

            <KeyList />

            <CreateKeyModal />
            <EditKeyModal />
            <NewKeyDisplayModal />
        </div>
    );
}
