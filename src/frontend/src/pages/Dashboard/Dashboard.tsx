import { useEffect, useState } from 'react';
import { ListServerProfiles, type profile } from '@/api';
import { Button } from '@/components/Button';
import { navigate } from '@/hooks/useHashRoute';

export function Dashboard() {
  const [views, setViews] = useState<profile.View[]>([]);
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    (async () => {
      try {
        setViews((await ListServerProfiles()) ?? []);
      } catch (e) {
        setErr(String(e));
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  if (loading) {
    return (
      <section>
        <h2 className="text-xl font-semibold">Dashboard</h2>
        <p className="mt-2 text-slate-400">Loading server profiles…</p>
      </section>
    );
  }

  if (err) {
    return (
      <section>
        <h2 className="text-xl font-semibold">Dashboard</h2>
        <p className="mt-2 text-red-400 text-sm">{err}</p>
      </section>
    );
  }

  if (views.length === 0) {
    return (
      <section>
        <h2 className="text-xl font-semibold">Dashboard</h2>
        <p className="mt-2 text-slate-400">
          No server profiles yet. Create one to start a backup or restore.
        </p>
        <Button className="mt-4" onClick={() => navigate('/profiles')}>
          Create a server profile
        </Button>
      </section>
    );
  }

  return (
    <section>
      <h2 className="text-xl font-semibold">Dashboard</h2>
      <p className="mt-1 text-slate-400 text-sm">
        Pick a server profile to back up from or restore to.
      </p>
      <ul className="mt-6 space-y-3 max-w-2xl">
        {views.map((v) => (
          <li
            key={v.id}
            className="bg-slate-800 border border-slate-700 rounded-lg p-4 flex items-center justify-between gap-4"
          >
            <div>
              <div className="font-semibold">{v.name}</div>
              <div className="text-sm text-slate-400">
                {v.host}:{v.port} · {v.user} · ssl={v.sslMode}
              </div>
            </div>
            <div className="flex gap-2">
              <Button onClick={() => navigate('/backup', { profile: v.id })}>Use for Backup</Button>
              <Button onClick={() => navigate('/restore', { profile: v.id })}>
                Use for Restore
              </Button>
            </div>
          </li>
        ))}
      </ul>
    </section>
  );
}
