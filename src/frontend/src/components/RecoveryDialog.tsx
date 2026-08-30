import { AlertTriangle, KeyRound } from 'lucide-react';
import { useState } from 'react';
import { RecoveryDecision } from '@/api';
import { useWailsEvent } from '@/hooks/useWailsEvent';
import { Button } from './Button';
import { Dialog } from './Dialog';

type Reason = 'missing_key' | string;

interface ModalEvent {
  Reason: Reason;
}

export function RecoveryDialog() {
  const [open, setOpen] = useState(false);
  const [reason, setReason] = useState<Reason | null>(null);
  const [busy, setBusy] = useState(false);

  useWailsEvent<ModalEvent>('recovery:show', (payload) => {
    setReason(payload?.Reason ?? null);
    setOpen(true);
  });

  const close = () => {
    if (busy) return;
    setOpen(false);
    setReason(null);
  };

  const decide = async (decision: 'cancel' | 'reset') => {
    setBusy(true);
    try {
      await RecoveryDecision(decision);
    } finally {
      setBusy(false);
      setOpen(false);
      setReason(null);
    }
  };

  return (
    <Dialog
      open={open}
      onClose={close}
      title="Smart Recovery"
      footer={
        <>
          <Button variant="ghost" onClick={() => decide('cancel')} disabled={busy}>
            Cancel
          </Button>
          <Button variant="danger" onClick={() => decide('reset')} disabled={busy}>
            Reset & Re-init
          </Button>
        </>
      }
    >
      {reason === 'missing_key' ? (
        <div className="space-y-3">
          <div className="flex items-start gap-3 p-3 bg-amber-500/10 border border-amber-500/20 rounded-xl text-amber-200">
            <KeyRound className="size-5 text-amber-400 shrink-0 mt-0.5" />
            <div className="text-xs leading-relaxed">
              <span className="font-semibold block mb-0.5">Missing Encryption Key</span>
              The{' '}
              <code className="bg-slate-900 px-1 py-0.5 rounded font-mono text-amber-300">
                app.key
              </code>{' '}
              file beside the binary is missing, but an existing catalog was found.
            </div>
          </div>
          <p className="text-slate-300 text-sm">
            Cancel so you can restore the key, or Reset &amp; Re-init to wipe the catalog, generate
            a new key, and start fresh.
          </p>
        </div>
      ) : (
        <div className="flex items-center gap-3 text-slate-300">
          <AlertTriangle className="size-5 text-amber-400 shrink-0" />
          <p>The application needs your attention before continuing.</p>
        </div>
      )}
    </Dialog>
  );
}
