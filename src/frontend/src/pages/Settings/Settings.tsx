import {
  AlertTriangle,
  ArrowLeft,
  Check,
  FolderTree,
  Save,
  ShieldCheck,
  Terminal,
  Trash2,
  XCircle,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { getSettings, resetAndReinit, type Settings, saveSettings } from '@/api';
import { Button } from '@/components/Button';
import { Dialog } from '@/components/Dialog';
import { navigate } from '@/hooks/useHashRoute';

const inputClass =
  'bg-slate-900/90 border-slate-700/80 hover:border-slate-600 focus:border-sky-500 h-10 rounded-xl border px-3.5 py-2 text-sm text-white w-full transition-all focus:outline-none focus:ring-2 focus:ring-sky-500/30 placeholder:text-slate-500 font-mono text-xs';

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
    <section className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-white">Settings</h2>
          <p className="mt-1 text-slate-400 text-sm">
            Configure executable scope paths, MariaDB CLI binaries, and encryption parameters.
          </p>
        </div>
        <Button variant="ghost" onClick={() => navigate('/')}>
          <ArrowLeft className="size-4" />
          Back to Dashboard
        </Button>
      </div>

      {err && (
        <div className="flex items-start gap-3 p-4 bg-rose-500/10 border border-rose-500/20 rounded-2xl text-rose-300 text-sm">
          <XCircle className="size-5 text-rose-400 shrink-0 mt-0.5" />
          <p>{err}</p>
        </div>
      )}

      <div className="space-y-6">
        {/* Executable Scope Card */}
        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-6 shadow-sm space-y-4">
          <div className="flex items-center gap-2.5">
            <div className="size-8 rounded-lg bg-sky-500/10 text-sky-400 flex items-center justify-center">
              <FolderTree className="size-4" />
            </div>
            <div>
              <h3 className="text-base font-semibold text-white">Executable Scope</h3>
              <p className="text-xs text-slate-400">
                All app files live beside the binary (CONTEXT: Executable Scope).
              </p>
            </div>
          </div>

          <div className="bg-slate-950/60 border border-slate-800/80 rounded-xl p-4 space-y-3 text-xs">
            <div className="grid grid-cols-1 md:grid-cols-[11rem_1fr] gap-1 items-center">
              <span className="text-slate-400 font-medium">Binary directory</span>
              <span className="font-mono text-slate-200 bg-slate-900 px-2.5 py-1 rounded border border-slate-800 break-all select-all">
                {settings?.exeDir ?? '…'}
              </span>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-[11rem_1fr] gap-1 items-center">
              <span className="text-slate-400 font-medium">Catalog path</span>
              <span className="font-mono text-slate-200 bg-slate-900 px-2.5 py-1 rounded border border-slate-800 break-all select-all">
                {settings?.catalogPath ?? '…'}
              </span>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-[11rem_1fr] gap-1 items-center">
              <span className="text-slate-400 font-medium">app.key path</span>
              <span className="font-mono text-slate-200 bg-slate-900 px-2.5 py-1 rounded border border-slate-800 break-all select-all">
                {settings?.appKeyPath ?? '…'}
              </span>
            </div>
          </div>
        </div>

        {/* MariaDB Binaries Card */}
        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-6 shadow-sm space-y-5">
          <div className="flex items-center gap-2.5">
            <div className="size-8 rounded-lg bg-indigo-500/10 text-indigo-400 flex items-center justify-center">
              <Terminal className="size-4" />
            </div>
            <div>
              <h3 className="text-base font-semibold text-white">MariaDB binaries</h3>
              <p className="text-xs text-slate-400">
                Absolute paths to the <code className="text-indigo-300">mariadb</code> and{' '}
                <code className="text-indigo-300">mariadb-dump</code> command line utilities.
              </p>
            </div>
          </div>

          <div className="grid gap-4">
            <div className="grid gap-1.5">
              <label htmlFor="mariadb" className="text-xs font-medium text-slate-300">
                mariadb path
              </label>
              <input
                id="mariadb"
                value={mariadb}
                onChange={(e) => setMariadb(e.target.value)}
                placeholder="/usr/bin/mariadb"
                className={inputClass}
              />
            </div>
            <div className="grid gap-1.5">
              <label htmlFor="mariadbDump" className="text-xs font-medium text-slate-300">
                mariadb-dump path
              </label>
              <input
                id="mariadbDump"
                value={mariadbDump}
                onChange={(e) => setMariadbDump(e.target.value)}
                placeholder="/usr/bin/mariadb-dump"
                className={inputClass}
              />
            </div>
          </div>

          <div className="flex items-center gap-3 pt-2 border-t border-slate-800/80">
            <Button onClick={onSave}>
              <Save className="size-4" />
              Save
            </Button>
            {saved && (
              <span className="flex items-center gap-1.5 text-xs text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 px-3 py-1.5 rounded-lg font-medium">
                <Check className="size-3.5" />
                Saved
              </span>
            )}
          </div>
        </div>

        {/* Encryption Card */}
        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-6 shadow-sm space-y-3">
          <div className="flex items-center gap-2.5">
            <div className="size-8 rounded-lg bg-emerald-500/10 text-emerald-400 flex items-center justify-center">
              <ShieldCheck className="size-4" />
            </div>
            <div>
              <h3 className="text-base font-semibold text-white">Encryption</h3>
              <p className="text-xs text-slate-400">
                Hardware-accelerated AES-GCM encrypted catalog storage.
              </p>
            </div>
          </div>
          <div className="text-xs text-slate-300 bg-slate-950/60 border border-slate-800/80 rounded-xl p-3.5 flex items-center justify-between">
            <span>Key fingerprint:</span>
            <span className="font-mono text-emerald-400 font-semibold bg-emerald-500/10 border border-emerald-500/20 px-2 py-0.5 rounded">
              {settings?.keyBits ?? 256}-bit AES-GCM
            </span>
          </div>
        </div>

        {/* Danger Zone Card */}
        <div className="bg-slate-900/80 border border-rose-900/40 rounded-2xl p-6 shadow-sm space-y-4">
          <div className="flex items-center gap-2.5">
            <div className="size-8 rounded-lg bg-rose-500/10 text-rose-400 flex items-center justify-center">
              <AlertTriangle className="size-4" />
            </div>
            <div>
              <h3 className="text-base font-semibold text-rose-300">Danger zone</h3>
              <p className="text-xs text-slate-400">
                Reset &amp; Re-init wipes the catalog, generates a new app.key, and starts fresh.
              </p>
            </div>
          </div>

          <div className="pt-2">
            <Button
              variant="outline"
              onClick={() => setConfirm(true)}
              className="border-rose-800/80 text-rose-300 hover:bg-rose-950/40 hover:border-rose-700 hover:text-rose-200"
            >
              <Trash2 className="size-4" />
              Reset & Re-init
            </Button>
          </div>
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
            <Button variant="danger" onClick={onReset} disabled={resetting}>
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
