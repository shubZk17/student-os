import React, { useState, useEffect } from 'react';
import { useAuth } from '../../context/AuthContext';
import { Bell, Sparkles, LogOut, CheckCircle2, User as UserIcon } from 'lucide-react';
import { apiRequest } from '../../api/client';
import { NotificationItem } from '../../types';

export const Navbar: React.FC = () => {
  const { user, logout } = useAuth();
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const [unreadCount, setUnreadCount] = useState<number>(0);
  const [showNotifs, setShowNotifs] = useState<boolean>(false);

  useEffect(() => {
    if (user) {
      apiRequest<{ unread_count: number; notifications: NotificationItem[] }>('/notifications')
        .then((res) => {
          setNotifications(res.notifications || []);
          setUnreadCount(res.unread_count || 0);
        })
        .catch(() => {
          // Non-critical: leave the bell empty if notifications fail to load.
        });
    }
  }, [user]);

  const markAllRead = () => {
    // ponytail: one request per unread item (max 50); add a bulk endpoint if that grows.
    notifications
      .filter((n) => !n.is_read)
      .forEach((n) => apiRequest(`/notifications/${n.id}/read`, { method: 'PATCH' }).catch(() => {}));
    setUnreadCount(0);
    setNotifications((prev) => prev.map((n) => ({ ...n, is_read: true })));
  };

  return (
    <header className="sticky top-0 z-30 h-16 bg-slate-900/80 backdrop-blur-md border-b border-slate-800 px-4 md:px-8 flex items-center justify-between">
      <div className="flex items-center space-x-3">
        <div className="h-9 w-9 rounded-xl bg-gradient-to-tr from-indigo-500 to-violet-500 flex items-center justify-center shadow-lg shadow-indigo-500/20">
          <Sparkles className="h-5 w-5 text-white" />
        </div>
        <div>
          <span className="font-extrabold text-lg tracking-tight bg-gradient-to-r from-white via-slate-100 to-indigo-300 bg-clip-text text-transparent">
            StudentOS
          </span>
          <span className="ml-2 text-[10px] font-semibold tracking-wide uppercase px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
            Production MVP
          </span>
        </div>
      </div>

      <div className="flex items-center space-x-4">
        {/* Notifications Popover */}
        <div className="relative">
          <button
            onClick={() => setShowNotifs(!showNotifs)}
            className="p-2 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800/60 transition relative"
            title="Notifications"
          >
            <Bell className="h-5 w-5" />
            {unreadCount > 0 && (
              <span className="absolute top-1 right-1 h-4 w-4 rounded-full bg-rose-500 text-[10px] font-bold text-white flex items-center justify-center ring-2 ring-slate-900">
                {unreadCount}
              </span>
            )}
          </button>

          {showNotifs && (
            <div className="absolute right-0 mt-2 w-80 sm:w-96 rounded-2xl bg-slate-900 border border-slate-800 shadow-2xl p-4 z-50 animate-in fade-in zoom-in-95 duration-100">
              <div className="flex items-center justify-between pb-3 border-b border-slate-800">
                <span className="font-semibold text-sm text-white">Notifications</span>
                {unreadCount > 0 && (
                  <button
                    onClick={markAllRead}
                    className="text-xs text-indigo-400 hover:text-indigo-300 flex items-center space-x-1"
                  >
                    <CheckCircle2 className="h-3.5 w-3.5" />
                    <span>Mark all as read</span>
                  </button>
                )}
              </div>
              <div className="mt-2 space-y-2 max-h-72 overflow-y-auto">
                {notifications.length === 0 ? (
                  <p className="text-xs text-slate-500 text-center py-6">No notifications yet.</p>
                ) : (
                  notifications.map((n) => (
                    <div
                      key={n.id}
                      className={`p-3 rounded-xl text-xs transition ${
                        n.is_read ? 'bg-slate-800/30 text-slate-400' : 'bg-indigo-950/40 text-slate-200 border border-indigo-500/20'
                      }`}
                    >
                      <div className="font-semibold text-slate-100 mb-0.5">{n.title}</div>
                      <div className="line-clamp-2">{n.message}</div>
                    </div>
                  ))
                )}
              </div>
            </div>
          )}
        </div>

        {/* User Badge */}
        {user && (
          <div className="flex items-center space-x-3 pl-3 border-l border-slate-800">
            <div className="h-8 w-8 rounded-full bg-gradient-to-tr from-indigo-600 to-violet-600 flex items-center justify-center font-bold text-xs text-white shadow-md">
              {user.full_name?.charAt(0) || <UserIcon className="h-4 w-4" />}
            </div>
            <div className="hidden sm:block text-left">
              <div className="text-xs font-semibold text-white leading-tight">{user.full_name}</div>
              <div className="text-[10px] text-slate-400 capitalize">{user.role}</div>
            </div>
            <button
              onClick={logout}
              className="p-1.5 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-slate-800 transition"
              title="Logout"
            >
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        )}
      </div>
    </header>
  );
};
