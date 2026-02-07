import { useState } from "preact/hooks";
import type { Key } from "../types";

export function getKeyStatus(key: Key) {
    const { expires_at, budget_usage, configuration } = key;
    let isExpired = false;
    let expiresText = "Never";

    if (expires_at) {
        const expDate = new Date(expires_at * 1000);
        expiresText = expDate.toLocaleDateString();
        if (expDate < new Date()) {
            isExpired = true;
        }
    }

    // Extract budget limit from any middleware with a "limit" role
    let budgetLimit = 0;
    const limitMw = configuration?.middlewares.find(m => {
        const info = (window as any).middlewareInfo?.find((info: any) => info.id === m.id);
        return info && Object.values(info.schema).some((s: any) => s.role === "limit");
    });

    if (limitMw) {
        const info = (window as any).middlewareInfo?.find((info: any) => info.id === limitMw.id);
        const limitKey = Object.keys(info.schema).find(k => info.schema[k].role === "limit");
        if (limitKey) budgetLimit = parseFloat(limitMw.config[limitKey] || "0");
    }

    const isMock = configuration?.provider.id === "mock";
    const usagePercent = budgetLimit > 0 ? Math.min((budget_usage / budgetLimit) * 100, 100) : 0;
    const isDepleted = budgetLimit > 0 && budget_usage >= budgetLimit;

    return { isExpired, expiresText, usagePercent, isDepleted, isMock, budgetLimit };
}

export function StatusBadge({ status, isMock }: { status: ReturnType<typeof getKeyStatus>, isMock: boolean }) {
    const { isExpired, isDepleted } = status;

    if (isExpired) {
        return (
            <span class="flex items-center gap-2 px-3 py-1 rounded-full bg-error/10 border border-error/20 text-[10px] font-black uppercase tracking-widest text-error">
                <div class="w-1.5 h-1.5 rounded-full bg-error"></div>
                Expired
            </span>
        );
    }

    if (isDepleted && !isMock) {
        return (
            <span class="flex items-center gap-2 px-3 py-1 rounded-full bg-warning/10 border border-warning/20 text-[10px] font-black uppercase tracking-widest text-warning">
                <div class="w-1.5 h-1.5 rounded-full bg-warning"></div>
                Capped
            </span>
        );
    }

    if (isMock) {
        return (
            <span class="flex items-center gap-2 px-3 py-1 rounded-full bg-info/10 border border-info/20 text-[10px] font-black uppercase tracking-widest text-info">
                <div class="w-1.5 h-1.5 rounded-full bg-info animate-bounce"></div>
                Simulation
            </span>
        );
    }

    return (
        <span class="flex items-center gap-2 px-3 py-1 rounded-full bg-success/10 border border-success/20 text-[10px] font-black uppercase tracking-widest text-success shadow-[0_0_15px_rgba(var(--s-rgb),0.1)]">
            <div class="w-1.5 h-1.5 rounded-full bg-success animate-pulse shadow-[0_0_8px_rgba(var(--s-rgb),0.8)]"></div>
            Online
        </span>
    );
}

export function ModeBadge({ period }: { period?: any }) {
    const isRecurrent = (typeof period === "number" && period > 0) || (typeof period === "string" && period !== "none" && period !== "");
    const label = isRecurrent ? "Recurrent" : "Disposable";

    return (
        <span class="text-[9px] font-black uppercase tracking-[0.2em] text-white/30 border border-white/5 rounded-lg px-2 py-1 bg-white/[0.02]">{label}</span>
    );
}

export function CopyButton({ text }: { text: string }) {
    const [copied, setCopied] = useState(false);

    const handleCopy = () => {
        navigator.clipboard.writeText(text);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div class="relative group/copy">
            <code class="text-xs font-mono text-white/40 bg-black/20 px-3 py-1.5 rounded-lg border border-white/5 block group-hover/copy:text-white/60 transition-colors">
                {text}••••••••
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
    );
}

export function UsageBar({ percent }: { percent: number }) {
    let barColor = "from-primary to-secondary";
    if (percent > 80) barColor = "from-warning to-error";
    if (percent >= 100) barColor = "from-error to-error";

    return (
        <div class="w-20 h-1 bg-white/5 rounded-full overflow-hidden mt-1.5">
            <div class={`h-full rounded-full bg-gradient-to-r ${barColor} transition-all duration-1000`} style={{ width: `${percent}%` }}></div>
        </div>
    );
}
