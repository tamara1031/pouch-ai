import type { MiddlewareInfo, PluginConfig } from "../../types";
import SchemaField from "./SchemaField";

interface Props {
    middlewares: PluginConfig[];
    middlewareInfos: MiddlewareInfo[];
    setMiddlewares: (updater: (prev: PluginConfig[]) => PluginConfig[]) => void;
}

export default function MiddlewareComposition({ middlewares, middlewareInfos, setMiddlewares }: Props) {
    const addMiddleware = (mwId: string) => {
        const mw = middlewareInfos.find(m => m.id === mwId);
        if (!mw) return;
        setMiddlewares(prev => [...prev, {
            id: mwId,
            config: Object.keys(mw.schema).reduce((acc, key) => {
                acc[key] = mw.schema[key].default !== undefined ? mw.schema[key].default : "";
                return acc;
            }, {} as Record<string, any>)
        }]);
    };

    const removeMiddleware = (index: number) => {
        setMiddlewares(prev => prev.filter((_, i) => i !== index));
    };

    const moveMiddleware = (index: number, direction: 'up' | 'down') => {
        setMiddlewares(prev => {
            const next = [...prev];
            const newIndex = direction === 'up' ? index - 1 : index + 1;
            if (newIndex < 0 || newIndex >= next.length) return prev;
            [next[index], next[newIndex]] = [next[newIndex], next[index]];
            return next;
        });
    };

    const updateMiddlewareConfig = (index: number, key: string, value: string, type: string) => {
        setMiddlewares(prev => prev.map((m, i) => {
            if (i !== index) return m;
            let val: any = value;
            if (type === "number") {
                val = value === "" ? 0 : parseFloat(value);
            } else if (type === "boolean") {
                val = value === "true";
            }
            return { ...m, config: { ...m.config, [key]: val } };
        }));
    };

    return (
        <div class="space-y-4">
            <div class="flex justify-between items-center">
                <div class="flex items-center gap-2">
                    <div class="w-2 h-2 rounded-full bg-secondary"></div>
                    <span class="text-sm font-semibold text-white/70">Middlewares</span>
                    <span class="text-xs text-white/30">({middlewares.length})</span>
                </div>
                <div class="dropdown dropdown-end">
                    <div tabindex={0} role="button" class="btn btn-sm btn-primary rounded-lg">+ Add</div>
                    <ul tabindex={0} class="dropdown-content z-[20] menu p-2 shadow-2xl bg-base-300 border border-white/10 rounded-xl w-52 mt-2">
                        {middlewareInfos.map(mw => (
                            <li key={mw.id}><a onClick={() => addMiddleware(mw.id)} class="text-sm text-white/60 hover:text-white capitalize">{mw.id.replace(/_/g, " ")}</a></li>
                        ))}
                    </ul>
                </div>
            </div>

            <div class="space-y-3">
                {middlewares.map((emw, idx) => {
                    const mwInfo = middlewareInfos.find(m => m.id === emw.id);
                    if (!mwInfo) return null;
                    return (
                        <div class="p-4 rounded-xl bg-base-200/30 border border-white/5 space-y-4 relative group" key={`mw-${idx}`}>
                            <div class="flex items-center justify-between">
                                <div class="flex items-center gap-3">
                                    <div class="flex flex-col gap-0.5">
                                        <button type="button" onClick={() => moveMiddleware(idx, 'up')} class={`btn btn-ghost btn-xs p-0 min-h-0 h-5 w-5 text-white/30 hover:text-white ${idx === 0 ? 'invisible' : ''}`}>▲</button>
                                        <button type="button" onClick={() => moveMiddleware(idx, 'down')} class={`btn btn-ghost btn-xs p-0 min-h-0 h-5 w-5 text-white/30 hover:text-white ${idx === middlewares.length - 1 ? 'invisible' : ''}`}>▼</button>
                                    </div>
                                    <span class="text-sm font-medium text-white/80 capitalize">{emw.id.replace(/_/g, " ")}</span>
                                </div>
                                <button type="button" onClick={() => removeMiddleware(idx)} class="btn btn-ghost btn-sm h-7 w-7 btn-circle text-white/20 hover:text-red-400 hover:bg-red-400/10">✕</button>
                            </div>

                            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 pl-5 border-l-2 border-white/10">
                                {Object.keys(mwInfo.schema).map(key => (
                                    <SchemaField
                                        key={`${idx}-${key}`}
                                        id={key}
                                        schema={mwInfo.schema[key]}
                                        value={emw.config[key]}
                                        onUpdate={(val, type) => updateMiddlewareConfig(idx, key, val, type)}
                                    />
                                ))}
                            </div>
                        </div>
                    )
                })}
            </div>
        </div>
    );
}
