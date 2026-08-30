import { HardDriveDownload, Lock, Plus, RotateCcw, Server, Sparkles, Zap } from 'lucide-react';
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
      <section className="space-y-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-white">Dashboard</h2>
          <p className="mt-1 text-slate-400 text-sm">Loading server profiles…</p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="h-28 bg-slate-900/60 border border-slate-800 rounded-2xl animate-pulse"
            />
          ))}
        </div>
      </section>
    );
  }

  if (err) {
    return (
      <section className="space-y-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-white">Dashboard</h2>
          <p className="mt-2 text-rose-400 text-sm bg-rose-500/10 border border-rose-500/20 p-4 rounded-xl">
            {err}
          </p>
        </div>
      </section>
    );
  }

  if (views.length === 0) {
    return (
      <section className="space-y-6">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-white">Dashboard</h2>
          <p className="mt-1 text-slate-400 text-sm">
            Zero-overhead selective restore and logical backup for MariaDB.
          </p>
        </div>

        <div className="bg-slate-900/70 border border-slate-800/80 rounded-2xl p-10 flex flex-col items-center justify-center text-center shadow-xl">
          <div className="size-16 rounded-2xl bg-sky-500/10 border border-sky-500/20 flex items-center justify-center text-sky-400 mb-4">
            <Server className="size-8" />
          </div>
          <h3 className="text-lg font-semibold text-white">Setup Connection</h3>
          <p className="mt-2 text-slate-400 text-sm max-w-md">
            No server profiles yet. Create one to start a backup or restore.
          </p>
          <Button className="mt-6" onClick={() => navigate('/profiles')}>
            <Plus className="size-4" />
            Create a server profile
          </Button>
        </div>
      </section>
    );
  }

  return (
    <section className="space-y-8">
      <div>
        <h2 className="text-2xl font-bold tracking-tight text-white">Dashboard</h2>
        <p className="mt-1 text-slate-400 text-sm">
          Pick a server profile to back up from or restore to.
        </p>
      </div>

      {/* KPI & Quick Shortcut Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-5 shadow-sm relative overflow-hidden group">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">
              Server Profiles
            </span>
            <div className="size-8 rounded-lg bg-sky-500/10 text-sky-400 flex items-center justify-center">
              <Server className="size-4" />
            </div>
          </div>
          <div className="mt-3 text-3xl font-bold text-white tabular-nums">{views.length}</div>
          <div className="mt-1 text-xs text-slate-400 flex items-center gap-1.5">
            <span className="size-1.5 rounded-full bg-emerald-400" />
            <span>Encrypted credentials</span>
          </div>
        </div>

        <button
          type="button"
          onClick={() => navigate('/backup')}
          className="bg-slate-900/80 border border-slate-800 hover:border-indigo-500/40 rounded-2xl p-5 shadow-sm text-left transition-all duration-150 group hover:bg-slate-850 cursor-pointer"
        >
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-400 group-hover:text-indigo-300">
              Logical Backup
            </span>
            <div className="size-8 rounded-lg bg-indigo-500/10 text-indigo-400 group-hover:scale-110 transition-transform flex items-center justify-center">
              <HardDriveDownload className="size-4" />
            </div>
          </div>
          <div className="mt-3 text-sm font-semibold text-slate-100 flex items-center gap-1">
            <span>New Backup Dump</span>
            <Zap className="size-3.5 text-indigo-400" />
          </div>
          <p className="mt-1 text-xs text-slate-400">Stream directly to local .sql</p>
        </button>

        <button
          type="button"
          onClick={() => navigate('/restore')}
          className="bg-slate-900/80 border border-slate-800 hover:border-sky-500/40 rounded-2xl p-5 shadow-sm text-left transition-all duration-150 group hover:bg-slate-850 cursor-pointer"
        >
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-400 group-hover:text-sky-300">
              Zero-Disk Restore
            </span>
            <div className="size-8 rounded-lg bg-sky-500/10 text-sky-400 group-hover:scale-110 transition-transform flex items-center justify-center">
              <RotateCcw className="size-4" />
            </div>
          </div>
          <div className="mt-3 text-sm font-semibold text-slate-100 flex items-center gap-1">
            <span>Selective or Full</span>
            <Sparkles className="size-3.5 text-sky-400" />
          </div>
          <p className="mt-1 text-xs text-slate-400">Virtual AST catalog scanner</p>
        </button>
      </div>

      {/* Profiles List */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-base font-semibold text-slate-200">Configured Servers</h3>
          <Button variant="outline" size="sm" onClick={() => navigate('/profiles')}>
            <Plus className="size-3.5" />
            Add Server
          </Button>
        </div>

        <ul className="space-y-3">
          {views.map((v) => (
            <li
              key={v.id}
              className="bg-slate-900/90 border border-slate-800/90 hover:border-slate-700/80 rounded-2xl p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4 transition-all duration-150 shadow-sm"
            >
              <div className="flex items-start gap-3.5">
                <div className="size-10 rounded-xl bg-slate-800 border border-slate-700/80 flex items-center justify-center text-slate-300 shrink-0 mt-0.5">
                  <Server className="size-5 text-sky-400" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-white text-base">{v.name}</span>
                    <span className="text-[11px] bg-slate-800 text-slate-300 px-2 py-0.5 rounded-full border border-slate-700 font-mono">
                      {v.host}:{v.port}
                    </span>
                  </div>
                  <div className="mt-1 text-xs text-slate-400 flex items-center gap-2 flex-wrap">
                    <span>
                      User: <strong className="text-slate-300 font-normal">{v.user}</strong>
                    </span>
                    <span>•</span>
                    <span className="flex items-center gap-1 text-slate-300">
                      <Lock className="size-3 text-slate-400" />
                      ssl={v.sslMode}
                    </span>
                  </div>
                </div>
              </div>

              <div className="flex items-center gap-2.5 shrink-0 self-end sm:self-center">
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => navigate('/backup', { profile: v.id })}
                >
                  <HardDriveDownload className="size-3.5 text-indigo-400" />
                  Use for Backup
                </Button>
                <Button
                  variant="primary"
                  size="sm"
                  onClick={() => navigate('/restore', { profile: v.id })}
                >
                  <RotateCcw className="size-3.5" />
                  Use for Restore
                </Button>
              </div>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}
