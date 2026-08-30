import {
  ArrowLeft,
  CheckCircle2,
  Database,
  FileArchive,
  HardDriveDownload,
  Server,
  XCircle,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { cancelBackup, ListServerProfiles, type profile, startBackup } from '@/api';
import { Button } from '@/components/Button';
import { ProgressBar } from '@/components/ProgressBar';
import { navigate, useHashRoute } from '@/hooks/useHashRoute';
import { useWailsEvent } from '@/hooks/useWailsEvent';

const inputClass =
  'bg-slate-900/90 border-slate-700/80 hover:border-slate-600 focus:border-indigo-500 h-10 rounded-xl border px-3.5 py-2 text-sm text-white w-full transition-all focus:outline-none focus:ring-2 focus:ring-indigo-500/30 placeholder:text-slate-500 disabled:opacity-50 disabled:cursor-not-allowed';

type Phase = 'idle' | 'running' | 'done' | 'error';

export function Backup() {
  const { query } = useHashRoute();
  const presetProfile = query.get('profile') ?? '';

  const [profiles, setProfiles] = useState<profile.View[]>([]);
  const [profileId, setProfileId] = useState<string>(presetProfile);
  const [databases, setDatabases] = useState('');
  const [outputPath, setOutputPath] = useState('');
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
    'backup:progress',
    (p) => {
      if (p.jobId !== jobId) return;
      setSoFar(p.soFar);
      setTotal(p.total);
    },
    { throttleMs: 150 },
  );

  useWailsEvent<{ jobId: string; status: 'success' | 'error'; message?: string }>(
    'backup:done',
    (p) => {
      if (p.jobId !== jobId) return;
      if (p.status === 'success') {
        setPhase('done');
        setResult('Backup completed');
      } else {
        setPhase('error');
        setErr(p.message ?? 'Backup failed');
      }
    },
  );

  const canStart = phase === 'idle' && profileId && databases.trim() && outputPath.trim();

  const onStart = async () => {
    setErr(null);
    setResult(null);
    setSoFar(0);
    setTotal(0);
    setPhase('running');
    try {
      const handle = await startBackup({
        profileId,
        databases: databases
          .split(',')
          .map((s) => s.trim())
          .filter(Boolean),
        outputPath,
      });
      setJobId(handle.jobId);
    } catch (e) {
      setPhase('error');
      setErr(String(e));
    }
  };

  const onCancel = async () => {
    if (!jobId) return;
    try {
      await cancelBackup(jobId);
    } catch (e) {
      setErr(String(e));
    }
  };

  return (
    <section className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-white">Backup</h2>
          <p className="mt-1 text-slate-400 text-sm">
            Run{' '}
            <code className="bg-slate-900 border border-slate-800 px-1.5 py-0.5 rounded font-mono text-indigo-300">
              mariadb-dump
            </code>{' '}
            on demand and save the result to a local file.
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

      {phase === 'done' && result && (
        <div className="flex items-start gap-3 p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-2xl text-emerald-300 text-sm">
          <CheckCircle2 className="size-5 text-emerald-400 shrink-0 mt-0.5" />
          <div>
            <p className="font-semibold text-emerald-200">{result}</p>
            <p className="text-xs text-slate-400 mt-1 font-mono">{outputPath}</p>
          </div>
        </div>
      )}

      <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-6 shadow-sm space-y-6">
        <div className="grid gap-5">
          <div className="grid gap-1.5">
            <label
              htmlFor="profile"
              className="text-xs font-medium text-slate-300 flex items-center gap-1.5"
            >
              <Server className="size-3.5 text-indigo-400" />
              Server profile
            </label>
            <select
              id="profile"
              value={profileId}
              onChange={(e) => setProfileId(e.target.value)}
              className={inputClass}
              disabled={phase === 'running'}
            >
              <option value="">— select —</option>
              {profiles.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name} ({p.host}:{p.port})
                </option>
              ))}
            </select>
          </div>

          <div className="grid gap-1.5">
            <label
              htmlFor="dbs"
              className="text-xs font-medium text-slate-300 flex items-center gap-1.5"
            >
              <Database className="size-3.5 text-indigo-400" />
              Databases (comma-separated)
            </label>
            <input
              id="dbs"
              value={databases}
              onChange={(e) => setDatabases(e.target.value)}
              placeholder="e.g. app, auth, billing"
              className={inputClass}
              disabled={phase === 'running'}
            />
            <span className="text-[11px] text-slate-400">
              Specify one or more database names separated by commas.
            </span>
          </div>

          <div className="grid gap-1.5">
            <label
              htmlFor="out"
              className="text-xs font-medium text-slate-300 flex items-center gap-1.5"
            >
              <FileArchive className="size-3.5 text-indigo-400" />
              Output file (.sql)
            </label>
            <input
              id="out"
              type="text"
              value={outputPath}
              onChange={(e) => setOutputPath(e.target.value)}
              placeholder="backup-2026-08-29.sql"
              className={inputClass}
              disabled={phase === 'running'}
            />
          </div>
        </div>

        <div className="flex gap-3 pt-2 border-t border-slate-800/80">
          <Button
            onClick={onStart}
            disabled={!canStart}
            className="bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 shadow-indigo-950/50"
          >
            <HardDriveDownload className="size-4" />
            Start backup
          </Button>
          {phase === 'running' && (
            <Button variant="ghost" onClick={onCancel}>
              Cancel
            </Button>
          )}
        </div>
      </div>

      {phase === 'running' && (
        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-6 shadow-sm">
          <ProgressBar soFar={soFar} total={total} label="Backup progress" />
        </div>
      )}
    </section>
  );
}
