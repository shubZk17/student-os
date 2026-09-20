import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { Link } from 'react-router-dom';
import {
  Sparkles,
  Briefcase,
  Kanban,
  Calendar,
  ArrowUpRight,
  Clock,
  CheckCircle2,
  AlertCircle,
  MapPin,
} from 'lucide-react';
import { apiRequest } from '../api/client';
import { Opportunity, DashboardSummary } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

export const DashboardPage: React.FC = () => {
  const { user } = useAuth();
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [allRecommendations, setAllRecommendations] = useState<Opportunity[]>([]);
  const [loading, setLoading] = useState(true);

  const [error, setError] = useState('');

  useEffect(() => {
    Promise.all([
      apiRequest<DashboardSummary>('/dashboard/summary'),
      apiRequest<{ recommendations: Opportunity[] }>('/recommendations'),
    ])
      .then(([sumData, recData]) => {
        setSummary(sumData);
        setAllRecommendations(recData.recommendations || []);
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [user]);

  const recommendations = allRecommendations.slice(0, 3);
  const highMatchesCount = allRecommendations.filter((r) => (r.match_score ?? 0) >= 80).length;
  const hour = new Date().getHours();
  const greeting = hour < 12 ? 'Good morning' : hour < 17 ? 'Good afternoon' : 'Good evening';

  if (loading) {
    return (
      <div className="space-y-6 animate-pulse">
        <div className="h-20 bg-slate-800/40 rounded-2xl"></div>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="h-28 bg-slate-800/40 rounded-2xl"></div>
          <div className="h-28 bg-slate-800/40 rounded-2xl"></div>
          <div className="h-28 bg-slate-800/40 rounded-2xl"></div>
          <div className="h-28 bg-slate-800/40 rounded-2xl"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return <ErrorBanner message={error} />;
  }

  return (
    <div className="space-y-8">
      {/* Greeting Banner */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 p-6 rounded-3xl bg-gradient-to-r from-indigo-950/60 via-slate-900 to-slate-900 border border-indigo-500/20 shadow-xl">
        <div>
          <div className="text-xs font-semibold uppercase tracking-wider text-indigo-400 flex items-center space-x-1.5 mb-1">
            <Sparkles className="h-3.5 w-3.5" />
            <span>Personal Command Center</span>
          </div>
          <h1 className="text-2xl md:text-3xl font-extrabold text-white">
            {greeting}, {summary?.student_name.split(' ')[0]} 👋
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Here is what needs your attention across your opportunities, projects, and deadlines today.
          </p>
        </div>

        {/* Profile Strength Widget */}
        <div className="bg-slate-900/80 border border-slate-800 p-4 rounded-2xl min-w-[220px]">
          <div className="flex items-center justify-between text-xs mb-1.5">
            <span className="text-slate-400 font-medium">Profile Strength</span>
            <span className="font-bold text-indigo-400">{summary?.profile_strength}%</span>
          </div>
          <div className="w-full bg-slate-800 h-2.5 rounded-full overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-indigo-500 to-violet-500 rounded-full transition-all duration-500"
              style={{ width: `${summary?.profile_strength}%` }}
            ></div>
          </div>
          <Link
            to="/profile"
            className="text-[11px] text-indigo-400 hover:text-indigo-300 mt-2 block text-right font-medium"
          >
            Add projects & skills +15% →
          </Link>
        </div>
      </div>

      {/* Metric Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-5 rounded-2xl bg-slate-900/60 border border-slate-800 hover:border-indigo-500/30 transition">
          <div className="flex items-center justify-between text-slate-400 mb-3">
            <span className="text-xs font-medium">High Match Roles</span>
            <div className="p-2 rounded-xl bg-indigo-500/10 text-indigo-400">
              <Sparkles className="h-4 w-4" />
            </div>
          </div>
          <div className="text-2xl font-bold text-white">{highMatchesCount}</div>
          <div className="text-[11px] text-emerald-400 flex items-center space-x-1 mt-1">
            <span>90%+ match with your profile</span>
          </div>
        </div>

        <div className="p-5 rounded-2xl bg-slate-900/60 border border-slate-800 hover:border-indigo-500/30 transition">
          <div className="flex items-center justify-between text-slate-400 mb-3">
            <span className="text-xs font-medium">Active Applications</span>
            <div className="p-2 rounded-xl bg-blue-500/10 text-blue-400">
              <Kanban className="h-4 w-4" />
            </div>
          </div>
          <div className="text-2xl font-bold text-white">{summary?.active_applications_count}</div>
          <div className="text-[11px] text-slate-400 mt-1">In progress on Kanban board</div>
        </div>

        <div className="p-5 rounded-2xl bg-slate-900/60 border border-slate-800 hover:border-indigo-500/30 transition">
          <div className="flex items-center justify-between text-slate-400 mb-3">
            <span className="text-xs font-medium">Interviews This Week</span>
            <div className="p-2 rounded-xl bg-purple-500/10 text-purple-400">
              <Calendar className="h-4 w-4" />
            </div>
          </div>
          <div className="text-2xl font-bold text-white">{summary?.upcoming_interviews_count}</div>
          <div className="text-[11px] text-indigo-400 mt-1">Tomorrow: CognitiveScale Labs</div>
        </div>

        <div className="p-5 rounded-2xl bg-slate-900/60 border border-slate-800 hover:border-indigo-500/30 transition">
          <div className="flex items-center justify-between text-slate-400 mb-3">
            <span className="text-xs font-medium">Urgent Deadlines</span>
            <div className="p-2 rounded-xl bg-amber-500/10 text-amber-400">
              <Clock className="h-4 w-4" />
            </div>
          </div>
          <div className="text-2xl font-bold text-white">{summary?.impending_deadlines_count}</div>
          <div className="text-[11px] text-amber-400 mt-1">Closing within 7 days</div>
        </div>
      </div>

      {/* Upcoming Interview Spotlight */}
      {summary?.upcoming_interview && (
        <div className="p-5 rounded-2xl bg-gradient-to-r from-purple-950/30 via-slate-900 to-slate-900 border border-purple-500/20 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
          <div className="flex items-center space-x-4">
            <div className="h-12 w-12 rounded-2xl bg-purple-500/20 border border-purple-500/30 flex items-center justify-center text-purple-400 shrink-0">
              <Calendar className="h-6 w-6" />
            </div>
            <div>
              <div className="text-[10px] font-bold text-purple-400 uppercase tracking-wider">
                Interview Alert
              </div>
              <h2 className="text-base font-bold text-white">
                {summary.upcoming_interview.company} — {summary.upcoming_interview.title}
              </h2>
              <p className="text-xs text-slate-400">
                Scheduled for Tomorrow, 3:00 PM IST (Technical Round 1)
              </p>
            </div>
          </div>
          <Link
            to="/applications"
            className="px-4 py-2 rounded-xl bg-purple-600 hover:bg-purple-500 text-white text-xs font-semibold transition flex items-center space-x-1.5 shadow-lg shadow-purple-600/20"
          >
            <span>View Prep Notes</span>
            <ArrowUpRight className="h-3.5 w-3.5" />
          </Link>
        </div>
      )}

      {/* Personalized Recommendations */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-lg font-bold text-white">Recommended for You</h2>
            <p className="text-xs text-slate-400">
              Multi-signal explainable matches based on your verified skills and academic timeline.
            </p>
          </div>
          <Link
            to="/opportunities"
            className="text-xs font-semibold text-indigo-400 hover:text-indigo-300 flex items-center space-x-1"
          >
            <span>View all opportunities</span>
            <ArrowUpRight className="h-3.5 w-3.5" />
          </Link>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {recommendations.map((job) => (
            <div
              key={job.id}
              className="p-5 rounded-2xl bg-slate-900/60 border border-slate-800 hover:border-indigo-500/40 transition flex flex-col justify-between"
            >
              <div>
                <div className="flex items-start justify-between gap-3 mb-3">
                  <div className="flex items-center space-x-3">
                    {job.company_logo_url ? (
                      <img src={job.company_logo_url} alt={job.company_name} className="h-10 w-10 rounded-xl object-cover bg-slate-800 border border-slate-700" />
                    ) : (
                      <div aria-hidden className="h-10 w-10 rounded-xl shrink-0 bg-slate-800 border border-slate-700 flex items-center justify-center font-bold text-slate-300">
                        {job.company_name.charAt(0)}
                      </div>
                    )}
                    <div>
                      <h3 className="font-bold text-sm text-white hover:text-indigo-300 transition">
                        {job.title}
                      </h3>
                      <p className="text-xs text-slate-400">{job.company_name}</p>
                    </div>
                  </div>
                  {job.match_score && (
                    <span className="px-2.5 py-1 rounded-full text-xs font-extrabold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                      {job.match_score}% Match
                    </span>
                  )}
                </div>

                <div className="flex flex-wrap gap-2 text-[11px] text-slate-400 mb-3">
                  <span className="flex items-center space-x-1">
                    <MapPin className="h-3 w-3 text-slate-500" />
                    <span>{job.location}</span>
                  </span>
                  <span>•</span>
                  <span>{job.stipend_or_salary}</span>
                </div>

                {/* Explainable Matching Highlights */}
                <div className="p-3 rounded-xl bg-slate-950/60 border border-slate-800/80 mb-4 space-y-1">
                  <div className="text-[10px] font-bold text-slate-400 uppercase tracking-wide">
                    Why you match:
                  </div>
                  {job.matched_reasons?.slice(0, 2).map((r, i) => (
                    <div key={i} className="flex items-center space-x-1.5 text-xs text-emerald-400">
                      <CheckCircle2 className="h-3 w-3 shrink-0" />
                      <span className="truncate">{r}</span>
                    </div>
                  ))}
                  {job.missing_requirements && job.missing_requirements.length > 0 && (
                    <div className="flex items-center space-x-1.5 text-xs text-amber-400 pt-0.5">
                      <AlertCircle className="h-3 w-3 shrink-0" />
                      <span className="truncate">{job.missing_requirements[0]}</span>
                    </div>
                  )}
                </div>
              </div>

              <div className="flex items-center justify-between pt-2 border-t border-slate-800">
                <span className="text-[11px] text-slate-500 flex items-center space-x-1">
                  <Clock className="h-3 w-3" />
                  <span>Closing in 2 weeks</span>
                </span>
                <Link
                  to="/opportunities"
                  className="px-3.5 py-1.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold transition"
                >
                  View Details
                </Link>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
