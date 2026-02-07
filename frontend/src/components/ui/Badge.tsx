import type { ComponentChildren } from "preact";

interface Props {
    children: ComponentChildren;
    variant?: "default" | "success" | "warning" | "error" | "info" | "primary";
    pulse?: boolean;
    className?: string;
}

export default function Badge({ children, variant = "default", pulse, className = "" }: Props) {
    const variants = {
        default: "bg-white/5 border-white/10 text-white/40",
        success: "bg-success/10 border-success/20 text-success",
        warning: "bg-warning/10 border-warning/20 text-warning",
        error: "bg-error/10 border-error/20 text-error",
        info: "bg-info/10 border-info/20 text-info",
        primary: "bg-primary/10 border-primary/20 text-primary",
    };

    const dotColors = {
        default: "bg-white/20",
        success: "bg-success",
        warning: "bg-warning",
        error: "bg-error",
        info: "bg-info",
        primary: "bg-primary",
    };

    return (
        <span class={`flex items-center gap-2 px-2.5 py-0.5 rounded-full border text-[10px] font-bold uppercase tracking-wider ${variants[variant]} ${className}`}>
            <div class={`w-1.5 h-1.5 rounded-full ${dotColors[variant]} ${pulse ? "animate-pulse" : ""}`}></div>
            {children}
        </span>
    );
}
