import { RecoveryDialog } from '@/components/RecoveryDialog';
import { useHashRoute } from '@/hooks/useHashRoute';
import { Backup } from '@/pages/Backup/Backup';
import { Dashboard } from '@/pages/Dashboard/Dashboard';
import { Profiles } from '@/pages/Profiles/Profiles';
import { ObjectGrid } from '@/pages/Restore/ObjectGrid';
import { Restore } from '@/pages/Restore/Restore';
import { SettingsPage } from '@/pages/Settings/Settings';

type NavItem = { label: string; path: string };

const NAV: NavItem[] = [
  { label: 'Dashboard', path: '/dashboard' },
  { label: 'Server Profiles', path: '/profiles' },
  { label: 'Backup', path: '/backup' },
  { label: 'Restore', path: '/restore' },
  { label: 'Settings', path: '/settings' },
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
    <div className="min-h-screen flex bg-slate-900 text-slate-100">
      <nav className="w-56 border-r border-slate-800 p-4 flex flex-col gap-2">
        <h1 className="text-lg font-semibold mb-4">MariaDB Tools</h1>
        {NAV.map((item) => (
          <NavLink key={item.path} active={current === item.path} path={item.path}>
            {item.label}
          </NavLink>
        ))}
      </nav>
      <main className="flex-1 p-8 overflow-auto">
        <Page path={current} />
      </main>
      <RecoveryDialog />
    </div>
  );
}

function NavLink({
  active,
  path,
  children,
}: {
  active: boolean;
  path: string;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={() => {
        window.location.hash = `#${path}`;
      }}
      className={`text-left px-3 py-2 rounded text-sm transition-colors ${
        active ? 'bg-slate-700 text-white' : 'text-slate-300 hover:bg-slate-800'
      }`}
    >
      {children}
    </button>
  );
}
