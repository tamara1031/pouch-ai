import type { ProviderInfo } from "../../../../types";
import SchemaField from "./SchemaField";

interface Props {
    providerId: string;
    providerInfos: ProviderInfo[];
    config: Record<string, any>;
    onConfigUpdate: (key: string, value: any) => void;
}

export default function ProviderConfigSection({ providerId, providerInfos, config, onConfigUpdate }: Props) {
    const providerInfo = providerInfos.find(p => p.id === providerId);

    if (!providerInfo || !providerInfo.schema || Object.keys(providerInfo.schema).length === 0) {
        return null;
    }

    const handleUpdate = (key: string, value: string, type: string) => {
        let val: any = value;
        if (type === "number") {
            val = value === "" ? 0 : parseFloat(value);
        } else if (type === "boolean") {
            val = value === "true";
        }
        onConfigUpdate(key, val);
    };

    const providerName = providerId.charAt(0).toUpperCase() + providerId.slice(1);

    return (
        <div class="space-y-4">
            <div class="flex items-center gap-2">
                <div class="w-2 h-2 rounded-full bg-primary"></div>
                <span class="text-sm font-semibold text-white/70">
                    {providerName} Settings
                </span>
            </div>
            <div class="p-5 rounded-xl bg-base-200/30 border border-white/5 space-y-4">
                {Object.keys(providerInfo.schema).map(key => (
                    <SchemaField
                        key={key}
                        id={key}
                        schema={providerInfo.schema[key]}
                        value={config[key] ?? providerInfo.schema[key].default ?? ""}
                        onUpdate={(val, type) => handleUpdate(key, val, type)}
                    />
                ))}
            </div>
        </div>
    );
}
