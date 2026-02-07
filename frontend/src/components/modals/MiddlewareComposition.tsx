import type { MiddlewareInfo, PluginConfig } from "../../types";
import SchemaField from "./SchemaField";

interface Props {
    middlewares: PluginConfig[];
    middlewareInfos: MiddlewareInfo[];
    setMiddlewares: (updater: (prev: PluginConfig[]) => PluginConfig[]) => void;
}

export default function MiddlewareComposition({ middlewares, middlewareInfos, setMiddlewares }: Props) {
    const toggleMiddleware = (mwId: string) => {
        setMiddlewares(prev => {
            const exists = prev.find(m => m.id === mwId);
            if (exists) return prev.filter(m => m.id !== mwId);
            const mw = middlewareInfos.find(m => m.id === mwId);
            if (!mw) return prev;
            return [...prev, {
                id: mwId,
                config: Object.keys(mw.schema).reduce((acc, key) => {
                    acc[key] = mw.schema[key].default !== undefined ? mw.schema[key].default : "";
                    return acc;
                }, {} as Record<string, any>)
            }];
        });
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

    const updateMiddlewareConfig = (mwId: string, key: string, value: string, type: string) => {
        setMiddlewares(prev => prev.map(m => {
            if (m.id !== mwId) return m;
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
        <div class="border-t border-white/5 pt-6 space-y-4">
            <div class="flex justify-between items-center">
                <label class="label p-0"><span class="label-text font-bold text-white/60 text-[10px] uppercase tracking-widest">Middleware Composition</span></label>
                <div class="dropdown dropdown-end">
                    <div tabindex={0} role="button" class="btn btn-xs btn-primary rounded-lg font-bold">Add Middleware</div>
                    <ul tabindex={0} class="dropdown-content z-[20] menu p-2 shadow-2xl bg-base-300 border border-white/10 rounded-xl w-52 mt-2">
                        {middlewareInfos.filter(mw => !middlewares.some(em => em.id === mw.id)).map(mw => (
                            <li key={mw.id}><a onClick={() => toggleMiddleware(mw.id)} class="text-xs font-bold text-white/60 hover:text-white">{mw.id.replace(/_/g, " ")}</a></li>
                        ))}
                    </ul>
                </div>
            </div>

            <div class="grid grid-cols-1 gap-3">
                {middlewares.map((emw, idx) => {
                    const mwInfo = middlewareInfos.find(m => m.id === emw.id);
                    if (!mwInfo) return null;
                    return (
                        <div class="p-4 rounded-xl bg-white/5 border border-white/5 space-y-4 relative group" key={emw.id}>
                            <div class="flex items-center justify-between">
                                <div class="flex items-center gap-3">
                                    <div class="flex flex-col gap-1">
                                        <button type="button" onClick={() => moveMiddleware(idx, 'up')} class={`btn btn-ghost btn-xs p-0 min-h-0 h-4 w-4 ${idx === 0 ? 'invisible' : ''}`}>▲</button>
                                        <button type="button" onClick={() => moveMiddleware(idx, 'down')} class={`btn btn-ghost btn-xs p-0 min-h-0 h-4 w-4 ${idx === middlewares.length - 1 ? 'invisible' : ''}`}>▼</button>
                                    </div>
                                    <span class="text-xs font-bold text-white/80">{emw.id.replace(/_/g, " ")}</span>
                                </div>
                                <button type="button" onClick={() => toggleMiddleware(emw.id)} class="btn btn-ghost btn-xs h-6 w-6 btn-circle text-white/20 group-hover:text-red-400">✕</button>
                            </div>

                            <div class="grid grid-cols-1 md:grid-cols-2 gap-3 pl-6 border-l border-white/10">
                                {Object.keys(mwInfo.schema).map(key => (
                                    <SchemaField
                                        key={key}
                                        id={key}
                                        schema={mwInfo.schema[key]}
                                        value={emw.config[key]}
                                        onUpdate={(val, type) => updateMiddlewareConfig(emw.id, key, val, type)}
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
