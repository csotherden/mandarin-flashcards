import { useEffect } from "react";

/**
 * A short-lived bar confirming an action with an Undo button. Sits above the phone tab bar.
 * Dismisses itself after `seconds`; Escape doesn't close it, so a stray key can't lose the undo.
 */
export function UndoBar({ message, onUndo, onClose, seconds = 6 }: { message: string; onUndo: () => void; onClose: () => void; seconds?: number }) {
  useEffect(() => {
    const timer = window.setTimeout(onClose, seconds * 1000);
    return () => window.clearTimeout(timer);
  }, [message, onClose, seconds]);
  return (
    <div className="fixed inset-x-0 bottom-[92px] z-40 flex justify-center px-4 lg:bottom-8 lg:pl-64" role="status" aria-live="polite">
      <div className="rise flex w-full max-w-md items-center gap-3 rounded-2xl bg-ink py-2 pl-4 pr-2 text-white shadow-float">
        <span className="min-w-0 flex-1 truncate">{message}</span>
        <button type="button" onClick={onUndo} className="min-h-11 shrink-0 rounded-xl px-4 font-bold text-gold hover:bg-white/10">
          Undo
        </button>
      </div>
    </div>
  );
}
