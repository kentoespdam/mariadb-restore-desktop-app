import { useEffect, useState } from 'react';
import { getSettings, resetAndReinit, type Settings, saveSettings } from '@/api';
import { Button } from '@/components/Button';
import { Dialog } from '@/components/Dialog';
import { navigate } from '@/hooks/useHashRoute';

const inputClass =
  'bg-slate-900 border-slate-700 h-9 rounded-md border px-3 py-1 text-sm text-white w-full';

export function SettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null);
  const [mariadb, setMariadb] = useState('');
  const [mariadbDump, setMariadbDump] = useState('');
  const [err, setErr] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);
  const [confirm, setConfirm] = useState(false);
  const [resetting, setResetting] = useState(false);

  useEffect(() => {
    (async () => {
      try {
        const s = await getSettings();
        setSettings(s);
        setMariadb(s.mariadbPath);
        setMariadbDump(s.mariadbDumpPath);
      } catch (e) {
        setErr(String(e));
      }
    })();
  }, []);

  const onSave = async () => {
    setErr(null);
    setSaved(false);
    try {
      await saveSettings({ mariadbPath: mariadb, mariadbDumpPath: mariadbDump });
      setSaved(true);
    } catch (e) {
      setErr(String(e));
    }
  };

  const onReset = async () => {
    setResetting(true);
    try {
      await resetAndReinit();
      setConfirm(false);
    } catch (e) {
      setErr(String(e));
    } finally {
      setResetting(false);
    }
  };

  return (
    <section>
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Settings</h2>
        <Button variant="ghost" onClick={() => navigate('/')}>
          Back to Dashboard
        </Button>
      </div>
      {err && <p className="mt-3 text-red-400 text-sm">{err}</p>}

      <div className="mt-6 max-w-2xl space-y-6">
        <div className="bg-slate-800 border-slate-700 rounded-lg p-6 space-y-3">
          <h3 className="font-semibold">Executable Scope</h3>
          <p className="text-sm text-slate-400">
            All app files live beside the binary (CONTEXT: Executable Scope).
          </p>
          <dl className="text-sm grid grid-cols-[10rem_1fr] gap-y-1">
            <dt className="text-slate-400">Binary directory</dt>
            <dd className="font-mono break-all">{settings?.exeDir ?? '…'}</dd>
            <dt className="text-slate-400">Catalog path</dt>
            <dd className="font-mono break-all">{settings?.catalogPath ?? '…'}</dd>
            <dt className="text-slate-400">app.key path</dt>
            <dd className="font-mono break-all">{settings?.appKeyPath ?? '…'}</dd>
          </dl>
        </div>

        <div className="bg-slate-800 border-slate-700 rounded-lg p-6 space-y-4">
          <h3 className="font-semibold">MariaDB binaries</h3>
          <div className="grid gap-1.5">
            <label htmlFor="mariadb" className="text-sm text-slate-300">
              mariadb path
            </label>
            <input
              id="mariadb"
              value={mariadb}
              onChange={(e) => setMariadb(e.target.value)}
              className={inputClass}
            />
          </div>
          <div className="grid gap-1.5">
            <label htmlFor="mariadbDump" className="text-sm text-slate-300">
              mariadb-dump path
            </label>
            <input
              id="mariadbDump"
              value={mariadbDump}
              onChange={(e) => setMariadbDump(e.target.value)}
              className={inputClass}
            />
          </div>
          <div className="flex items-center gap-3">
            <Button onClick={onSave}>Save</Button>
            {saved && <span className="text-emerald-400 text-sm">Saved</span>}
          </div>
        </div>

        <div className="bg-slate-800 border-slate-700 rounded-lg p-6 space-y-3">
          <h3 className="font-semibold">Encryption</h3>
          <p className="text-sm text-slate-400">
            Key fingerprint:{' '}
            <span className="font-mono text-slate-200">{settings?.keyBits ?? 256}-bit AES-GCM</span>{' '}
            (never the key bytes).
          </p>
        </div>

        <div className="bg-slate-800 border-red-900 rounded-lg p-6 space-y-3">
          <h3 className="font-semibold text-red-300">Danger zone</h3>
          <p className="text-sm text-slate-400">
            Reset &amp; Re-init wipes the catalog, generates a new app.key, and starts fresh.
          </p>
          <Button variant="ghost" onClick={() => setConfirm(true)}>
            Reset & Re-init
          </Button>
        </div>
      </div>

      <Dialog
        open={confirm}
        onClose={() => !resetting && setConfirm(false)}
        title="Reset & Re-init"
        footer={
          <>
            <Button variant="ghost" onClick={() => setConfirm(false)} disabled={resetting}>
              Cancel
            </Button>
            <Button onClick={onReset} disabled={resetting}>
              Confirm reset
            </Button>
          </>
        }
      >
        <p>
          This will wipe the catalog, generate a new app.key, and start fresh. Any existing server
          profiles will be lost.
        </p>
        <p className="text-slate-400">
          The Smart Recovery dialog will guide you through the next step.
        </p>
      </Dialog>
    </section>
  );
}
