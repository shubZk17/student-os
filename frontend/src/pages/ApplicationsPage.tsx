import React, { useState, useEffect } from 'react';
import {
  Plus,
  Calendar,
  Clock,
  Trash2,
  Building,
  MoreVertical,
} from 'lucide-react';
import { apiRequest } from '../api/client';
import { Application, ApplicationStage } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

const STAGES: { key: ApplicationStage; label: string; color: string }[] = [
  { key: 'SAVED', label: 'Saved', color: 'border-slate-700' },
  { key: 'APPLIED', label: 'Applied', color: 'border-blue-500/40' },
  { key: 'ASSESSMENT', label: 'Assessment', color: 'border-amber-500/40' },
  { key: 'INTERVIEW', label: 'Interview', color: 'border-brand-500/40' },
  { key: 'OFFER', label: 'Offer', color: 'border-emerald-500/40' },
  { key: 'REJECTED', label: 'Rejected', color: 'border-rose-500/40' },
];

export const ApplicationsPage: React.FC = () => {
  const [applications, setApplications] = useState<Application[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingNotesApp, setEditingNotesApp] = useState<Application | null>(null);
  const [noteText, setNoteText] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    fetchApplications();
  }, []);

  const fetchApplications = () => {
    setLoading(true);
    apiRequest<{ applications: Application[] }>('/applications')
      .then((res) => {
        setApplications(res.applications || []);
        setLoading(false);
      })
      .catch((err: Error) => {
        setError(err.message);
        setLoading(false);
      });
  };

  // Optimistic update; on failure, show the error and reload the server state.
  const mutate = async (optimistic: (prev: Application[]) => Application[], endpoint: string, init: RequestInit) => {
    setApplications(optimistic);
    try {
      await apiRequest(endpoint, init);
    } catch (err: any) {
      setError(err.message);
      fetchApplications();
    }
  };

  const handleStageChange = (appId: string, newStage: ApplicationStage) =>
    mutate(
      (prev) => prev.map((a) => (a.id === appId ? { ...a, stage: newStage } : a)),
      `/applications/${appId}`,
      { method: 'PATCH', body: JSON.stringify({ stage: newStage }) }
    );

  const handleDelete = (appId: string) =>
    mutate((prev) => prev.filter((a) => a.id !== appId), `/applications/${appId}`, { method: 'DELETE' });

  const handleSaveNotes = async () => {
    if (!editingNotesApp) return;
    const { id: appId, stage } = editingNotesApp;
    setEditingNotesApp(null);
    await mutate(
      (prev) => prev.map((a) => (a.id === appId ? { ...a, notes: noteText } : a)),
      `/applications/${appId}`,
      { method: 'PATCH', body: JSON.stringify({ stage, notes: noteText }) }
    );
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-white tracking-tight">Application Tracker</h1>
        <p className="text-xs text-slate-400 mt-1">
          Keep track of your active job pipeline from discovery to offer. Drag or update stages with one click.
        </p>
      </div>

      {error && <ErrorBanner message={error} onDismiss={() => setError('')} />}

      {/* Kanban Board Container */}
      <div className="flex gap-4 overflow-x-auto pb-6 pt-2">
        {STAGES.map((col) => {
          const colApps = applications.filter((a) => a.stage === col.key);
          return (
            <div
              key={col.key}
              className="flex-shrink-0 w-72 bg-slate-900/50 rounded-3xl border border-slate-800 p-4 flex flex-col max-h-[78vh]"
            >
              {/* Column Header */}
              <div className="flex items-center justify-between pb-3 mb-3 border-b border-slate-800">
                <div className="flex items-center space-x-2">
                  <span className="font-semibold text-xs text-white">{col.label}</span>
                  <span className="h-5 w-5 rounded-full bg-slate-800 text-2xs font-semibold text-slate-300 flex items-center justify-center">
                    {colApps.length}
                  </span>
                </div>
              </div>

              {/* Column Cards */}
              <div className="space-y-3 overflow-y-auto flex-1 pr-1">
                {colApps.length === 0 ? (
                  <div className="text-center py-8 border border-dashed border-slate-800/80 rounded-2xl">
                    <p className="text-2xs text-slate-500">No applications</p>
                  </div>
                ) : (
                  colApps.map((app) => (
                    <div
                      key={app.id}
                      className="p-3.5 rounded-2xl bg-slate-950 border border-slate-800/90 hover:border-brand-500/40 transition shadow-sm space-y-2.5"
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="flex items-center space-x-2.5">
                          {app.company_logo_url ? (
                            <img
                              src={app.company_logo_url}
                              alt={app.company_name}
                              className="h-7 w-7 rounded-xl object-cover bg-slate-800 border border-slate-700"
                            />
                          ) : (
                            <div className="h-7 w-7 rounded-xl bg-slate-800 flex items-center justify-center text-slate-400">
                              <Building className="h-4 w-4" />
                            </div>
                          )}
                          <div>
                            <h4 className="font-semibold text-xs text-white leading-tight line-clamp-1">
                              {app.title}
                            </h4>
                            <span className="text-2xs text-slate-400">{app.company_name}</span>
                          </div>
                        </div>

                        <button
                          onClick={() => handleDelete(app.id)}
                          className="text-slate-500 hover:text-rose-400 p-1 rounded-lg transition"
                          title="Remove from board"
                        >
                          <Trash2 className="h-3 w-3" />
                        </button>
                      </div>

                      {/* Interview Date Pill */}
                      {app.interview_date && (
                        <div className="flex items-center space-x-1 text-2xs text-slate-300 bg-slate-800 border border-slate-700 px-2 py-1 rounded-md">
                          <Calendar className="h-3 w-3" />
                          <span>Interview: {new Date(app.interview_date).toLocaleDateString()}</span>
                        </div>
                      )}

                      {/* Notes snippet */}
                      {app.notes && (
                        <p
                          onClick={() => {
                            setEditingNotesApp(app);
                            setNoteText(app.notes);
                          }}
                          className="text-2xs text-slate-400 bg-slate-900/60 p-2 rounded-xl border border-slate-800 cursor-pointer hover:border-slate-700 transition line-clamp-2"
                        >
                          {app.notes}
                        </p>
                      )}

                      {/* Move to Stage Selector */}
                      <div className="pt-2 border-t border-slate-900 flex items-center justify-between text-2xs">
                        <button
                          onClick={() => {
                            setEditingNotesApp(app);
                            setNoteText(app.notes || '');
                          }}
                          className="text-brand-400 hover:text-slate-300 font-medium"
                        >
                          {app.notes ? 'Edit note' : '+ Add note'}
                        </button>

                        <select
                          value={app.stage}
                          onChange={(e) => handleStageChange(app.id, e.target.value as ApplicationStage)}
                          className="bg-slate-900 border border-slate-800 rounded-lg px-2 py-0.5 text-2xs text-slate-300 focus:outline-none"
                        >
                          {STAGES.map((s) => (
                            <option key={s.key} value={s.key}>
                              Move: {s.label}
                            </option>
                          ))}
                        </select>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          );
        })}
      </div>

      {/* Edit Note Modal */}
      {editingNotesApp && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-3xl max-w-md w-full p-6 space-y-4 shadow-lg">
            <h3 className="text-sm font-semibold text-white">
              Application Notes: {editingNotesApp.company_name}
            </h3>
            <textarea
              rows={4}
              value={noteText}
              onChange={(e) => setNoteText(e.target.value)}
              placeholder="Record interview notes, technical questions asked, next steps..."
              className="w-full bg-slate-950 border border-slate-800 rounded-2xl p-3 text-xs text-slate-200 focus:outline-none focus:border-brand-500"
            />
            <div className="flex items-center justify-end space-x-3">
              <button
                onClick={() => setEditingNotesApp(null)}
                className="px-3 py-1.5 rounded-xl text-xs text-slate-400 hover:text-white"
              >
                Cancel
              </button>
              <button
                onClick={handleSaveNotes}
                className="px-4 py-1.5 rounded-xl bg-brand-600 hover:bg-brand-500 text-white text-xs font-semibold"
              >
                Save Notes
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
