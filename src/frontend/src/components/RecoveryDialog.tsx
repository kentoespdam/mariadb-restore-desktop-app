import { useState } from 'react';
import { RecoveryDecision } from '@/api';
import { useWailsEvent } from '@/hooks/useWailsEvent';
import { Button } from './Button';
import { Dialog } from './Dialog';

type Reason = 'missing_key' | string;

interface ModalEvent {
  Reason: Reason;
}

// ponytail: CONTEXT says the modal fires when app.key is missing but a
// catalog exists. The backend already decides when to fire ("recovery:show")
// and reads the user's choice back; this component is a pure FE listener.
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
          <Button onClick={() => decide('reset')} disabled={busy}>
            Reset & Re-init
          </Button>
        </>
      }
    >
      {reason === 'missing_key' ? (
        <>
          <p>
            The <code className="bg-slate-900 px-1 rounded">app.key</code> file beside the binary is
            missing, but an existing catalog was found.
          </p>
          <p className="text-slate-400">
            Cancel so you can restore the key, or Reset &amp; Re-init to wipe the catalog, generate
            a new key, and start fresh.
          </p>
        </>
      ) : (
        <p>The application needs your attention before continuing.</p>
      )}
    </Dialog>
  );
}
