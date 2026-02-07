import type { Key } from "../types";
import { getKeyStatus, StatusBadge, CopyButton, UsageBar } from "./KeyCardParts";

interface Props {
    keyData: Key;
    onEdit: (key: Key) => void;
    onRevoke: (id: number) => void;
}

export default function KeyCard({ keyData, onEdit, onRevoke }: Props) {
    const {
        id,
        name,
        prefix,
        budget_usage,
        configuration,
    } = keyData;

    const status = getKeyStatus(keyData);
    const { expiresText, usagePercent, isMock, budgetLimit } = status;

    // Extract Rate Limit from any middleware with both "limit" and "period" roles
    let rate_limit = 0;
    let rate_period: any = "none";

    for (const emw of configuration?.middlewares || []) {
        const info = (window as any).middlewareInfo?.find((info: any) => info.id === emw.id);
        if (!info) continue;

        const limitKey = Object.keys(info.schema).find(k => info.schema[k].role === "limit");
        const periodKey = Object.keys(info.schema).find(k => info.schema[k].role === "period");

        if (limitKey && periodKey) {
            rate_limit = parseFloat(emw.config[limitKey] || "0");
            rate_period = emw.config[periodKey];
            break; // Use the first one found
        }
    }

    // Rate Limit Text
    let rateLimitText = "Unlimited";
    if (rate_limit > 0) {
        if (typeof rate_period === "number" && rate_period > 0) {
            if (rate_period === 1) rateLimitText = `${rate_limit}/sec`;
            else if (rate_period === 60) rateLimitText = `${rate_limit}/min`;
            else if (rate_period === 3600) rateLimitText = `${rate_limit}/hr`;
            else rateLimitText = `${rate_limit}/${rate_period}s`;
        } else if (typeof rate_period === "string" && rate_period !== "none" && rate_period !== "") {
            rateLimitText = `${rate_limit}/${rate_period === "second" ? "sec" : "min"}`;
        }
    }

    return (
        <div class="group relative overflow-hidden bg-base-200/50 border border-white/5 rounded-2xl transition-all hover:bg-base-200/80">
            <div class="p-6 relative">
                <div class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-6">
                    <div class="flex-1 space-y-3">
                        <div class="flex flex-wrap items-center gap-3">
                            <h2 class="text-xl font-bold text-white tracking-tight">{name}</h2>
                            <div class="px-2 py-0.5 rounded bg-white/5 text-[9px] font-bold uppercase text-white/40 tracking-wider border border-white/5">{configuration?.provider.id || "openai"}</div>
                            <StatusBadge status={status} isMock={isMock} />
                        </div>
                        <div class="flex items-center gap-2">
                            <CopyButton text={prefix} />
                        </div>
                    </div>

                    <div class="w-full lg:w-auto grid grid-cols-2 md:grid-cols-4 lg:flex lg:items-center gap-4 sm:gap-8">
                        <div class="space-y-1">
                            <span class="text-[9px] font-bold uppercase tracking-wider text-white/20">Usage</span>
                            <div class="flex items-baseline gap-1">
                                <span class="text-lg font-bold text-white tracking-tight">${budget_usage.toFixed(2)}</span>
                                <span class="text-[10px] font-medium text-white/20">/ {budgetLimit > 0 ? "$" + budgetLimit.toFixed(0) : "∞"}</span>
                            </div>
                            <UsageBar percent={usagePercent} />
                        </div>

                        <div class="space-y-1 min-w-0">
                            <span class="text-[9px] font-bold uppercase tracking-wider text-white/20">Rate</span>
                            <div class="text-lg font-bold text-white tracking-tight font-mono truncate">{rateLimitText}</div>
                        </div>

                        <div class="space-y-1 hidden md:block">
                            <span class="text-[9px] font-bold uppercase tracking-wider text-white/20">Expiry</span>
                            <div class={`text-sm font-bold tracking-tight ${status.isExpired ? 'text-error' : 'text-white/60'}`}>{expiresText}</div>
                        </div>

                        <div class="flex flex-row lg:flex-row gap-2 justify-end items-center flex-1 sm:flex-none">
                            <button
                                class="btn btn-sm h-9 px-4 rounded-lg bg-white/5 border-none text-[10px] font-bold uppercase tracking-wider text-white/40 hover:text-white hover:bg-white/10 transition-all"
                                onClick={() => onEdit(keyData)}
                            >
                                Edit
                            </button>
                            <button
                                class="btn btn-sm h-9 px-4 rounded-lg bg-error/5 hover:bg-error/10 border-none text-[10px] font-bold uppercase tracking-wider text-error/40 hover:text-error transition-all"
                                onClick={() => onRevoke(id)}
                            >
                                Revoke
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
