import {
  ArrowLeft,
  CheckCircle2,
  FileCode,
  RotateCcw,
  Search,
  Server,
  Sparkles,
  XCircle,
  Zap,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import {
  analyzeDump,
  cancelRestore,
  ListServerProfiles,
  type profile,
  startFullRestore,
} from '@/api';
import { Button } from '@/components/Button';
import { FilePicker } from '@/components/FilePicker';
import { ProgressBar } from '@/components/ProgressBar';
import { navigate, useHashRoute } from '@/hooks/useHashRoute';
import { useWailsEvent } from '@/hooks/useWailsEvent';

const inputClass =
  'bg-slate-900/90 border-slate-700/80 hover:border-slate-600 focus:border-sky-500 h-10 rounded-xl border px-3.5 py-2 text-sm text-white w-full transition-all focus:outline-none focus:ring-2 focus:ring-sky-500/30 placeholder:text-slate-500 disabled:opacity-50 disabled:cursor-not-allowed';

type Phase = 'idle' | 'analyzing' | 'restoring' | 'done' | 'error';

export function Restore() {
  const { query } = useHashRoute();
  const presetProfile = query.get('profile') ?? '';

  const [profiles, setProfiles] = useState<profile.View[]>([]);
  const [profileId, setProfileId] = useState<string>(presetProfile);
  const [filePath, setFilePath] = useState<string>('');
  const [phase, setPhase] = useState<Phase>('idle');
  const [jobId, setJobId] = useState<string | null>(null);
  const [soFar, setSoFar] = useState(0);
  const [total, setTotal] = useState(0);
  const [err, setErr] = useState<string | null>(null);
  const [result, setResult] = useState<string | null>(null);

  useEffect(() => {
    (async () => {
      try {
        setProfiles((await ListServerProfiles()) ?? []);
      } catch (e) {
        setErr(String(e));
      }
    })();
  }, []);

  useWailsEvent<{ jobId: string; soFar: number; total: number }>(
    'restore:progress',
    (p) => {
      if (p.jobId !== jobId) return;
      setSoFar(p.soFar);
      setTotal(p.total);
    },
    { throttleMs: 150 },
  );

  useWailsEvent<{ jobId: string; status: 'success' | 'error'; message?: string }>(
    'restore:done',
    (p) => {
      if (p.jobId !== jobId) return;
      if (p.status === 'success') {
        setPhase('done');
        setResult('Restore completed');
      } else {
        setPhase('error');
        setErr(p.message ?? 'Restore failed');
      }
    },
  );

  const canStart = Boolean(filePath.trim() && profileId);
  const busy = phase === 'restoring' || phase === 'analyzing';

  const onRestore = async () => {
    setErr(null);
    setResult(null);
    setSoFar(0);
    setTotal(0);
    setPhase('restoring');
    try {
      const handle = await startFullRestore({ filePath, profileId });
      setJobId(handle.jobId);
    } catch (e) {
      setPhase('error');
      setErr(String(e));
    }
  };

  const onAnalyze = async () => {
    setErr(null);
    setPhase('analyzing');
    try {
      await analyzeDump(filePath);
      navigate('/restore/select', { file: filePath, profile: profileId });
    } catch (e) {
      setPhase('error');
      setErr(String(e));
    }
  };

  const onCancel = async () => {
    if (!jobId) return;
    try {
      await cancelRestore(jobId);
    } catch (e) {
      setErr(String(e));
    }
  };

  return (
    <section className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-white">Restore</h2>
          <p className="mt-1 text-slate-400 text-sm">
            Full Restore pipes the entire dump to{' '}
            <code className="bg-slate-900 border border-slate-800 px-1.5 py-0.5 rounded font-mono text-sky-300">
              mariadb
            </code>{' '}
            without scanning. Partial Restore first analyzes the dump, then lets you pick tables.
          </p>
        </div>
        <Button variant="ghost" onClick={() => navigate('/')}>
          <ArrowLeft className="size-4" />
          Back to Dashboard
        </Button>
      </div>

      {/* Mode Comparison Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-5 shadow-sm space-y-2">
          <div className="flex items-center gap-2.5 text-sky-400 font-semibold text-sm">
            <div className="size-7 rounded-lg bg-sky-500/10 flex items-center justify-center">
              <Zap className="size-4" />
            </div>
            <span>Full Stream Restore</span>
          </div>
          <p className="text-xs text-slate-400 leading-relaxed">
            Directly pipes raw dump bytes to the MariaDB CLI subprocess. Optimal for restoring the
            complete archive with zero temporary disk usage.
          </p>
        </div>

        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-5 shadow-sm space-y-2">
          <div className="flex items-center gap-2.5 text-indigo-400 font-semibold text-sm">
            <div className="size-7 rounded-lg bg-indigo-500/10 flex items-center justify-center">
              <Sparkles className="size-4" />
            </div>
            <span>Selective / Partial Restore</span>
          </div>
          <p className="text-xs text-slate-400 leading-relaxed">
            Fast byte-offset scanner indexes tables, triggers, and routines into a lightweight
            SQLite catalog so you can choose exactly what to restore.
          </p>
        </div>
      </div>

      {err && (
        <div className="flex items-start gap-3 p-4 bg-rose-500/10 border border-rose-500/20 rounded-2xl text-rose-300 text-sm">
          <XCircle className="size-5 text-rose-400 shrink-0 mt-0.5" />
          <p>{err}</p>
        </div>
      )}

      {phase === 'done' && result && (
        <div className="flex items-start gap-3 p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-2xl text-emerald-300 text-sm">
          <CheckCircle2 className="size-5 text-emerald-400 shrink-0 mt-0.5" />
          <div>
            <p className="font-semibold text-emerald-200">{result}</p>
            <p className="text-xs text-slate-400 mt-1 font-mono">{filePath}</p>
          </div>
        </div>
      )}

      {/* Input Configuration Card */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-6 shadow-sm space-y-6">
        <div className="grid gap-5">
          <div className="grid gap-1.5">
            <label
              htmlFor="file"
              className="text-xs font-medium text-slate-300 flex items-center gap-1.5"
            >
              <FileCode className="size-3.5 text-sky-400" />
              Dump file (.sql)
            </label>
            <FilePicker id="file" accept=".sql" onChange={setFilePath} disabled={busy} />
            {filePath && (
              <div className="flex items-center gap-2 text-xs text-slate-400 bg-slate-950/60 border border-slate-800/80 px-3 py-1.5 rounded-lg font-mono truncate">
                <span className="text-sky-400 font-semibold">Selected:</span>
                <span className="truncate">{filePath}</span>
              </div>
            )}
          </div>

          <div className="grid gap-1.5">
            <label
              htmlFor="profile"
              className="text-xs font-medium text-slate-300 flex items-center gap-1.5"
            >
              <Server className="size-3.5 text-sky-400" />
              Target server profile
            </label>
            <select
              id="profile"
              value={profileId}
              onChange={(e) => setProfileId(e.target.value)}
              className={inputClass}
              disabled={busy}
            >
              <option value="">— select —</option>
              {profiles.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name} ({p.host}:{p.port})
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-3 pt-2 border-t border-slate-800/80">
          <Button
            data-action="restore"
            onClick={onRestore}
            disabled={!canStart || busy}
            className="bg-sky-600 hover:bg-sky-500 active:bg-sky-700 shadow-sky-950/50"
          >
            <RotateCcw className="size-4" />
            Restore
          </Button>
          <Button
            data-action="analyze"
            variant="secondary"
            onClick={onAnalyze}
            disabled={!canStart || busy}
          >
            <Search className="size-4 text-indigo-400" />
            Analyze
          </Button>
          {phase === 'restoring' && (
            <Button variant="ghost" onClick={onCancel}>
              Cancel
            </Button>
          )}
        </div>
      </div>

      {phase === 'analyzing' && (
        <div className="flex items-center gap-3 p-5 bg-slate-900/80 border border-slate-800 rounded-2xl">
          <div className="size-4 border-2 border-indigo-400 border-t-transparent rounded-full animate-spin shrink-0" />
          <p className="text-sm text-slate-300 font-medium">Analyzing dump…</p>
        </div>
      )}

      {phase === 'restoring' && (
        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-6 shadow-sm">
          <ProgressBar soFar={soFar} total={total} label="Restore progress" />
        </div>
      )}
    </section>
  );
}
