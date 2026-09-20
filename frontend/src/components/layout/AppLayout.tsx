import React from 'react';
import { Outlet, NavLink } from 'react-router-dom';
import { Navbar } from './Navbar';
import { Sidebar } from './Sidebar';
import { LayoutDashboard, Compass, Kanban, FolderGit2, User } from 'lucide-react';

const mobileNav = [
  { name: 'Home', path: '/', icon: LayoutDashboard },
  { name: 'Explore', path: '/opportunities', icon: Compass },
  { name: 'Tracker', path: '/applications', icon: Kanban },
  { name: 'Projects', path: '/projects', icon: FolderGit2 },
  { name: 'Profile', path: '/profile', icon: User },
];

export const AppLayout: React.FC = () => {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
      <Navbar />
      <div className="flex-1 flex pb-16 md:pb-0">
        <Sidebar />
        <main className="flex-1 p-4 md:p-8 max-w-7xl mx-auto w-full overflow-x-hidden">
          <Outlet />
        </main>
      </div>

      {/* Mobile Bottom Navigation */}
      <nav className="md:hidden fixed bottom-0 left-0 right-0 h-16 bg-slate-900/90 backdrop-blur-md border-t border-slate-800 flex items-center justify-around z-40 px-2">
        {mobileNav.map((item) => {
          const Icon = item.icon;
          return (
            <NavLink
              key={item.path}
              to={item.path}
              end={item.path === '/'}
              className={({ isActive }) =>
                `flex flex-col items-center justify-center py-1 px-3 text-[10px] font-medium transition ${
                  isActive ? 'text-indigo-400' : 'text-slate-400 hover:text-slate-200'
                }`
              }
            >
              <Icon className="h-5 w-5 mb-0.5" />
              <span>{item.name}</span>
            </NavLink>
          );
        })}
      </nav>
    </div>
  );
};
