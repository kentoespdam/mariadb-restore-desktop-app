import { X } from 'lucide-react';
import type { ReactNode } from 'react';
import { useEffect, useRef } from 'react';

export interface DialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  footer?: ReactNode;
}

export function Dialog({ open, onClose, title, children, footer }: DialogProps) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const firstButton = ref.current?.querySelector<HTMLElement>(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
    );
    firstButton?.focus();
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open, onClose]);

  if (!open) return null;

  return (
    <>
      {/* biome-ignore lint/a11y/useKeyWithClickEvents: backdrop closes on click; Escape is handled by the window keydown listener above. */}
      <div
        role="dialog"
        aria-modal="true"
        aria-label={title}
        className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/75 backdrop-blur-sm animate-in fade-in duration-150"
        onClick={onClose}
      >
        {/* biome-ignore lint/a11y/noStaticElementInteractions: card body intentionally swallows click events so the backdrop close-on-click does not fire when interacting with the modal content. */}
        {/* biome-ignore lint/a11y/useKeyWithClickEvents: card body has no keyboard semantics; the Escape key closes globally. */}
        <div
          ref={ref}
          className="bg-slate-900 border border-slate-750 rounded-2xl p-6 max-w-md w-full shadow-2xl shadow-black/80 space-y-4"
          onClick={(e) => e.stopPropagation()}
        >
          <div className="flex items-start justify-between gap-4">
            <h2 className="text-lg font-semibold text-slate-100">{title}</h2>
            <button
              type="button"
              onClick={onClose}
              aria-label="Close"
              className="text-slate-400 hover:text-slate-200 hover:bg-slate-800 p-1.5 rounded-lg transition-colors cursor-pointer"
            >
              <X className="size-4" />
            </button>
          </div>
          <div className="text-sm text-slate-300 space-y-2.5 leading-relaxed">{children}</div>
          {footer && <div className="mt-6 flex justify-end gap-2.5 pt-2">{footer}</div>}
        </div>
      </div>
    </>
  );
}
