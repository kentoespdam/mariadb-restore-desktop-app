import { X } from 'lucide-react';
import type { ReactNode } from 'react';
import { useEffect, useRef } from 'react';

export interface DialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  footer?: ReactNode;
  // ponytail: minimal a11y — focus first interactive child on open,
  // Esc + click-outside close, no focus trap (not needed for our
  // two-button modals). Add focus trap if a future modal needs it.
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
        className="fixed inset-0 z-50 flex items-center justify-center bg-black/60"
        onClick={onClose}
      >
        {/* biome-ignore lint/a11y/noStaticElementInteractions: card body intentionally swallows click events so the backdrop close-on-click does not fire when interacting with the modal content. */}
        {/* biome-ignore lint/a11y/useKeyWithClickEvents: card body has no keyboard semantics; the Escape key closes globally. */}
        <div
          ref={ref}
          className="bg-slate-800 border border-slate-700 rounded-lg p-6 max-w-md w-full shadow-xl"
          onClick={(e) => e.stopPropagation()}
        >
          <div className="flex items-start justify-between gap-4 mb-3">
            <h2 className="text-lg font-semibold text-white">{title}</h2>
            <button
              type="button"
              onClick={onClose}
              aria-label="Close"
              className="text-slate-400 hover:text-white p-1 -m-1"
            >
              <X />
            </button>
          </div>
          <div className="text-sm text-slate-300 space-y-3">{children}</div>
          {footer && <div className="mt-6 flex justify-end gap-2">{footer}</div>}
        </div>
      </div>
    </>
  );
}
