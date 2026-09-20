import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { Link } from 'react-router-dom';
import {
  Sparkles,
  Kanban,
  Calendar,
  ArrowRight,
  Clock,
  CheckCircle2,
  AlertCircle,
  MapPin,
} from 'lucide-react';
import { apiRequest } from '../api/client';
import { Opportunity, DashboardSummary } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

const HIGH_MATCH_THRESHOLD = 80;

/** "Tue, 24 Sep, 3:00 PM" — or null when the API gave us nothing usable. */
function formatDateTime(value?: string): string | null {
  if (!value) return null;
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return null;
  return d.toLocaleString(undefined, {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    hour: 'numeric',
    minute: '2-digit',
  });
}

/** "Closes in 12 days" / "Closes today" / null when there's no deadline. */
function formatDeadline(value?: string): string | null {
  if (!value) return null;
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return null;
  const days = Math.ceil((d.getTime() - Date.now()) / 86_400_000);
  if (days < 0) return 'Closed';
  if (days === 0) return 'Closes today';
  if (days === 1) return 'Closes tomorrow';
  return `Closes in ${days} days`;
}

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
  const highMatchesCount = allRecommendations.filter(
    (r) => (r.match_score ?? 0) >= HIGH_MATCH_THRESHOLD,
  ).length;
  const hour = new Date().getHours();
  const greeting = hour < 12 ? 'Good morning' : hour < 17 ? 'Good afternoon' : 'Good evening';
  const today = new Date().toLocaleDateString(undefined, {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
  });
  const interviewAt = formatDateTime(summary?.upcoming_interview?.date);

  if (loading) {
    return (
      <div className="space-y-6 animate-pulse">
        <div className="h-14 rounded-lg bg-slate-900/60"></div>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-px overflow-hidden rounded-lg bg-slate-800">
          <div className="h-24 bg-slate-900"></div>
          <div className="h-24 bg-slate-900"></div>
          <div className="h-24 bg-slate-900"></div>
          <div className="h-24 bg-slate-900"></div>
        </div>
        <div className="h-48 rounded-lg bg-slate-900/60"></div>
      </div>
    );
  }

  if (error) {
    return <ErrorBanner message={error} />;
  }

  const metrics = [
    {
      label: 'High match roles',
      value: highMatchesCount,
      hint: `${HIGH_MATCH_THRESHOLD}% match or above`,
      icon: Sparkles,
    },
    {
      label: 'Active applications',
      value: summary?.active_applications_count ?? 0,
      hint: 'In progress',
      icon: Kanban,
    },
    {
      label: 'Upcoming interviews',
      value: summary?.upcoming_interviews_count ?? 0,
      hint: interviewAt ? `Next: ${interviewAt}` : 'None scheduled',
      icon: Calendar,
    },
    {
      label: 'Closing deadlines',
      value: summary?.impending_deadlines_count ?? 0,
      hint: 'Within 7 days',
      icon: Clock,
    },
  ];

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col gap-4 border-b border-slate-800 pb-6 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1>
            {greeting}
            {summary?.student_name ? `, ${summary.student_name.split(' ')[0]}` : ''}
          </h1>
          <p className="mt-1 text-sm text-slate-500">{today}</p>
        </div>

        <div className="w-full sm:w-64">
          <div className="mb-1.5 flex items-baseline justify-between">
            <span className="text-xs text-slate-400">Profile strength</span>
            <span className="text-xs font-medium tabular-nums text-slate-200">
              {summary?.profile_strength ?? 0}%
            </span>
          </div>
          <div
            className="h-1.5 w-full overflow-hidden rounded-full bg-slate-800"
            role="progressbar"
            aria-valuenow={summary?.profile_strength ?? 0}
            aria-valuemin={0}
            aria-valuemax={100}
            aria-label="Profile strength"
          >
            <div
              className="h-full rounded-full bg-brand-500 transition-all duration-500"
              style={{ width: `${summary?.profile_strength ?? 0}%` }}
            />
          </div>
          <Link
            to="/profile"
            className="mt-2 inline-block text-xs text-slate-400 hover:text-slate-200"
          >
            Complete your profile
          </Link>
        </div>
      </div>

      {/* Metrics: one bordered group, not four floating cards. */}
      <div className="grid grid-cols-1 gap-px overflow-hidden rounded-lg border border-slate-800 bg-slate-800 sm:grid-cols-2 lg:grid-cols-4">
        {metrics.map((m) => {
          const Icon = m.icon;
          return (
            <div key={m.label} className="bg-slate-900/60 p-4">
              <div className="mb-3 flex items-center gap-2 text-slate-400">
                <Icon className="h-3.5 w-3.5" />
                <span className="eyebrow">{m.label}</span>
              </div>
              <div className="text-2xl font-semibold tabular-nums text-white">{m.value}</div>
              <div className="mt-1 text-xs text-slate-500">{m.hint}</div>
            </div>
          );
        })}
      </div>

      {/* Next interview */}
      {summary?.upcoming_interview && (
        <div className="card flex flex-col items-start justify-between gap-4 p-4 sm:flex-row sm:items-center">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-slate-700 bg-slate-800 text-slate-300">
              <Calendar className="h-4 w-4" />
            </div>
            <div>
              <div className="eyebrow">Next interview</div>
              <h2 className="text-sm font-medium text-white">
                {summary.upcoming_interview.company} — {summary.upcoming_interview.title}
              </h2>
              {interviewAt && <p className="text-xs text-slate-500">{interviewAt}</p>}
            </div>
          </div>
          <Link to="/applications" className="btn-secondary w-full sm:w-auto">
            <span>Open tracker</span>
            <ArrowRight className="h-3.5 w-3.5" />
          </Link>
        </div>
      )}

      {/* Recommendations */}
      <section>
        <div className="mb-4 flex items-end justify-between gap-4">
          <div>
            <h2>Recommended for you</h2>
            <p className="mt-0.5 text-sm text-slate-500">
              Matched on your verified skills, target roles and graduation year.
            </p>
          </div>
          <Link
            to="/opportunities"
            className="flex shrink-0 items-center gap-1 text-xs text-slate-400 hover:text-slate-200"
          >
            <span>View all</span>
            <ArrowRight className="h-3.5 w-3.5" />
          </Link>
        </div>

        {recommendations.length === 0 ? (
          <div className="card p-8 text-center">
            <p className="text-sm text-slate-400">No recommendations yet.</p>
            <p className="mt-1 text-xs text-slate-500">
              Add your skills and target roles to start getting matches.
            </p>
            <Link to="/profile" className="btn-secondary mt-4">
              Update profile
            </Link>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            {recommendations.map((job) => {
              const closing = formatDeadline(job.deadline);
              return (
                <article key={job.id} className="card card-hover flex flex-col justify-between p-4">
                  <div>
                    <div className="mb-3 flex items-start justify-between gap-3">
                      <div className="flex items-center gap-3">
                        {job.company_logo_url ? (
                          <img
                            src={job.company_logo_url}
                            alt=""
                            className="h-9 w-9 rounded-lg border border-slate-700 bg-slate-800 object-cover"
                          />
                        ) : (
                          <div
                            aria-hidden
                            className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-slate-700 bg-slate-800 text-sm font-medium text-slate-300"
                          >
                            {job.company_name.charAt(0)}
                          </div>
                        )}
                        <div>
                          <h3>{job.title}</h3>
                          <p className="text-xs text-slate-500">{job.company_name}</p>
                        </div>
                      </div>
                      {job.match_score != null && (
                        <span className="badge-success shrink-0 tabular-nums">
                          {job.match_score}% match
                        </span>
                      )}
                    </div>

                    <div className="mb-3 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-slate-500">
                      <span className="flex items-center gap-1">
                        <MapPin className="h-3 w-3" />
                        {job.location}
                      </span>
                      {job.stipend_or_salary && <span>{job.stipend_or_salary}</span>}
                    </div>

                    {(job.matched_reasons?.length || job.missing_requirements?.length) && (
                      <div className="mb-4 space-y-1.5 rounded-lg border border-slate-800 bg-slate-950/50 p-3">
                        {job.matched_reasons?.slice(0, 2).map((r, i) => (
                          <div key={i} className="flex items-start gap-1.5 text-xs text-slate-300">
                            <CheckCircle2 className="mt-0.5 h-3 w-3 shrink-0 text-emerald-500" />
                            <span className="truncate">{r}</span>
                          </div>
                        ))}
                        {job.missing_requirements?.slice(0, 1).map((r, i) => (
                          <div key={i} className="flex items-start gap-1.5 text-xs text-slate-300">
                            <AlertCircle className="mt-0.5 h-3 w-3 shrink-0 text-amber-500" />
                            <span className="truncate">{r}</span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>

                  <div className="flex items-center justify-between border-t border-slate-800 pt-3">
                    <span className="text-xs text-slate-500">{closing ?? 'No deadline listed'}</span>
                    <Link to="/opportunities" className="btn-secondary py-1.5">
                      View details
                    </Link>
                  </div>
                </article>
              );
            })}
          </div>
        )}
      </section>
    </div>
  );
};
