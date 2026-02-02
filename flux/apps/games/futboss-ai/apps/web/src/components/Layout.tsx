// FutBoss AI - Layout Component
// Author: Bruno Lucena (bruno@lucena.cloud)

import { Outlet, Link, useLocation } from 'react-router-dom';
import { useStore } from '../store';

const navItems = [
  { path: '/', label: 'Dashboard', icon: '🏠' },
  { path: '/team', label: 'Team', icon: '⚽' },
  { path: '/market', label: 'Market', icon: '🏪' },
  { path: '/match', label: 'Match', icon: '🎮' },
  { path: '/wallet', label: 'Wallet', icon: '💰' },
];

export default function Layout() {
  const location = useLocation();
  const { user, balance } = useStore();

  return (
    <div className="min-h-screen flex">
      {/* Sidebar */}
      <aside className="w-64 bg-darker border-r border-primary/20 p-4 flex flex-col">
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-primary">⚽ FutBoss AI</h1>
          <p className="text-sm text-gray-400">by Bruno Lucena</p>
        </div>

        <nav className="flex-1">
          {navItems.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              className={`flex items-center gap-3 px-4 py-3 rounded-lg mb-2 transition-colors ${
                location.pathname === item.path
                  ? 'bg-primary text-dark font-bold'
                  : 'hover:bg-primary/10 text-gray-300'
              }`}
            >
              <span>{item.icon}</span>
              <span>{item.label}</span>
            </Link>
          ))}
        </nav>

        {/* User info */}
        <div className="card mt-auto">
          <p className="text-sm text-gray-400">Balance</p>
          <p className="text-2xl font-bold text-accent">{balance.toLocaleString()} FTC</p>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 p-8 overflow-auto">
        <Outlet />
      </main>
    </div>
  );
}

