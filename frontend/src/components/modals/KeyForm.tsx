import type { MiddlewareInfo, PluginConfig, ProviderInfo } from "../../types";
import MiddlewareComposition from "./MiddlewareComposition";
import ProviderConfigSection from "./ProviderConfigSection";

interface KeyFormData {
    name: string;
    providerId: string;
    providerConfig: Record<string, any>;
    autoRenew: boolean;
    middlewares: PluginConfig[];
    expiresAt: number | null;
    budgetLimit: string;
    resetPeriod: string;
}

interface Props {
    formData: KeyFormData;
    setFormData: (data: KeyFormData | ((prev: KeyFormData) => KeyFormData)) => void;
    middlewareInfos: MiddlewareInfo[];
    providerInfos: ProviderInfo[];
    showExpirationField?: boolean; // For Edit mode with datetime-local
    expirationDays?: string; // For Create mode with select
    setExpirationDays?: (days: string) => void;
}

export default function KeyForm({
    formData,
    setFormData,
    middlewareInfos,
    providerInfos,
    showExpirationField = false,
    expirationDays,
    setExpirationDays
}: Props) {
    const handleProviderChange = (newProviderId: string) => {
        const info = providerInfos.find(p => p.id === newProviderId);
        const defaults = info?.schema ? Object.keys(info.schema).reduce((acc, key) => {
            acc[key] = info.schema[key].default ?? "";
            return acc;
        }, {} as Record<string, any>) : {};

        setFormData(prev => ({
            ...prev,
            providerId: newProviderId,
            providerConfig: defaults,
        }));
    };

    const formatDateForInput = (timestamp: number | null) => {
        if (!timestamp) return "";
        const date = new Date(timestamp * 1000);
        return date.toISOString().slice(0, 16);
    };

    const handleExpiryChange = (val: string) => {
        if (setFormData) {
            setFormData(prev => ({
                ...prev,
                expiresAt: val === "" ? null : Math.floor(new Date(val).getTime() / 1000)
            }));
        }
    };

    return (
        <div class="space-y-6">
            {/* Basic Settings */}
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div class="form-control">
                    <label class="label pb-1"><span class="label-text text-xs text-white/50 font-medium">Name</span></label>
                    <input
                        type="text"
                        value={formData.name}
                        onInput={(e) => setFormData(prev => ({ ...prev, name: e.currentTarget.value }))}
                        placeholder="e.g. My App"
                        class="input input-bordered w-full bg-base-200/50 border-white/10 rounded-lg h-10"
                        required
                    />
                </div>
                <div class="form-control">
                    <label class="label pb-1"><span class="label-text text-xs text-white/50 font-medium">Provider</span></label>
                    <select
                        value={formData.providerId}
                        onChange={(e) => handleProviderChange(e.currentTarget.value)}
                        class="select select-bordered w-full bg-base-200/50 border-white/10 rounded-lg h-10"
                    >
                        {providerInfos.map(p => (
                            <option key={p.id} value={p.id}>{p.id.charAt(0).toUpperCase() + p.id.slice(1)}</option>
                        ))}
                    </select>
                </div>

                {!showExpirationField ? (
                    <div class="form-control">
                        <label class="label pb-1"><span class="label-text text-xs text-white/50 font-medium">Expiration</span></label>
                        <select
                            value={expirationDays}
                            onChange={(e) => setExpirationDays?.(e.currentTarget.value)}
                            class="select select-bordered w-full bg-base-200/50 border-white/10 rounded-lg h-10"
                        >
                            <option value="0">No Expiration</option>
                            <option value="7">1 Week</option>
                            <option value="30">1 Month</option>
                            <option value="90">3 Months</option>
                        </select>
                    </div>
                ) : (
                    <div class="form-control">
                        <label class="label pb-1"><span class="label-text text-xs text-white/50 font-medium">Expires At</span></label>
                        <input
                            type="datetime-local"
                            value={formatDateForInput(formData.expiresAt)}
                            onChange={(e) => handleExpiryChange(e.currentTarget.value)}
                            class="input input-bordered w-full bg-base-200/50 border-white/10 rounded-lg h-10 text-sm"
                        />
                    </div>
                )}

                <div class="form-control">
                    <label class="label pb-1"><span class="label-text text-xs text-white/50 font-medium">Budget Limit (USD)</span></label>
                    <input
                        type="number"
                        step="0.01"
                        value={formData.budgetLimit}
                        onInput={(e) => setFormData(prev => ({ ...prev, budgetLimit: e.currentTarget.value }))}
                        class="input input-bordered w-full bg-base-200/50 border-white/10 rounded-lg h-10"
                    />
                </div>
                <div class="form-control">
                    <label class="label pb-1 cursor-pointer flex justify-start gap-3">
                        <input
                            type="checkbox"
                            checked={formData.autoRenew}
                            onChange={(e) => setFormData(prev => ({ ...prev, autoRenew: e.currentTarget.checked }))}
                            class="checkbox checkbox-primary checkbox-sm rounded-md"
                        />
                        <span class="label-text text-sm font-medium text-white/70">Auto-Renew</span>
                    </label>
                    <div class="text-[10px] text-white/30 pl-8">Automatically reset budget and extend expiration</div>
                </div>
                <div class="form-control sm:col-span-2">
                    <label class="label pb-1"><span class="label-text text-xs text-white/50 font-medium">Reset Period (Seconds)</span></label>
                    <input
                        type="number"
                        value={formData.resetPeriod}
                        onInput={(e) => setFormData(prev => ({ ...prev, resetPeriod: e.currentTarget.value }))}
                        placeholder="2592000 = 30 days"
                        class="input input-bordered w-full bg-base-200/50 border-white/10 rounded-lg h-10"
                    />
                </div>
            </div>

            {/* Provider Config */}
            <ProviderConfigSection
                providerId={formData.providerId}
                providerInfos={providerInfos}
                config={formData.providerConfig}
                onConfigUpdate={(key, val) => setFormData(prev => ({
                    ...prev,
                    providerConfig: { ...prev.providerConfig, [key]: val }
                }))}
            />

            {/* Middlewares */}
            <MiddlewareComposition
                middlewares={formData.middlewares}
                middlewareInfos={middlewareInfos}
                setMiddlewares={(val) => {
                    const nextMws = typeof val === 'function' ? val(formData.middlewares) : val;
                    setFormData(prev => ({ ...prev, middlewares: nextMws }));
                }}
            />
        </div>
    );
}
