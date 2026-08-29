import { Trash2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  CreateServerProfile,
  DeleteServerProfile,
  ListServerProfiles,
} from '../../../wailsjs/go/app/App';
import type { profile } from '../../../wailsjs/go/models';

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

      <Card className="mt-6 max-w-2xl bg-slate-800 border-slate-700">
        <CardHeader>
          <CardTitle>Add profile</CardTitle>
          <CardDescription>Connection credentials for a MariaDB server.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4">
          <div className="grid gap-1.5">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="bg-slate-900 border-slate-700"
            />
          </div>
          <div className="grid grid-cols-3 gap-3">
            <div className="grid gap-1.5 col-span-2">
              <Label htmlFor="host">Host</Label>
              <Input
                id="host"
                value={host}
                onChange={(e) => setHost(e.target.value)}
                className="bg-slate-900 border-slate-700"
              />
            </div>
            <div className="grid gap-1.5">
              <Label htmlFor="port">Port</Label>
              <Input
                id="port"
                type="number"
                value={port}
                onChange={(e) => setPort(Number(e.target.value))}
                className="bg-slate-900 border-slate-700"
              />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="grid gap-1.5">
              <Label htmlFor="user">User</Label>
              <Input
                id="user"
                value={user}
                onChange={(e) => setUser(e.target.value)}
                className="bg-slate-900 border-slate-700"
              />
            </div>
            <div className="grid gap-1.5">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="bg-slate-900 border-slate-700"
              />
            </div>
          </div>
          <Button onClick={onCreate} className="self-start">
            Add profile
          </Button>
        </CardContent>
      </Card>

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
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => onDelete(v.id)}
                  aria-label="Delete"
                >
                  <Trash2 />
                </Button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
