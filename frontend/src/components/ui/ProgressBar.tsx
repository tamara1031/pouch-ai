interface Props {
    percent: number;
    className?: string;
    size?: "sm" | "md" | "lg";
}

export default function ProgressBar({ percent, className = "", size = "sm" }: Props) {
    let barColor = "from-primary to-secondary";
    if (percent > 80) barColor = "from-warning to-secondary";
    if (percent >= 100) barColor = "from-error to-error";

    const sizes = {
        sm: "h-1",
        md: "h-2",
        lg: "h-3",
    };

    return (
        <div class={`w-full ${sizes[size]} bg-white/5 rounded-full overflow-hidden ${className}`}>
            <div
                class={`h-full rounded-full bg-gradient-to-r ${barColor} transition-all duration-1000 ease-out`}
                style={{ width: `${Math.min(percent, 100)}%` }}
            ></div>
        </div>
    );
}
