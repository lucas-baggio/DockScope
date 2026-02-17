import { NavLink } from 'react-router-dom';
import { Boxes, Image, HardDrive, LayoutDashboard } from 'lucide-react';

const nav = [
  { to: '/dashboard', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/containers', icon: Boxes, label: 'Containers' },
  { to: '/images', icon: Image, label: 'Images' },
  { to: '/volumes', icon: HardDrive, label: 'Volumes' },
];

export function Sidebar() {
  return (
    <aside className="w-56 shrink-0 bg-zinc-900 border-r border-zinc-800 flex flex-col">
      <div className="p-4 border-b border-zinc-800">
        <h1 className="font-semibold text-lg text-white tracking-tight">
          DockScope
        </h1>
        <p className="text-xs text-zinc-500 mt-0.5">Observability</p>
      </div>
      <nav className="p-2 flex-1">
        {nav.map(({ to, icon: Icon, label }) => (
          <NavLink
            key={to}
            to={to}
            end={to === '/' || to === '/dashboard'}
            className={({ isActive }) =>
              `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-zinc-800 text-white'
                  : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'
              }`
            }
          >
            <Icon className="w-5 h-5 shrink-0" />
            {label}
          </NavLink>
        ))}
      </nav>
    </aside>
  );
}
