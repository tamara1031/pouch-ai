export default function ModeBadge({ period }: { period?: any }) {
    const isRecurrent = (typeof period === "number" && period > 0) || (typeof period === "string" && period !== "none" && period !== "");
    const label = isRecurrent ? "Recurrent" : "Disposable";

    return (
        <span class="text-[9px] font-black uppercase tracking-[0.2em] text-white/30 border border-white/5 rounded-lg px-2 py-1 bg-white/[0.02]">{label}</span>
    );
}
