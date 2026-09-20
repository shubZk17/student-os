import React, { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { Link, useNavigate } from 'react-router-dom';
import { Sparkles, ArrowRight, Lock, Mail, User, GraduationCap, Briefcase } from 'lucide-react';

export const RegisterPage: React.FC = () => {
  const { register } = useAuth();
  const navigate = useNavigate();

  const [fullName, setFullName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [college, setCollege] = useState('');
  const [degree, setDegree] = useState('B.Tech');
  const [gradYear, setGradYear] = useState(2026);
  const [targetRoles, setTargetRoles] = useState('Software Engineer, ML Intern');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await register({
        full_name: fullName,
        email,
        password,
        college,
        degree,
        graduation_year: gradYear,
        target_roles: targetRoles.split(',').map((r) => r.trim()).filter(Boolean),
      });
      navigate('/');
    } catch (err: any) {
      setError(err.message || 'Registration failed. Please verify details.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col justify-center items-center p-4 py-8">
      <div className="max-w-md w-full space-y-6">
        <div className="text-center space-y-2">
          <div className="h-10 w-10 rounded-lg bg-brand-600 flex items-center justify-center mx-auto">
            <Sparkles className="h-6 w-6 text-white" />
          </div>
          <h1 className="">Create your StudentOS</h1>
          <p className="text-xs text-slate-400">
            Get personalized opportunity recommendations tailored to your batch and skills.
          </p>
        </div>

        <div className="card p-6 space-y-5">
          {error && (
            <div className="p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-xs text-rose-400">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="label">Full Name *</label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                <input
                  type="text"
                  required
                  value={fullName}
                  onChange={(e) => setFullName(e.target.value)}
                  placeholder="e.g. Shubham Verma"
                  className="input pl-9"
                />
              </div>
            </div>

            <div>
              <label className="label">College Email *</label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                <input
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="name@college.ac.in"
                  className="input pl-9"
                />
              </div>
            </div>

            <div>
              <label className="label">Password *</label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                <input
                  type="password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Min. 8 characters"
                  className="input pl-9"
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="label">Degree</label>
                <div className="relative">
                  <GraduationCap className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-500" />
                  <input
                    type="text"
                    value={degree}
                    onChange={(e) => setDegree(e.target.value)}
                    placeholder="B.Tech"
                    className="input pl-9"
                  />
                </div>
              </div>
              <div>
                <label className="label">Grad Year</label>
                <input
                  type="number"
                  value={gradYear}
                  onChange={(e) => setGradYear(Number(e.target.value))}
                  className="input"
                />
              </div>
            </div>

            <div>
              <label className="label">Target Roles</label>
              <div className="relative">
                <Briefcase className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                <input
                  type="text"
                  value={targetRoles}
                  onChange={(e) => setTargetRoles(e.target.value)}
                  placeholder="e.g. Software Engineer, ML Intern"
                  className="input pl-9"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="btn-primary w-full"
            >
              <span>{loading ? 'Creating Account...' : 'Get Started'}</span>
              <ArrowRight className="h-4 w-4" />
            </button>
          </form>
        </div>

        <p className="text-center text-xs text-slate-400">
          Already have an account?{' '}
          <Link to="/login" className="font-medium text-brand-400 hover:text-slate-300">
            Sign In
          </Link>
        </p>
      </div>
    </div>
  );
};
