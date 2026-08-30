import {
  Database,
  HardDriveDownload,
  LayoutDashboard,
  RotateCcw,
  Server,
  Settings as SettingsIcon,
  ShieldCheck,
} from 'lucide-react';
import type { ComponentType } from 'react';
import { RecoveryDialog } from '@/components/RecoveryDialog';
import { useHashRoute } from '@/hooks/useHashRoute';
import { Backup } from '@/pages/Backup/Backup';
import { Dashboard } from '@/pages/Dashboard/Dashboard';
import { Profiles } from '@/pages/Profiles/Profiles';
import { ObjectGrid } from '@/pages/Restore/ObjectGrid';
import { Restore } from '@/pages/Restore/Restore';
import { SettingsPage } from '@/pages/Settings/Settings';

type NavItem = {
  label: string;
  path: string;
  icon: ComponentType<{ className?: string }>;
};

const NAV: NavItem[] = [
  { label: 'Dashboard', path: '/dashboard', icon: LayoutDashboard },
  { label: 'Server Profiles', path: '/profiles', icon: Server },
  { label: 'Backup', path: '/backup', icon: HardDriveDownload },
  { label: 'Restore', path: '/restore', icon: RotateCcw },
  { label: 'Settings', path: '/settings', icon: SettingsIcon },
];

function Page({ path }: { path: string }) {
  switch (path) {
    case '/profiles':
      return <Profiles />;
    case '/backup':
      return <Backup />;
    case '/restore':
      return <Restore />;
    case '/restore/select':
      return <ObjectGrid />;
    case '/settings':
      return <SettingsPage />;
    case '/dashboard':
    case '/':
      return <Dashboard />;
    default:
      return <Dashboard />;
  }
}

export function App() {
  const { path } = useHashRoute();
  const current = path === '/' ? '/dashboard' : path;

  return (
    <div className="min-h-screen flex bg-slate-950 text-slate-100 selection:bg-sky-500/30 selection:text-sky-200">
      <nav className="w-64 border-r border-slate-800/80 bg-slate-900/60 p-4 flex flex-col justify-between shrink-0 backdrop-blur-md">
        <div className="flex flex-col gap-6">
          <div className="flex items-center gap-3 px-2 py-1">
            <div className="size-9 rounded-xl bg-gradient-to-tr from-sky-500 to-indigo-600 flex items-center justify-center shadow-md shadow-sky-500/20 text-white">
              <Database className="size-5" />
            </div>
            <div>
              <h1 className="text-base font-bold tracking-tight text-white leading-none">
                MariaDB Tools
              </h1>
              <span className="text-[10px] font-medium text-slate-400 tracking-wide uppercase">
                Portable Desktop App
              </span>
            </div>
          </div>

          <div className="flex flex-col gap-1">
            <span className="px-3 text-[11px] font-semibold text-slate-300 uppercase tracking-wider mb-1">
              Menu
            </span>
            {NAV.map((item) => (
              <NavLink
                key={item.path}
                active={current === item.path}
                path={item.path}
                icon={item.icon}
              >
                {item.label}
              </NavLink>
            ))}
          </div>
        </div>

        <div className="pt-4 border-t border-slate-800/80 px-2 space-y-2">
          <div className="flex items-center justify-between text-xs text-slate-400">
            <span className="flex items-center gap-1.5 font-medium text-slate-300">
              <span className="size-2 rounded-full bg-emerald-400" />
              Engine Ready
            </span>
            <span className="text-[11px] font-mono text-slate-300">v1.0</span>
          </div>
          <div className="flex items-center gap-1.5 text-[11px] text-slate-300">
            <ShieldCheck className="size-3.5 text-emerald-400 shrink-0" />
            <span>256-bit AES Key Active</span>
          </div>
        </div>
      </nav>

      <main className="flex-1 p-8 overflow-auto bg-gradient-to-b from-slate-950 via-slate-900/50 to-slate-950">
        <div className="max-w-5xl mx-auto">
          <Page path={current} />
        </div>
      </main>

      <RecoveryDialog />
    </div>
  );
}

function NavLink({
  active,
  path,
  icon: Icon,
  children,
}: {
  active: boolean;
  path: string;
  icon: ComponentType<{ className?: string }>;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={() => {
        window.location.hash = `#${path}`;
      }}
      className={`group flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-150 cursor-pointer ${
        active
          ? 'bg-sky-500/15 text-sky-400 font-semibold shadow-sm border border-sky-500/20'
          : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
      }`}
    >
      <Icon
        className={`size-4 transition-colors ${
          active ? 'text-sky-400' : 'text-slate-400 group-hover:text-slate-300'
        }`}
      />
      <span>{children}</span>
    </button>
  );
}
