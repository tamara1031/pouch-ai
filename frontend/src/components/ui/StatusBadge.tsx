interface StatusProps {
    isExpired: boolean;
    isDepleted: boolean;
    isMock: boolean;
}

export default function StatusBadge({ isExpired, isDepleted, isMock }: StatusProps) {
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
