import { Lock, Plus, Server, ShieldCheck, Trash2, User } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Button } from '@/components/Button';
import {
  CreateServerProfile,
  DeleteServerProfile,
  ListServerProfiles,
} from '../../../wailsjs/go/app/App';
import type { profile } from '../../../wailsjs/go/models';

type View = profile.View;
type SSLMode = 'preferred' | 'required' | 'disabled';

const inputClass =
  'bg-slate-900/90 border-slate-700/80 hover:border-slate-600 focus:border-sky-500 h-10 rounded-xl border px-3.5 py-2 text-sm text-white w-full transition-all focus:outline-none focus:ring-2 focus:ring-sky-500/30 placeholder:text-slate-500';

export function Profiles() {
  const [views, setViews] = useState<View[]>([]);
  const [name, setName] = useState('');
  const [host, setHost] = useState('');
  const [port, setPort] = useState(3306);
  const [user, setUser] = useState('');
  const [password, setPassword] = useState('');
  const [sslMode, setSslMode] = useState<SSLMode>('preferred');
  const [err, setErr] = useState<string | null>(null);

  async function refresh() {
    try {
      setViews((await ListServerProfiles()) ?? []);
    } catch (e) {
      setErr(String(e));
    }
  }

  // biome-ignore lint/correctness/useExhaustiveDependencies: mount-only
  useEffect(() => {
    refresh();
  }, []);

  async function onCreate() {
    setErr(null);
    try {
      await CreateServerProfile({
        name,
        host,
        port,
        user,
        password,
        sslMode,
      } as profile.Input);
      setName('');
      setHost('');
      setUser('');
      setPassword('');
      await refresh();
    } catch (e) {
      setErr(String(e));
    }
  }

  async function onDelete(id: string) {
    try {
      await DeleteServerProfile(id);
      await refresh();
    } catch (e) {
      setErr(String(e));
    }
  }

  return (
    <section className="space-y-8">
      <div>
        <h2 className="text-2xl font-bold tracking-tight text-white">Server Profiles</h2>
        <p className="mt-1 text-slate-400 text-sm flex items-center gap-1.5">
          <ShieldCheck className="size-4 text-emerald-400" />
          Credentials are encrypted at rest with the app.key beside the binary.
        </p>
      </div>

      {err && (
        <p className="text-rose-400 text-sm bg-rose-500/10 border border-rose-500/20 p-4 rounded-xl">
          {err}
        </p>
      )}

      {/* Add Profile Form Card */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-6 shadow-sm space-y-5">
        <div className="border-b border-slate-800/80 pb-4">
          <h3 className="text-base font-semibold text-white flex items-center gap-2">
            <Plus className="size-4 text-sky-400" />
            Add profile
          </h3>
          <p className="text-xs text-slate-400 mt-0.5">
            Connection credentials for a MariaDB server.
          </p>
        </div>

        <div className="grid gap-4">
          <div className="grid gap-1.5">
            <label htmlFor="name" className="text-xs font-medium text-slate-300">
              Name
            </label>
            <input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Production Replica or Local Dev"
              className={inputClass}
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div className="grid gap-1.5 sm:col-span-2">
              <label htmlFor="host" className="text-xs font-medium text-slate-300">
                Host
              </label>
              <input
                id="host"
                value={host}
                onChange={(e) => setHost(e.target.value)}
                placeholder="127.0.0.1 or db.example.com"
                className={inputClass}
              />
            </div>
            <div className="grid gap-1.5">
              <label htmlFor="port" className="text-xs font-medium text-slate-300">
                Port
              </label>
              <input
                id="port"
                type="number"
                value={port}
                onChange={(e) => setPort(Number(e.target.value))}
                className={inputClass}
              />
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div className="grid gap-1.5">
              <label htmlFor="user" className="text-xs font-medium text-slate-300">
                User
              </label>
              <input
                id="user"
                value={user}
                onChange={(e) => setUser(e.target.value)}
                placeholder="root"
                className={inputClass}
              />
            </div>
            <div className="grid gap-1.5">
              <label htmlFor="password" className="text-xs font-medium text-slate-300">
                Password
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                className={inputClass}
              />
            </div>
            <div className="grid gap-1.5">
              <label htmlFor="sslMode" className="text-xs font-medium text-slate-300">
                SSL Mode
              </label>
              <select
                id="sslMode"
                value={sslMode}
                onChange={(e) => setSslMode(e.target.value as SSLMode)}
                className={inputClass}
              >
                <option value="preferred">preferred</option>
                <option value="required">required</option>
                <option value="disabled">disabled</option>
              </select>
            </div>
          </div>

          <div className="pt-2">
            <Button onClick={onCreate} disabled={!name || !host || !user}>
              <Plus className="size-4" />
              Add profile
            </Button>
          </div>
        </div>
      </div>

      {/* Configured Profiles List */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-6 shadow-sm space-y-4">
        <h3 className="text-base font-semibold text-white">Configured Profiles ({views.length})</h3>

        {views.length === 0 ? (
          <p className="text-sm text-slate-500 py-4 text-center">
            No server profiles configured yet.
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-xs uppercase tracking-wider text-slate-400 border-b border-slate-800">
                  <th className="py-3 px-3">Name</th>
                  <th className="py-3 px-3">Host</th>
                  <th className="py-3 px-3">User</th>
                  <th className="py-3 px-3">SSL</th>
                  <th className="py-3 px-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60">
                {views.map((v) => (
                  <tr key={v.id} className="hover:bg-slate-850/50 transition-colors">
                    <td className="py-3 px-3 font-semibold text-white flex items-center gap-2">
                      <Server className="size-4 text-sky-400 shrink-0" />
                      <span>{v.name}</span>
                    </td>
                    <td className="py-3 px-3 font-mono text-xs text-slate-300">
                      {v.host}:{v.port}
                    </td>
                    <td className="py-3 px-3 text-slate-300">
                      <span className="inline-flex items-center gap-1.5">
                        <User className="size-3.5 text-slate-400" />
                        {v.user}
                      </span>
                    </td>
                    <td className="py-3 px-3">
                      <span className="inline-flex items-center gap-1 text-xs bg-slate-800 text-slate-300 px-2 py-0.5 rounded-full border border-slate-700">
                        <Lock className="size-3 text-slate-400" />
                        {v.sslMode}
                      </span>
                    </td>
                    <td className="py-3 px-3 text-right">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => onDelete(v.id)}
                        aria-label="Delete"
                        className="text-slate-400 hover:text-rose-400 hover:bg-rose-500/10"
                      >
                        <Trash2 className="size-4" />
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </section>
  );
}
