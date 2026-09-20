import React, { useState, useEffect } from 'react';
import {
  Search,
  Filter,
  MapPin,
  Clock,
  CheckCircle2,
  AlertCircle,
  ExternalLink,
  Bookmark,
  BookmarkCheck,
  Send,
  Sparkles,
} from 'lucide-react';
import { apiRequest } from '../api/client';
import { Opportunity } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

export const OpportunitiesPage: React.FC = () => {
  const [opportunities, setOpportunities] = useState<Opportunity[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedType, setSelectedType] = useState('ALL');
  const [selectedLocation, setSelectedLocation] = useState('ALL');
  const [remoteOnly, setRemoteOnly] = useState(false);
  const [activeModalJob, setActiveModalJob] = useState<Opportunity | null>(null);
  const [savedJobs, setSavedJobs] = useState<Record<string, boolean>>({});
  const [appliedJobs, setAppliedJobs] = useState<Record<string, boolean>>({});
  const [error, setError] = useState('');

  useEffect(() => {
    fetchRecommendations();
  }, []);

  const fetchRecommendations = () => {
    setLoading(true);
    apiRequest<{ recommendations: Opportunity[] }>('/recommendations')
      .then((res) => {
        setOpportunities(res.recommendations || []);
        setLoading(false);
      })
      .catch((err: Error) => {
        setError(err.message);
        setLoading(false);
      });
  };

  const handleTrackApplication = async (job: Opportunity, stage: string = 'APPLIED') => {
    // Saving is one-way here; untracking happens from the Applications board.
    if (stage === 'SAVED' && savedJobs[job.id]) return;
    try {
      await apiRequest('/applications', {
        method: 'POST',
        body: JSON.stringify({
          opportunity_id: job.id,
          stage,
          notes: job.match_score != null ? `Tracked from Opportunities. Match score: ${job.match_score}%` : '',
        }),
      });
    } catch (err: any) {
      setError(err.message);
      return;
    }
    if (stage === 'SAVED') {
      setSavedJobs((prev) => ({ ...prev, [job.id]: true }));
    } else {
      setAppliedJobs((prev) => ({ ...prev, [job.id]: true }));
    }
  };

  const filteredOpportunities = opportunities.filter((job) => {
    if (selectedType !== 'ALL' && job.type !== selectedType) return false;
    if (remoteOnly && !job.is_remote) return false;
    if (selectedLocation !== 'ALL' && !job.location.toLowerCase().includes(selectedLocation.toLowerCase()))
      return false;
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      return (
        job.title.toLowerCase().includes(q) ||
        job.company_name.toLowerCase().includes(q) ||
        job.required_skills.some((s) => s.toLowerCase().includes(q))
      );
    }
    return true;
  });

  return (
    <div className="space-y-6">
      {/* Header & Description */}
      <div>
        <h1 className="text-2xl font-extrabold text-white tracking-tight">
          Personalized Opportunity Discovery
        </h1>
        <p className="text-xs text-slate-400 mt-1">
          Showing high-signal opportunities ranked deterministically by your skills, eligibility, and preferences.
        </p>
      </div>

      {error && <ErrorBanner message={error} onDismiss={() => setError('')} />}

      {/* Filter & Search Bar */}
      <div className="p-4 rounded-2xl bg-slate-900/70 border border-slate-800 space-y-4">
        <div className="flex flex-col md:flex-row gap-3 items-center">
          {/* Search Input */}
          <div className="relative flex-1 w-full">
            <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
            <input
              type="text"
              placeholder="Search by role, company, or skill (e.g. PyTorch, Go, Bangalore)..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full bg-slate-950/80 border border-slate-700/60 rounded-xl pl-10 pr-4 py-2.5 text-xs text-slate-200 placeholder:text-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition"
            />
          </div>

          {/* Quick Filter Buttons */}
          <div className="flex flex-wrap gap-2 w-full md:w-auto">
            {['ALL', 'INTERNSHIP', 'JOB', 'HACKATHON'].map((t) => (
              <button
                key={t}
                onClick={() => setSelectedType(t)}
                className={`px-3 py-2 rounded-xl text-xs font-semibold transition ${
                  selectedType === t
                    ? 'bg-indigo-600 text-white'
                    : 'bg-slate-800/60 text-slate-400 hover:text-white hover:bg-slate-800'
                }`}
              >
                {t === 'ALL' ? 'All Types' : t.charAt(0) + t.slice(1).toLowerCase()}
              </button>
            ))}

            <button
              onClick={() => setRemoteOnly(!remoteOnly)}
              className={`px-3 py-2 rounded-xl text-xs font-semibold border transition ${
                remoteOnly
                  ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
                  : 'bg-slate-800/40 text-slate-400 border-slate-700 hover:text-white'
              }`}
            >
              Remote Only
            </button>
          </div>
        </div>

        {/* Location Filter Pills */}
        <div className="flex items-center space-x-2 text-xs text-slate-400 overflow-x-auto pb-1">
          <span className="shrink-0 flex items-center space-x-1">
            <Filter className="h-3 w-3" />
            <span>Locations:</span>
          </span>
          {['ALL', 'Bangalore', 'Hyderabad', 'Pune', 'Delhi'].map((loc) => (
            <button
              key={loc}
              onClick={() => setSelectedLocation(loc)}
              className={`px-2.5 py-1 rounded-lg text-[11px] font-medium transition ${
                selectedLocation === loc
                  ? 'bg-slate-700 text-white'
                  : 'bg-slate-800/40 text-slate-400 hover:text-slate-200'
              }`}
            >
              {loc}
            </button>
          ))}
        </div>
      </div>

      {/* Opportunity Grid */}
      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 animate-pulse">
          <div className="h-64 bg-slate-800/40 rounded-2xl"></div>
          <div className="h-64 bg-slate-800/40 rounded-2xl"></div>
          <div className="h-64 bg-slate-800/40 rounded-2xl"></div>
          <div className="h-64 bg-slate-800/40 rounded-2xl"></div>
        </div>
      ) : filteredOpportunities.length === 0 ? (
        <div className="p-12 text-center rounded-3xl bg-slate-900/40 border border-slate-800">
          <div className="h-12 w-12 rounded-2xl bg-slate-800 flex items-center justify-center mx-auto text-slate-500 mb-3">
            <Search className="h-6 w-6" />
          </div>
          <h3 className="font-bold text-sm text-white">No matching opportunities yet</h3>
          <p className="text-xs text-slate-400 max-w-sm mx-auto mt-1">
            Try adjusting your search criteria or adding more skills and target locations to your profile.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
          {filteredOpportunities.map((job) => (
            <div
              key={job.id}
              className="p-5 rounded-3xl bg-slate-900/70 border border-slate-800 hover:border-indigo-500/40 transition flex flex-col justify-between shadow-sm hover:shadow-indigo-500/5"
            >
              <div>
                {/* Header */}
                <div className="flex items-start justify-between gap-3 mb-3">
                  <div className="flex items-center space-x-3.5">
                    {job.company_logo_url ? (
                      <img src={job.company_logo_url} alt={job.company_name} className="h-11 w-11 rounded-2xl object-cover bg-slate-800 border border-slate-700" />
                    ) : (
                      <div aria-hidden className="h-11 w-11 rounded-2xl shrink-0 bg-slate-800 border border-slate-700 flex items-center justify-center font-bold text-slate-300">
                        {job.company_name.charAt(0)}
                      </div>
                    )}
                    <div>
                      <h2
                        onClick={() => setActiveModalJob(job)}
                        className="font-bold text-base text-white hover:text-indigo-300 transition cursor-pointer"
                      >
                        {job.title}
                      </h2>
                      <div className="text-xs text-slate-400 font-medium">{job.company_name}</div>
                    </div>
                  </div>

                  {job.match_score ? (
                    <div className="flex flex-col items-end">
                      <span className="px-3 py-1 rounded-full text-xs font-black bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 shadow-sm">
                        {job.match_score}% Match
                      </span>
                    </div>
                  ) : (
                    <span className="px-2.5 py-0.5 rounded-full text-[10px] font-semibold bg-slate-800 text-slate-400">
                      {job.type}
                    </span>
                  )}
                </div>

                {/* Sub-details */}
                <div className="flex flex-wrap items-center gap-y-1 gap-x-3 text-xs text-slate-400 mb-3.5">
                  <span className="flex items-center space-x-1">
                    <MapPin className="h-3.5 w-3.5 text-slate-500" />
                    <span>{job.location}</span>
                  </span>
                  <span>•</span>
                  <span className="font-medium text-slate-300">{job.stipend_or_salary}</span>
                  <span>•</span>
                  <span className="capitalize">{job.job_type.toLowerCase()}</span>
                </div>

                <p className="text-xs text-slate-300 line-clamp-2 mb-4 leading-relaxed">
                  {job.description}
                </p>

                {/* Explainable Signals */}
                <div className="p-3.5 rounded-2xl bg-slate-950/70 border border-slate-800/80 mb-4 space-y-1.5">
                  <div className="text-[10px] font-bold text-slate-400 uppercase tracking-wider flex items-center space-x-1">
                    <Sparkles className="h-3 w-3 text-indigo-400" />
                    <span>Why you match:</span>
                  </div>
                  {job.matched_reasons?.slice(0, 2).map((r, i) => (
                    <div key={i} className="flex items-center space-x-2 text-xs text-emerald-400">
                      <CheckCircle2 className="h-3.5 w-3.5 shrink-0" />
                      <span className="truncate">{r}</span>
                    </div>
                  ))}
                  {job.missing_requirements && job.missing_requirements.length > 0 && (
                    <div className="flex items-center space-x-2 text-xs text-amber-400 pt-0.5">
                      <AlertCircle className="h-3.5 w-3.5 shrink-0" />
                      <span className="truncate">{job.missing_requirements[0]}</span>
                    </div>
                  )}
                </div>
              </div>

              {/* Actions Footer */}
              <div className="flex items-center justify-between pt-3 border-t border-slate-800/80">
                <div className="text-[11px] text-slate-500 flex items-center space-x-1">
                  <Clock className="h-3.5 w-3.5" />
                  <span>Closing soon</span>
                </div>

                <div className="flex items-center space-x-2">
                  <button
                    onClick={() => handleTrackApplication(job, 'SAVED')}
                    className={`p-2 rounded-xl border transition ${
                      savedJobs[job.id]
                        ? 'bg-indigo-600/10 text-indigo-400 border-indigo-500/30'
                        : 'bg-slate-800/40 text-slate-400 border-slate-700 hover:text-white'
                    }`}
                    title={savedJobs[job.id] ? 'Saved to Tracker' : 'Save for later'}
                  >
                    {savedJobs[job.id] ? (
                      <BookmarkCheck className="h-4 w-4" />
                    ) : (
                      <Bookmark className="h-4 w-4" />
                    )}
                  </button>

                  <button
                    onClick={() => handleTrackApplication(job, 'APPLIED')}
                    disabled={appliedJobs[job.id]}
                    className={`px-4 py-2 rounded-xl text-xs font-semibold transition flex items-center space-x-1.5 shadow-md ${
                      appliedJobs[job.id]
                        ? 'bg-emerald-600/20 text-emerald-400 border border-emerald-500/30 cursor-default'
                        : 'bg-indigo-600 hover:bg-indigo-500 text-white shadow-indigo-600/20'
                    }`}
                  >
                    <Send className="h-3.5 w-3.5" />
                    <span>{appliedJobs[job.id] ? 'Tracked ✓' : 'Apply & Track'}</span>
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Opportunity Detail Modal */}
      {activeModalJob && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-3xl max-w-2xl w-full p-6 space-y-5 shadow-2xl animate-in zoom-in-95 duration-100 max-h-[90vh] overflow-y-auto">
            <div className="flex items-start justify-between">
              <div className="flex items-center space-x-4">
                {activeModalJob.company_logo_url ? (
                  <img src={activeModalJob.company_logo_url} alt={activeModalJob.company_name} className="h-14 w-14 rounded-2xl object-cover bg-slate-800 border border-slate-700" />
                ) : (
                  <div aria-hidden className="h-14 w-14 rounded-2xl shrink-0 bg-slate-800 border border-slate-700 flex items-center justify-center font-bold text-slate-300">
                    {activeModalJob.company_name.charAt(0)}
                  </div>
                )}
                <div>
                  <h2 className="text-xl font-extrabold text-white">{activeModalJob.title}</h2>
                  <div className="text-sm font-semibold text-slate-400">
                    {activeModalJob.company_name}
                  </div>
                </div>
              </div>
              {activeModalJob.match_score && (
                <span className="px-3.5 py-1.5 rounded-full text-xs font-black bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  {activeModalJob.match_score}% Match
                </span>
              )}
            </div>

            <div className="flex flex-wrap gap-2 text-xs">
              <span className="px-3 py-1 rounded-xl bg-slate-800 text-slate-300">
                {activeModalJob.location}
              </span>
              <span className="px-3 py-1 rounded-xl bg-slate-800 text-slate-300">
                {activeModalJob.stipend_or_salary}
              </span>
              <span className="px-3 py-1 rounded-xl bg-slate-800 text-slate-300">
                {activeModalJob.job_type}
              </span>
            </div>

            <div>
              <h4 className="text-xs font-bold uppercase tracking-wider text-slate-400 mb-2">
                Detailed Match Breakdown
              </h4>
              <div className="p-4 rounded-2xl bg-slate-950 border border-slate-800 space-y-2">
                {activeModalJob.matched_reasons?.map((r, i) => (
                  <div key={i} className="flex items-center space-x-2 text-xs text-emerald-400">
                    <CheckCircle2 className="h-4 w-4 shrink-0" />
                    <span>{r}</span>
                  </div>
                ))}
                {activeModalJob.missing_requirements?.map((m, i) => (
                  <div key={i} className="flex items-center space-x-2 text-xs text-amber-400">
                    <AlertCircle className="h-4 w-4 shrink-0" />
                    <span>{m}</span>
                  </div>
                ))}
              </div>
            </div>

            <div>
              <h4 className="text-xs font-bold uppercase tracking-wider text-slate-400 mb-1.5">
                Role Description
              </h4>
              <p className="text-xs text-slate-300 leading-relaxed whitespace-pre-line">
                {activeModalJob.description}
              </p>
            </div>

            <div className="flex items-center justify-end space-x-3 pt-4 border-t border-slate-800">
              <button
                onClick={() => setActiveModalJob(null)}
                className="px-4 py-2 rounded-xl text-xs font-semibold text-slate-400 hover:text-white"
              >
                Close
              </button>
              <a
                href={activeModalJob.source_url}
                target="_blank"
                rel="noreferrer"
                className="px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold flex items-center space-x-1.5"
              >
                <span>Apply on Official Portal</span>
                <ExternalLink className="h-3.5 w-3.5" />
              </a>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
