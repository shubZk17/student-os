import React, { createContext, useContext, useState, useEffect } from 'react';
import { User } from '../types';
import { apiRequest, saveSession, clearSession } from '../api/client';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, pass: string) => Promise<void>;
  register: (payload: any) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    const savedUser = localStorage.getItem('studentos_user');
    const token = localStorage.getItem('studentos_access_token');
    if (savedUser && token) {
      try {
        setUser(JSON.parse(savedUser));
      } catch (e) {
        clearSession();
      }
    }
    setLoading(false);
  }, []);

  const login = async (email: string, password: string) => {
    const data: any = await apiRequest('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
    saveSession(data.user, data.tokens);
    setUser(data.user);
  };

  const register = async (payload: any) => {
    const data: any = await apiRequest('/auth/register', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    saveSession(data.user, data.tokens);
    setUser(data.user);
  };

  const logout = () => {
    const refreshToken = localStorage.getItem('studentos_refresh_token');
    if (refreshToken) {
      // Best effort: revoke server-side; local session is cleared regardless.
      apiRequest('/auth/logout', {
        method: 'POST',
        body: JSON.stringify({ refresh_token: refreshToken }),
      }).catch(() => {});
    }
    clearSession();
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
