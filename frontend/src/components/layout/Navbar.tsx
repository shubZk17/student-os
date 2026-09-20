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
      <div className="flex items-center gap-2.5">
        <div className="h-8 w-8 rounded-lg bg-brand-600 flex items-center justify-center">
          <Sparkles className="h-4 w-4 text-white" />
        </div>
        <span className="text-base font-semibold tracking-tight text-white">StudentOS</span>
      </div>

      <div className="flex items-center space-x-4">
        {/* Notifications Popover */}
        <div className="relative">
          <button
            onClick={() => setShowNotifs(!showNotifs)}
            className="btn-ghost relative px-2 py-2"
            title="Notifications"
          >
            <Bell className="h-5 w-5" />
            {unreadCount > 0 && (
              <span className="absolute top-1 right-1 h-4 w-4 rounded-full bg-rose-500 text-2xs font-semibold text-white flex items-center justify-center ring-2 ring-slate-900">
                {unreadCount}
              </span>
            )}
          </button>

          {showNotifs && (
            <div className="absolute right-0 mt-2 w-80 sm:w-96 card bg-slate-900 shadow-xl p-4 z-50">
              <div className="flex items-center justify-between pb-3 border-b border-slate-800">
                <span className="text-sm font-medium text-white">Notifications</span>
                {unreadCount > 0 && (
                  <button
                    onClick={markAllRead}
                    className="text-xs text-brand-400 hover:text-slate-300 flex items-center gap-1"
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
                        n.is_read ? 'border border-transparent bg-slate-800/30 text-slate-400' : 'border border-slate-800 bg-slate-800/60 text-slate-200'
                      }`}
                    >
                      <div className="font-medium text-slate-100 mb-0.5">{n.title}</div>
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
            <div className="h-8 w-8 rounded-full bg-slate-800 border border-slate-700 flex items-center justify-center text-xs font-medium text-slate-300">
              {user.full_name?.charAt(0) || <UserIcon className="h-4 w-4" />}
            </div>
            <div className="hidden sm:block text-left">
              <div className="text-xs font-medium text-white leading-tight">{user.full_name}</div>
              <div className="text-2xs text-slate-500 capitalize">{user.role}</div>
            </div>
            <button
              onClick={logout}
              className="p-1.5 rounded-lg text-slate-500 hover:text-slate-200 hover:bg-slate-800 transition-colors"
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
