import { useState } from "preact/hooks";
import type { Key } from "../types";

interface Props {
    keyData: Key;
    onEdit: (key: Key) => void;
    onRevoke: (id: number) => void;
}

export default function KeyCard({ keyData, onEdit, onRevoke }: Props) {
    const [copied, setCopied] = useState(false);
    const {
        id,
        name,
        provider,
        prefix,
        expires_at,
        budget_limit,
        budget_usage,
        budget_period,
        is_mock,
        rate_limit,
        rate_period,
        created_at
    } = keyData;

    const handleCopy = () => {
        navigator.clipboard.writeText(prefix);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    // Expiration Logic
    let expiresText = "Never";
    let isExpired = false;
    if (expires_at) {
        const expDate = new Date(expires_at * 1000);
        expiresText = expDate.toLocaleDateString();
        if (expDate < new Date()) {
            isExpired = true;
        }
    }

    // Budget Logic
    const usagePercent = budget_limit > 0 ? Math.min((budget_usage / budget_limit) * 100, 100) : 0;
    let barColor = "from-primary to-secondary";
    if (usagePercent > 80) barColor = "from-warning to-error";
    if (usagePercent >= 100) barColor = "from-error to-error";

    const isDepleted = budget_limit > 0 && budget_usage >= budget_limit;

    // Status Badge
    let statusBadge = (
        <span class="flex items-center gap-2 px-3 py-1 rounded-full bg-success/10 border border-success/20 text-[10px] font-black uppercase tracking-widest text-success shadow-[0_0_15px_rgba(var(--s-rgb),0.1)]">
            <div class="w-1.5 h-1.5 rounded-full bg-success animate-pulse shadow-[0_0_8px_rgba(var(--s-rgb),0.8)]"></div>
            Online
        </span>
    );

    if (isExpired) {
        statusBadge = (
            <span class="flex items-center gap-2 px-3 py-1 rounded-full bg-error/10 border border-error/20 text-[10px] font-black uppercase tracking-widest text-error">
                <div class="w-1.5 h-1.5 rounded-full bg-error"></div>
                Expired
            </span>
        );
    } else if (isDepleted && !is_mock) {
        statusBadge = (
            <span class="flex items-center gap-2 px-3 py-1 rounded-full bg-warning/10 border border-warning/20 text-[10px] font-black uppercase tracking-widest text-warning">
                <div class="w-1.5 h-1.5 rounded-full bg-warning"></div>
                Capped
            </span>
        );
    } else if (is_mock) {
        statusBadge = (
            <span class="flex items-center gap-2 px-3 py-1 rounded-full bg-info/10 border border-info/20 text-[10px] font-black uppercase tracking-widest text-info">
                <div class="w-1.5 h-1.5 rounded-full bg-info animate-bounce"></div>
                Simulation
            </span>
        );
    }

    // Mode Badge
    const modeBadge = (budget_period && budget_period !== "none") ? (
        <span class="text-[9px] font-black uppercase tracking-[0.2em] text-white/30 border border-white/5 rounded-lg px-2 py-1 bg-white/[0.02]">Recurrent</span>
    ) : (
        <span class="text-[9px] font-black uppercase tracking-[0.2em] text-white/30 border border-white/5 rounded-lg px-2 py-1 bg-white/[0.02]">Disposable</span>
    );

    // Rate Limit Text
    const rateLimitText = (rate_period !== "none" && rate_limit > 0)
        ? `${rate_limit}/${rate_period === "second" ? "sec" : "min"}`
        : "Unlimited";

    return (
        <div class="group relative overflow-hidden bg-base-200/50 border border-white/5 rounded-2xl transition-all hover:bg-base-200/80">
            <div class="p-6 relative">
                <div class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-6">
                    <div class="flex-1 space-y-3">
                        <div class="flex flex-wrap items-center gap-3">
                            <h2 class="text-xl font-bold text-white tracking-tight">{name}</h2>
                            <div class="px-2 py-0.5 rounded bg-white/5 text-[9px] font-bold uppercase text-white/40 tracking-wider border border-white/5">{provider || "openai"}</div>
                            {statusBadge}
                        </div>
                        <div class="flex items-center gap-2">
                            <div class="relative group/copy">
                                <code class="text-xs font-mono text-white/40 bg-black/20 px-3 py-1.5 rounded-lg border border-white/5 block group-hover/copy:text-white/60 transition-colors">
                                    {prefix}••••••••
                                </code>
                                <button
                                    onClick={handleCopy}
                                    class="absolute right-1.5 top-1/2 -translate-y-1/2 p-1.5 rounded bg-white/5 hover:bg-white/10 opacity-0 group-hover/copy:opacity-100 transition-all"
                                    title="Copy prefix"
                                >
                                    {copied ? (
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" /></svg>
                                    ) : (
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 text-white/40" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" /></svg>
                                    )}
                                </button>
                                {copied && (
                                    <div class="absolute -top-8 left-1/2 -translate-x-1/2 px-2 py-0.5 bg-success text-white text-[9px] font-bold rounded animate-in fade-in zoom-in slide-in-from-bottom-1 duration-200">COPIED</div>
                                )}
                            </div>
                        </div>
                    </div>

                    <div class="w-full lg:w-auto grid grid-cols-2 md:grid-cols-4 lg:flex lg:items-center gap-8">
                        <div class="space-y-1">
                            <span class="text-[9px] font-bold uppercase tracking-wider text-white/20">Usage</span>
                            <div class="flex items-baseline gap-1">
                                <span class="text-lg font-bold text-white tracking-tight">${budget_usage.toFixed(2)}</span>
                                <span class="text-[10px] font-medium text-white/20">/ {budget_limit > 0 ? "$" + budget_limit.toFixed(0) : "∞"}</span>
                            </div>
                            <div class="w-20 h-1 bg-white/5 rounded-full overflow-hidden mt-1.5">
                                <div class={`h-full rounded-full bg-gradient-to-r ${barColor} transition-all duration-1000`} style={{ width: `${usagePercent}%` }}></div>
                            </div>
                        </div>

                        <div class="space-y-1">
                            <span class="text-[9px] font-bold uppercase tracking-wider text-white/20">Rate</span>
                            <div class="text-lg font-bold text-white tracking-tight font-mono">{rateLimitText}</div>
                        </div>

                        <div class="space-y-1 hidden md:block">
                            <span class="text-[9px] font-bold uppercase tracking-wider text-white/20">Expiry</span>
                            <div class={`text-sm font-bold tracking-tight ${isExpired ? 'text-error' : 'text-white/60'}`}>{expiresText}</div>
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
