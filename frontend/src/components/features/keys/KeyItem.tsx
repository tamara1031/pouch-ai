import { getKeyStatus } from "./utils";
import StatusBadge from "../../ui/StatusBadge";
import CopyButton from "../../ui/CopyButton";
import ProgressBar from "../../ui/ProgressBar";
import { modalStore } from "../../../lib/store";
import type { Key } from "../../../types";

interface Props {
    keyData: Key;
    onRevoke: (id: number) => void;
}

export default function KeyItem({ keyData, onRevoke }: Props) {
    const {
        id,
        name,
        prefix,
        budget_usage,
        configuration,
    } = keyData;

    const status = getKeyStatus(keyData);
    const { expiresText, usagePercent, isMock, budgetLimit } = status;

    const handleEdit = () => {
        modalStore.openEdit(keyData);
    };

    return (
        <div class="group relative overflow-hidden bg-base-200/50 border border-white/5 rounded-2xl transition-all hover:bg-base-200/80">
            <div class="p-6 relative">
                <div class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-6">
                    <div class="flex-1 space-y-3">
                        <div class="flex flex-wrap items-center gap-3">
                            <h2 class="text-xl font-bold text-white tracking-tight">{name}</h2>
                            <div class="px-2 py-0.5 rounded bg-white/5 text-[9px] font-bold uppercase text-white/40 tracking-wider border border-white/5">{configuration?.provider.id || "openai"}</div>
                            {keyData.auto_renew && (
                                <div class="px-2 py-0.5 rounded bg-primary/10 text-[9px] font-bold uppercase text-primary tracking-wider border border-primary/20">Auto-Renew</div>
                            )}
                            <StatusBadge isExpired={status.isExpired} isDepleted={status.isDepleted} isMock={status.isMock} />
                        </div>
                        <div class="flex items-center gap-2">
                            <CopyButton text={prefix} />
                        </div>
                    </div>

                    <div class="w-full lg:w-auto grid grid-cols-2 md:grid-cols-3 lg:flex lg:items-center gap-4 sm:gap-8">
                        <div class="space-y-1">
                            <span class="text-[9px] font-bold uppercase tracking-wider text-white/20">Usage</span>
                            <div class="flex items-baseline gap-1">
                                <span class="text-lg font-bold text-white tracking-tight">${budget_usage.toFixed(2)}</span>
                                <span class="text-[10px] font-medium text-white/20">/ {budgetLimit > 0 ? "$" + budgetLimit.toFixed(0) : "∞"}</span>
                            </div>
                            <ProgressBar percent={usagePercent} />
                        </div>

                        <div class="space-y-1 hidden md:block">
                            <span class="text-[9px] font-bold uppercase tracking-wider text-white/20">Expiry</span>
                            <div class={`text-sm font-bold tracking-tight ${status.isExpired ? 'text-error' : 'text-white/60'}`}>{expiresText}</div>
                        </div>

                        <div class="flex flex-row lg:flex-row gap-2 justify-end items-center flex-1 sm:flex-none">
                            <button
                                class="btn btn-sm h-9 px-4 rounded-lg bg-white/5 border-none text-[10px] font-bold uppercase tracking-wider text-white/40 hover:text-white hover:bg-white/10 transition-all"
                                onClick={handleEdit}
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
