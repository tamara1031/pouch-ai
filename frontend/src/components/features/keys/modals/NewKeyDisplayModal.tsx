import { useState } from "preact/hooks";
import { modalStore } from "../../../../lib/store";

export default function NewKeyDisplayModal() {
    const [newKeyCopied, setNewKeyCopied] = useState(false);
    const isOpen = modalStore.isNewKeyOpen.value;
    const newKeyRaw = modalStore.newKeyRaw.value;

    if (!isOpen) return null;

    const handleCopy = () => {
        if (newKeyRaw) {
            navigator.clipboard.writeText(newKeyRaw)
                .then(() => {
                    setNewKeyCopied(true);
                    setTimeout(() => setNewKeyCopied(false), 2000);
                })
                .catch(err => console.error("Failed to copy:", err));
        }
    };

    const handleDone = () => {
        window.dispatchEvent(new CustomEvent('refresh-keys'));
        modalStore.closeNewKey();
    };

    const onClose = () => modalStore.closeNewKey();

    return (
        <div class="modal modal-open modal-bottom sm:modal-middle backdrop-blur-sm transition-all duration-300">
            <div class="modal-box mx-auto max-w-md w-11/12 p-8 bg-base-100 border border-white/5 rounded-2xl shadow-2xl text-center scale-100 opacity-100 animate-in fade-in zoom-in duration-200">
                <div class="w-16 h-16 rounded-full bg-success/10 flex items-center justify-center mx-auto mb-4">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                </div>
                <h2 class="font-bold text-2xl text-white tracking-tight mb-2">API Key Generated</h2>
                <p class="text-white/40 text-sm mb-8">Save this key now. It <span class="text-error/80 font-bold">won't be shown again</span>.</p>

                <div class="bg-black/20 p-4 rounded-xl border border-white/5 flex flex-col gap-3 mb-8">
                    <code class="break-all font-mono font-bold text-xl text-primary tracking-tight">{newKeyRaw || "pk-xxxxxxxx"}</code>
                    <button class={`btn btn-sm btn-ghost bg-white/5 hover:bg-white/10 rounded-lg text-[10px] font-bold uppercase tracking-widest h-9 transition-all ${newKeyCopied ? 'text-success bg-success/10' : ''}`} onClick={handleCopy}>
                        {newKeyCopied ? (
                            <span class="flex items-center gap-2">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                Copied!
                            </span>
                        ) : (
                            "Copy to Clipboard"
                        )}
                    </button>
                </div>

                <button class="w-full btn btn-primary rounded-xl font-bold uppercase tracking-widest text-xs h-12 shadow-lg shadow-primary/20" onClick={handleDone}>
                    Done
                </button>
            </div>
            <div class="modal-backdrop bg-black/40" onClick={onClose}>Close</div>
        </div>
    );
}
