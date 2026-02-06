import type { Key } from "../types";

interface Props {
    keyData: Key;
    onEdit: (key: Key) => void;
    onRevoke: (id: number) => void;
}

export default function KeyCard({ keyData, onEdit, onRevoke }: Props) {
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
    let barColor = "bg-primary";
    if (usagePercent > 80) barColor = "bg-warning";
    if (usagePercent >= 100) barColor = "bg-error";

    const isDepleted = budget_limit > 0 && budget_usage >= budget_limit;

    // Status Badge
    let statusBadge = (
        <span class="badge badge-success badge-sm gap-1 pl-1.5 pr-2.5 py-2.5 font-medium">
            <div class="w-1.5 h-1.5 rounded-full bg-white animate-pulse"></div>
            Active
        </span>
    );

    if (isExpired) {
        statusBadge = <span class="badge badge-error badge-sm py-2.5 font-medium">Expired</span>;
    } else if (isDepleted && !is_mock) {
        statusBadge = <span class="badge badge-error badge-sm py-2.5 font-medium">Depleted</span>;
    } else if (is_mock) {
        statusBadge = <span class="badge badge-info badge-sm py-2.5 font-bold">MOCK</span>;
    }

    // Mode Badge
    const modeBadge = (budget_period && budget_period !== "none") ? (
        <span class="badge badge-secondary/20 text-secondary badge-sm capitalize font-medium">{budget_period}</span>
    ) : (
        <span class="badge badge-primary/20 text-primary badge-sm font-medium">Prepaid</span>
    );

    // Rate Limit Text
    const rateLimitText = (rate_period !== "none" && rate_limit > 0)
        ? `${rate_limit}/${rate_period === "second" ? "sec" : "min"}`
        : "Unlimited";

    return (
        <div class="card bg-base-100/50 backdrop-blur-sm border border-base-content/5 hover:border-primary/30 hover:shadow-lg hover:shadow-primary/5 transition-all duration-300 group">
            <div class="card-body p-6">
                <div class="flex justify-between items-start">
                    <div class="flex-1">
                        <div class="flex items-center gap-3 mb-2">
                            <h2 class="text-lg font-bold text-base-content">{name}</h2>
                            <div class="badge badge-sm badge-outline opacity-40 border-base-content/20 font-mono uppercase text-[10px] tracking-tighter px-1.5">{provider || "openai"}</div>
                            {statusBadge}
                        </div>
                        <code class="text-xs text-base-content/40 bg-base-content/5 px-2 py-1 rounded font-mono">{prefix}</code>
                    </div>
                    <div class="flex flex-col items-end gap-2">
                        {modeBadge}
                    </div>
                </div>

                <div class="grid grid-cols-3 gap-4 mt-6 pt-4 border-t border-base-content/5">
                    <div>
                        <div class="text-xs text-base-content/50 mb-1">Budget</div>
                        <div class="font-mono font-medium text-sm">
                            ${budget_usage.toFixed(2)} <span class="text-base-content/40">/ {budget_limit > 0 ? "$" + budget_limit.toFixed(0) : "∞"}</span>
                        </div>
                        <div class="w-full h-1.5 bg-base-content/10 rounded-full mt-2 overflow-hidden">
                            <div class={`${barColor} h-full rounded-full transition-all`} style={{ width: `${usagePercent}%` }}></div>
                        </div>
                    </div>
                    <div>
                        <div class="text-xs text-base-content/50 mb-1">Rate Limit</div>
                        <div class="font-mono font-medium text-sm">{rateLimitText}</div>
                    </div>
                    <div class="text-right">
                        <div class="text-xs text-base-content/50 mb-1">Expires</div>
                        <div class="font-medium text-sm ${isExpired ? 'text-error' : ''}">{expiresText}</div>
                    </div>
                </div>

                <div class="flex justify-between items-center mt-6 pt-4 border-t border-base-content/5">
                    <span class="text-xs text-base-content/40">Created {new Date(created_at * 1000).toLocaleDateString()}</span>
                    <div class="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        <button
                            class="btn btn-sm btn-ghost text-base-content/60 hover:text-base-content"
                            onClick={() => onEdit(keyData)}
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                            </svg>
                            Edit
                        </button>
                        <button
                            class="btn btn-sm btn-ghost text-error/60 hover:text-error hover:bg-error/10"
                            onClick={() => onRevoke(id)}
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
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
