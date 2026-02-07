import type { ProviderInfo } from "../../types";
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

    return (
        <div class="border-t border-white/5 pt-6 space-y-4">
            <label class="label p-0">
                <span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">
                    {providerId.replace(/_/g, " ")} Configuration
                </span>
            </label>
            <div class="p-4 rounded-xl bg-white/5 border border-white/5 space-y-4">
                <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
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
        </div>
    );
}
