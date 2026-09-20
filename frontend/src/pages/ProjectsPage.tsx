import React, { useState, useEffect } from 'react';
import {
  FolderGit2,
  Plus,
  Github,
  ExternalLink,
  Trash2,
  TrendingUp,
  Code2,
} from 'lucide-react';
import { apiRequest } from '../api/client';
import { Project } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

export const ProjectsPage: React.FC = () => {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAddModal, setShowAddModal] = useState(false);
  const [error, setError] = useState('');

  // Form State
  const [title, setTitle] = useState('');
  const [tagline, setTagline] = useState('');
  const [description, setDescription] = useState('');
  const [role, setRole] = useState('Lead Developer');
  const [githubUrl, setGithubUrl] = useState('');
  const [demoUrl, setDemoUrl] = useState('');
  const [metrics, setMetrics] = useState('');
  const [techInput, setTechInput] = useState('');

  useEffect(() => {
    fetchProjects();
  }, []);

  const fetchProjects = () => {
    setLoading(true);
    apiRequest<{ projects: Project[] }>('/projects')
      .then((res) => {
        setProjects(res.projects || []);
        setLoading(false);
      })
      .catch((err: Error) => {
        setError(err.message);
        setLoading(false);
      });
  };

  const handleCreateProject = async (e: React.FormEvent) => {
    e.preventDefault();
    const techArray = techInput
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean);

    try {
      const res = await apiRequest<{ project: Project }>('/projects', {
        method: 'POST',
        body: JSON.stringify({
          title,
          tagline,
          description,
          role,
          github_url: githubUrl,
          demo_url: demoUrl,
          metrics,
          technologies: techArray,
        }),
      });
      setProjects((prev) => [res.project, ...prev]);
    } catch (err: any) {
      setError(err.message);
      return; // keep the modal and form open so the user can fix and retry
    }

    setShowAddModal(false);
    setTitle('');
    setTagline('');
    setDescription('');
    setGithubUrl('');
    setDemoUrl('');
    setMetrics('');
    setTechInput('');
  };

  const handleDelete = async (id: string) => {
    try {
      await apiRequest(`/projects/${id}`, { method: 'DELETE' });
      setProjects((prev) => prev.filter((p) => p.id !== id));
    } catch (err: any) {
      setError(err.message);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold text-white tracking-tight">
            Projects & Portfolio
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Projects validate your hands-on expertise and dynamically boost your match score across relevant opportunities.
          </p>
        </div>

        <button
          onClick={() => setShowAddModal(true)}
          className="px-4 py-2.5 rounded-xl bg-brand-600 hover:bg-brand-500 text-white text-xs font-semibold flex items-center space-x-2 transition shadow-lg"
        >
          <Plus className="h-4 w-4" />
          <span>Add Project</span>
        </button>
      </div>

      {error && !showAddModal && <ErrorBanner message={error} onDismiss={() => setError('')} />}

      {/* Project Grid */}
      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-5 animate-pulse">
          <div className="h-60 bg-slate-800/40 rounded-3xl"></div>
          <div className="h-60 bg-slate-800/40 rounded-3xl"></div>
        </div>
      ) : projects.length === 0 ? (
        <div className="p-12 text-center rounded-3xl bg-slate-900/40 border border-slate-800">
          <div className="h-12 w-12 rounded-2xl bg-slate-800 flex items-center justify-center mx-auto text-slate-500 mb-3">
            <FolderGit2 className="h-6 w-6" />
          </div>
          <h3 className="font-semibold text-sm text-white">No projects added yet</h3>
          <p className="text-xs text-slate-400 max-w-sm mx-auto mt-1">
            Showcase side projects, research publications, or hackathon winners to demonstrate verified skills.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
          {projects.map((p) => (
            <div
              key={p.id}
              className="p-6 rounded-3xl bg-slate-900/70 border border-slate-800 hover:border-brand-500/40 transition flex flex-col justify-between shadow-sm"
            >
              <div>
                <div className="flex items-start justify-between gap-3 mb-2">
                  <div>
                    <h2 className="text-base font-semibold text-white hover:text-slate-300 transition">
                      {p.title}
                    </h2>
                    {p.tagline && <p className="text-xs text-slate-400 mt-0.5">{p.tagline}</p>}
                  </div>
                  <button
                    onClick={() => handleDelete(p.id)}
                    className="text-slate-500 hover:text-rose-400 p-1 transition"
                    title="Delete project"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>

                <p className="text-xs text-slate-300 line-clamp-3 mb-4 leading-relaxed">
                  {p.description}
                </p>

                {/* Measurable Metrics Highlight */}
                {p.metrics && (
                  <div className="p-3 rounded-2xl bg-slate-900 border border-slate-800 mb-4 flex items-center space-x-2 text-xs text-slate-300">
                    <TrendingUp className="h-4 w-4 text-brand-400 shrink-0" />
                    <span className="font-medium text-2xs">{p.metrics}</span>
                  </div>
                )}

                {/* Technologies / Skill Badges */}
                <div className="flex flex-wrap gap-1.5 mb-4">
                  {p.technologies?.map((tech, i) => (
                    <span
                      key={i}
                      className="px-2.5 py-1 rounded-lg text-2xs font-semibold bg-slate-800/80 text-slate-300 border border-slate-700/60"
                    >
                      {tech}
                    </span>
                  ))}
                </div>
              </div>

              {/* Action Links */}
              <div className="flex items-center space-x-4 pt-4 border-t border-slate-800 text-xs">
                {p.github_url && (
                  <a
                    href={p.github_url}
                    target="_blank"
                    rel="noreferrer"
                    className="text-slate-400 hover:text-white flex items-center space-x-1.5 transition"
                  >
                    <Github className="h-4 w-4" />
                    <span>Source Code</span>
                  </a>
                )}
                {p.demo_url && (
                  <a
                    href={p.demo_url}
                    target="_blank"
                    rel="noreferrer"
                    className="text-brand-400 hover:text-slate-300 flex items-center space-x-1.5 transition"
                  >
                    <ExternalLink className="h-4 w-4" />
                    <span>Live Demo</span>
                  </a>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Add Project Modal */}
      {showAddModal && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <form
            onSubmit={handleCreateProject}
            className="bg-slate-900 border border-slate-800 rounded-3xl max-w-lg w-full p-6 space-y-4 shadow-lg max-h-[90vh] overflow-y-auto"
          >
            <h3 className="text-base font-semibold text-white">Add Portfolio Project</h3>
            {error && <ErrorBanner message={error} onDismiss={() => setError('')} />}

            <div>
              <label className="block text-xs font-semibold text-slate-400 mb-1">Project Title *</label>
              <input
                type="text"
                required
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="e.g. Distributed Cache Engine"
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-brand-500"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-400 mb-1">One-line Tagline</label>
              <input
                type="text"
                value={tagline}
                onChange={(e) => setTagline(e.target.value)}
                placeholder="e.g. High-throughput in-memory LRU key-value store in Go"
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-brand-500"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-400 mb-1">Description *</label>
              <textarea
                required
                rows={3}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Explain the problem solved, architecture choices, and impact..."
                className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-xs text-white focus:outline-none focus:border-brand-500"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-400 mb-1">Technologies (Comma separated) *</label>
              <input
                type="text"
                required
                value={techInput}
                onChange={(e) => setTechInput(e.target.value)}
                placeholder="e.g. Go (Golang), Docker, PyTorch, PostgreSQL"
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-brand-500"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-400 mb-1">Impact Metrics</label>
              <input
                type="text"
                value={metrics}
                onChange={(e) => setMetrics(e.target.value)}
                placeholder="e.g. 120k QPS, 94.2% accuracy, 500+ active users"
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-brand-500"
              />
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-semibold text-slate-400 mb-1">GitHub Repo URL</label>
                <input
                  type="url"
                  value={githubUrl}
                  onChange={(e) => setGithubUrl(e.target.value)}
                  placeholder="https://github.com/..."
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-brand-500"
                />
              </div>
              <div>
                <label className="block text-xs font-semibold text-slate-400 mb-1">Live Demo URL</label>
                <input
                  type="url"
                  value={demoUrl}
                  onChange={(e) => setDemoUrl(e.target.value)}
                  placeholder="https://..."
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-brand-500"
                />
              </div>
            </div>

            <div className="flex items-center justify-end space-x-3 pt-3 border-t border-slate-800">
              <button
                type="button"
                onClick={() => setShowAddModal(false)}
                className="px-4 py-2 rounded-xl text-xs text-slate-400 hover:text-white"
              >
                Cancel
              </button>
              <button
                type="submit"
                className="px-5 py-2 rounded-xl bg-brand-600 hover:bg-brand-500 text-white text-xs font-semibold"
              >
                Save Project
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
};
