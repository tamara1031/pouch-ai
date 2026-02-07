import { useState } from "preact/hooks";

interface Props {
    text: string;
    label?: string;
    className?: string;
}

export default function CopyButton({ text, label, className = "" }: Props) {
    const [copied, setCopied] = useState(false);

    const handleCopy = () => {
        navigator.clipboard.writeText(text);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div class={`relative group/copy inline-flex items-center gap-2 ${className}`}>
            <code class="text-xs font-mono text-white/40 bg-black/20 px-3 py-1.5 rounded-lg border border-white/5 flex items-center group-hover/copy:text-white/60 transition-colors">
                {label || text}
            </code>
            <button
                onClick={handleCopy}
                class="p-1.5 rounded bg-white/5 hover:bg-white/10 opacity-60 group-hover/copy:opacity-100 transition-all"
                title="Copy to clipboard"
            >
                {copied ? (
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
                    </svg>
                ) : (
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-white/40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
                    </svg>
                )}
            </button>
            {copied && (
                <div class="absolute -top-8 left-1/2 -translate-x-1/2 px-2 py-0.5 bg-success text-white text-[9px] font-bold rounded animate-in fade-in zoom-in slide-in-from-bottom-1 duration-200 shadow-lg shadow-success/20">
                    COPIED
                </div>
            )}
        </div>
    );
}
