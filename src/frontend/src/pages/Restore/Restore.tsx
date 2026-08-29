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

  const canStart = filePath.trim() && profileId;
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
    <section>
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Restore</h2>
        <Button variant="ghost" onClick={() => navigate('/')}>
          Back to Dashboard
        </Button>
      </div>
      <p className="mt-1 text-slate-400 text-sm">
        Full Restore pipes the entire dump to{' '}
        <code className="bg-slate-900 px-1 rounded">mariadb</code> without scanning. Partial Restore
        first analyzes the dump, then lets you pick tables.
      </p>
      {err && <p className="mt-3 text-red-400 text-sm">{err}</p>}

      <div className="mt-6 max-w-2xl bg-slate-800 border-slate-700 rounded-lg p-6 space-y-4">
        <div className="grid gap-1.5">
          <label htmlFor="file" className="text-sm text-slate-300">
            Dump file (.sql)
          </label>
          <FilePicker id="file" accept=".sql" onChange={setFilePath} />
          {filePath && <p className="text-xs text-slate-500">Selected: {filePath}</p>}
        </div>

        <div className="grid gap-1.5">
          <label htmlFor="profile" className="text-sm text-slate-300">
            Target server profile
          </label>
          <select
            id="profile"
            value={profileId}
            onChange={(e) => setProfileId(e.target.value)}
            className="bg-slate-900 border-slate-700 h-9 rounded-md border px-3 py-1 text-sm text-white w-full"
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

        <div className="flex gap-2">
          <Button data-action="restore" onClick={onRestore} disabled={!canStart || busy}>
            Restore
          </Button>
          <Button
            data-action="analyze"
            variant="ghost"
            onClick={onAnalyze}
            disabled={!canStart || busy}
          >
            Analyze
          </Button>
          {phase === 'restoring' && (
            <Button variant="ghost" onClick={onCancel}>
              Cancel
            </Button>
          )}
        </div>
      </div>

      {phase === 'analyzing' && <p className="mt-4 text-sm text-slate-400">Analyzing dump…</p>}
      {phase === 'restoring' && (
        <div className="mt-6 max-w-2xl">
          <ProgressBar soFar={soFar} total={total} label="Restore progress" />
        </div>
      )}
      {phase === 'done' && result && <p className="mt-4 text-emerald-400 text-sm">{result}</p>}
    </section>
  );
}
