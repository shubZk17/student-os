import React, { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { Link, useNavigate } from 'react-router-dom';
import { Sparkles, ArrowRight, Lock, Mail } from 'lucide-react';

export const LoginPage: React.FC = () => {
  const { login } = useAuth();
  const navigate = useNavigate();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await login(email, password);
      navigate('/');
    } catch (err: any) {
      setError(err.message || 'Login failed. Please check your credentials.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col justify-center items-center p-4">
      <div className="w-full max-w-sm space-y-6">
        {/* Brand Header */}
        <div className="text-center space-y-2">
          <div className="h-10 w-10 rounded-lg bg-brand-600 flex items-center justify-center mx-auto">
            <Sparkles className="h-5 w-5 text-white" />
          </div>
          <h1>Sign in to StudentOS</h1>
          <p className="text-sm text-slate-500">
            Discover opportunities matched to your profile.
          </p>
        </div>

        {/* Card */}
        <div className="card p-6 space-y-5">
          {error && (
            <div className="rounded-lg border border-rose-500/30 bg-rose-500/10 p-3 text-xs text-rose-400">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="label">College Email</label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                <input
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="student@college.ac.in"
                  className="input pl-9"
                />
              </div>
            </div>

            <div>
              <label className="label">Password</label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                <input
                  type="password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                  className="input pl-9"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="btn-primary w-full"
            >
              <span>{loading ? 'Signing in�' : 'Sign in'}</span>
              <ArrowRight className="h-4 w-4" />
            </button>
          </form>
        </div>

        <p className="text-center text-sm text-slate-500">
          New to StudentOS?{' '}
          <Link to="/register" className="font-medium text-brand-400 hover:text-slate-300">
            Create an account
          </Link>
        </p>
      </div>
    </div>
  );
};
