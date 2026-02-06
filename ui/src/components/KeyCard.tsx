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
        <span class="badge badge-success/10 text-success border-success/20 badge-sm gap-1.5 pl-2 pr-3 py-3 font-semibold">
            <div class="w-1.5 h-1.5 rounded-full bg-success animate-pulse"></div>
            Active
        </span>
    );

    if (isExpired) {
        statusBadge = <span class="badge badge-error/10 text-error border-error/20 badge-sm py-3 font-semibold px-3 text-[10px] uppercase">Expired</span>;
    } else if (isDepleted && !is_mock) {
        statusBadge = <span class="badge badge-error/10 text-error border-error/20 badge-sm py-3 font-semibold px-3 text-[10px] uppercase">Depleted</span>;
    } else if (is_mock) {
        statusBadge = <span class="badge badge-info/10 text-info border-info/20 badge-sm py-3 font-bold px-3 text-[10px] uppercase">MOCKING</span>;
    }

    // Mode Badge
    const modeBadge = (budget_period && budget_period !== "none") ? (
        <span class="text-[10px] uppercase tracking-wider font-bold text-secondary/70 bg-secondary/5 px-2 py-1 rounded-lg border border-secondary/10">{budget_period}</span>
    ) : (
        <span class="text-[10px] uppercase tracking-wider font-bold text-primary/70 bg-primary/5 px-2 py-1 rounded-lg border border-primary/10">Prepaid</span>
    );

    // Rate Limit Text
    const rateLimitText = (rate_period !== "none" && rate_limit > 0)
        ? `${rate_limit}/${rate_period === "second" ? "sec" : "min"}`
        : "Unlimited";

    return (
        <div class="group relative overflow-hidden bg-white/5 backdrop-blur-md border border-white/5 rounded-[1.5rem] transition-all duration-500 hover:bg-white/[0.08] hover:border-white/10 hover:shadow-2xl hover:shadow-primary/5 hover:-translate-y-0.5">
            {/* Hover Glow */}
            <div class="absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700 pointer-events-none"></div>

            <div class="card-body p-7">
                <div class="flex justify-between items-start">
                    <div class="flex-1">
                        <div class="flex items-center gap-3 mb-3">
                            <h2 class="text-xl font-bold text-white tracking-tight">{name}</h2>
                            <div class="px-2 py-0.5 rounded-md bg-white/5 text-[10px] uppercase font-bold text-white/40 tracking-widest border border-white/5">{provider || "openai"}</div>
                            {statusBadge}
                        </div>
                        <div class="flex items-center gap-2">
                            <code class="text-sm text-white/30 bg-black/20 px-3 py-1.5 rounded-xl font-mono border border-white/5">
                                {prefix}••••••••
                            </code>
                            <button
                                onClick={handleCopy}
                                class="btn btn-ghost btn-xs rounded-lg hover:bg-white/10 transition-colors"
                                title="Copy prefix"
                            >
                                {copied ? (
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                                    </svg>
                                ) : (
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-white/30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
                                    </svg>
                                )}
                            </button>
                        </div>
                    </div>
                    <div class="flex flex-col items-end gap-2">
                        {modeBadge}
                    </div>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-3 gap-6 mt-8 pt-6 border-t border-white/5">
                    <div class="flex flex-col gap-3">
                        <div class="flex justify-between items-end">
                            <span class="text-[10px] font-bold uppercase tracking-widest text-white/30">Budget Allocation</span>
                            <span class="text-xs font-mono font-bold text-white/80">
                                ${budget_usage.toFixed(2)} <span class="text-white/20">/ {budget_limit > 0 ? "$" + budget_limit.toFixed(0) : "∞"}</span>
                            </span>
                        </div>
                        <div class="w-full h-2 bg-white/5 rounded-full overflow-hidden shadow-inner">
                            <div
                                class={`bg-gradient-to-r ${barColor} h-full rounded-full transition-all duration-1000 ease-out shadow-[0_0_10px_rgba(var(--p-rgb),0.3)]`}
                                style={{ width: `${usagePercent}%` }}
                            ></div>
                        </div>
                    </div>
                    <div class="flex flex-col gap-1">
                        <span class="text-[10px] font-bold uppercase tracking-widest text-white/30 text-center sm:text-left">Throughput</span>
                        <div class="font-mono font-bold text-base text-white/90 text-center sm:text-left">{rateLimitText}</div>
                    </div>
                    <div class="flex flex-col gap-1 items-center sm:items-end text-right">
                        <span class="text-[10px] font-bold uppercase tracking-widest text-white/30">Lifecycle</span>
                        <div class={`font-bold text-base ${isExpired ? 'text-error animate-pulse' : 'text-white/90'}`}>{expiresText}</div>
                    </div>
                </div>

                <div class="flex justify-between items-center mt-8 pt-4 border-t border-white/5">
                    <span class="text-[10px] font-bold uppercase tracking-widest text-white/20">Commissioned {new Date(created_at * 1000).toLocaleDateString()}</span>
                    <div class="flex gap-2">
                        <button
                            class="btn btn-sm btn-ghost rounded-xl bg-white/5 text-white/60 hover:text-white hover:bg-white/10 border-none transition-all"
                            onClick={() => onEdit(keyData)}
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                            </svg>
                            Configure
                        </button>
                        <button
                            class="btn btn-sm btn-ghost rounded-xl text-error/60 hover:text-error hover:bg-error/10 border-none transition-all"
                            onClick={() => onRevoke(id)}
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                            Revoke
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}
