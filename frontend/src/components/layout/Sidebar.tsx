import React from 'react';
import { NavLink } from 'react-router-dom';
import {
  LayoutDashboard,
  Compass,
  Kanban,
  FolderGit2,
  User,
} from 'lucide-react';

const navItems = [
  { name: 'Dashboard', path: '/', icon: LayoutDashboard },
  { name: 'Opportunities', path: '/opportunities', icon: Compass },
  { name: 'Applications', path: '/applications', icon: Kanban },
  { name: 'Projects', path: '/projects', icon: FolderGit2 },
  { name: 'Profile', path: '/profile', icon: User },
];

export const Sidebar: React.FC = () => {
  return (
    <aside className="w-64 border-r border-slate-800 bg-slate-900/50 flex flex-col justify-between py-6 px-4 hidden md:flex min-h-[calc(100vh-4rem)]">
      <nav className="space-y-1.5">
        <div className="px-3 pb-2 text-[10px] font-semibold text-slate-500 uppercase tracking-wider">
          Student Portal
        </div>
        {navItems.map((item) => {
          const Icon = item.icon;
          return (
            <NavLink
              key={item.path}
              to={item.path}
              end={item.path === '/'}
              className={({ isActive }) =>
                `flex items-center space-x-3 px-3.5 py-2.5 rounded-xl text-sm font-medium transition ${
                  isActive
                    ? 'bg-indigo-600/10 text-indigo-400 border border-indigo-500/20 shadow-sm'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/40'
                }`
              }
            >
              <Icon className="h-4 w-4" />
              <span>{item.name}</span>
            </NavLink>
          );
        })}
      </nav>

      {/* College Life Wedge Info */}
      <div className="p-4 rounded-2xl bg-gradient-to-br from-indigo-950/40 to-slate-900 border border-indigo-500/10 text-xs text-slate-400">
        <div className="font-semibold text-white mb-1">StudentOS Engine</div>
        <p className="text-[11px] leading-relaxed text-slate-400">
          Personalized multi-signal scoring powered by your verified skills & projects.
        </p>
      </div>
    </aside>
  );
};
