import { useState } from 'react';
import { Dashboard } from './pages/Dashboard/Dashboard';
import { Profiles } from './pages/Profiles/Profiles';

type Route = 'dashboard' | 'profiles';

export function App() {
  const [route, setRoute] = useState<Route>('dashboard');

  return (
    <div className="min-h-screen flex bg-slate-900 text-slate-100">
      <nav className="w-56 border-r border-slate-800 p-4 flex flex-col gap-2">
        <h1 className="text-lg font-semibold mb-4">MariaDB Tools</h1>
        <NavLink active={route === 'dashboard'} onClick={() => setRoute('dashboard')}>
          Dashboard
        </NavLink>
        <NavLink active={route === 'profiles'} onClick={() => setRoute('profiles')}>
          Server Profiles
        </NavLink>
      </nav>
      <main className="flex-1 p-8 overflow-auto">
        {route === 'dashboard' && <Dashboard />}
        {route === 'profiles' && <Profiles />}
      </main>
    </div>
  );
}

function NavLink({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`text-left px-3 py-2 rounded text-sm transition-colors ${
        active ? 'bg-slate-700 text-white' : 'text-slate-300 hover:bg-slate-800'
      }`}
    >
      {children}
    </button>
  );
}
