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
    <aside className="w-60 border-r border-slate-800 flex-col py-5 px-3 hidden md:flex min-h-[calc(100vh-4rem)]">
      <nav className="space-y-0.5">
        <div className="eyebrow px-3 pb-2">Menu</div>
        {navItems.map((item) => {
          const Icon = item.icon;
          return (
            <NavLink
              key={item.path}
              to={item.path}
              end={item.path === '/'}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                  isActive
                    ? 'bg-slate-800 font-medium text-white'
                    : 'text-slate-400 hover:text-slate-100 hover:bg-slate-800/50'
                }`
              }
            >
              <Icon className="h-4 w-4 shrink-0" />
              <span>{item.name}</span>
            </NavLink>
          );
        })}
      </nav>
    </aside>
  );
};
