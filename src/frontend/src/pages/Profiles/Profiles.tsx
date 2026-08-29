import { useEffect, useState } from 'react';
import { ListServerProfiles, CreateServerProfile, DeleteServerProfile } from '../../../wailsjs/go/app/App';
import { profile } from '../../../wailsjs/go/models';

type View = profile.View;

export function Profiles() {
  const [views, setViews] = useState<View[]>([]);
  const [name, setName] = useState('');
  const [host, setHost] = useState('');
  const [port, setPort] = useState(3306);
  const [user, setUser] = useState('');
  const [password, setPassword] = useState('');
  const [err, setErr] = useState<string | null>(null);

  async function refresh() {
    try {
      setViews((await ListServerProfiles()) ?? []);
    } catch (e) {
      setErr(String(e));
    }
  }

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
        sslMode: 'preferred',
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
    <section>
      <h2 className="text-xl font-semibold">Server Profiles</h2>
      <p className="mt-1 text-slate-400 text-sm">
        Credentials are encrypted at rest with the app.key beside the binary.
      </p>
      {err && <p className="mt-3 text-red-400 text-sm">{err}</p>}

      <div className="mt-6 grid gap-4 max-w-2xl">
        <label className="grid gap-1">
          <span className="text-xs text-slate-400">Name</span>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="bg-slate-800 border border-slate-700 rounded px-2 py-1"
          />
        </label>
        <div className="grid grid-cols-3 gap-3">
          <label className="grid gap-1 col-span-2">
            <span className="text-xs text-slate-400">Host</span>
            <input
              value={host}
              onChange={(e) => setHost(e.target.value)}
              className="bg-slate-800 border border-slate-700 rounded px-2 py-1"
            />
          </label>
          <label className="grid gap-1">
            <span className="text-xs text-slate-400">Port</span>
            <input
              type="number"
              value={port}
              onChange={(e) => setPort(Number(e.target.value))}
              className="bg-slate-800 border border-slate-700 rounded px-2 py-1"
            />
          </label>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <label className="grid gap-1">
            <span className="text-xs text-slate-400">User</span>
            <input
              value={user}
              onChange={(e) => setUser(e.target.value)}
              className="bg-slate-800 border border-slate-700 rounded px-2 py-1"
            />
          </label>
          <label className="grid gap-1">
            <span className="text-xs text-slate-400">Password</span>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="bg-slate-800 border border-slate-700 rounded px-2 py-1"
            />
          </label>
        </div>
        <button
          type="button"
          onClick={onCreate}
          className="self-start bg-blue-600 hover:bg-blue-500 text-white px-3 py-1.5 rounded"
        >
          Add profile
        </button>
      </div>

      <table className="mt-8 w-full max-w-2xl text-sm">
        <thead>
          <tr className="text-left text-slate-400">
            <th className="py-2">Name</th>
            <th>Host</th>
            <th>User</th>
            <th>SSL</th>
            <th />
          </tr>
        </thead>
        <tbody>
          {views.map((v) => (
            <tr key={v.id} className="border-t border-slate-800">
              <td className="py-2">{v.name}</td>
              <td>{v.host}</td>
              <td>{v.user}</td>
              <td>{v.sslMode}</td>
              <td className="text-right">
                <button
                  type="button"
                  onClick={() => onDelete(v.id)}
                  className="text-red-400 hover:text-red-300 text-xs"
                >
                  Delete
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
