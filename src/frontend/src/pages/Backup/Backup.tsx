import { useEffect, useState } from 'react';
import { cancelBackup, ListServerProfiles, type profile, startBackup } from '@/api';
import { Button } from '@/components/Button';
import { ProgressBar } from '@/components/ProgressBar';
import { navigate, useHashRoute } from '@/hooks/useHashRoute';
import { useWailsEvent } from '@/hooks/useWailsEvent';

const inputClass =
  'bg-slate-900 border-slate-700 h-9 rounded-md border px-3 py-1 text-sm text-white w-full';

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

  // ponytail: BE throttles too (CONTEXT: 100-250ms) but we also throttle
  // here so React state doesn't drown. Using FE-01's hook.
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
    <section>
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Backup</h2>
        <Button variant="ghost" onClick={() => navigate('/')}>
          Back to Dashboard
        </Button>
      </div>
      <p className="mt-1 text-slate-400 text-sm">
        Run <code className="bg-slate-900 px-1 rounded">mariadb-dump</code> on demand and save the
        result to a local file. (CONTEXT: on-demand only in v1.)
      </p>
      {err && <p className="mt-3 text-red-400 text-sm">{err}</p>}

      <div className="mt-6 max-w-2xl bg-slate-800 border-slate-700 rounded-lg p-6 space-y-4">
        <div className="grid gap-1.5">
          <label htmlFor="profile" className="text-sm text-slate-300">
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
          <label htmlFor="dbs" className="text-sm text-slate-300">
            Databases (comma-separated)
          </label>
          <input
            id="dbs"
            value={databases}
            onChange={(e) => setDatabases(e.target.value)}
            placeholder="app, auth, billing"
            className={inputClass}
            disabled={phase === 'running'}
          />
        </div>

        <div className="grid gap-1.5">
          <label htmlFor="out" className="text-sm text-slate-300">
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

        <div className="flex gap-2">
          <Button onClick={onStart} disabled={!canStart}>
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
        <div className="mt-6 max-w-2xl">
          <ProgressBar soFar={soFar} total={total} label="Backup progress" />
        </div>
      )}
      {phase === 'done' && result && <p className="mt-4 text-emerald-400 text-sm">{result}</p>}
    </section>
  );
}
