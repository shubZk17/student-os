import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import {
  User,
  GraduationCap,
  Briefcase,
  MapPin,
  Github,
  Linkedin,
  FileText,
  Sparkles,
  Save,
  CheckCircle2,
  Plus,
  X,
} from 'lucide-react';
import { apiRequest } from '../api/client';
import { StudentProfile } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

export const ProfilePage: React.FC = () => {
  const { user } = useAuth();
  const [profile, setProfile] = useState<StudentProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [savedSuccess, setSavedSuccess] = useState(false);
  const [newSkillInput, setNewSkillInput] = useState('');

  // Form State
  const [fullName, setFullName] = useState('');
  const [college, setCollege] = useState('');
  const [degree, setDegree] = useState('');
  const [branch, setBranch] = useState('');
  const [currentYear, setCurrentYear] = useState(3);
  const [gradYear, setGradYear] = useState(2026);
  const [cgpa, setCgpa] = useState(0);
  const [targetRoles, setTargetRoles] = useState('');
  const [preferredLocations, setPreferredLocations] = useState('');
  const [workPreference, setWorkPreference] = useState('ANY');
  const [bio, setBio] = useState('');
  const [githubUrl, setGithubUrl] = useState('');
  const [linkedinUrl, setLinkedinUrl] = useState('');
  const [skillsList, setSkillsList] = useState<string[]>([]);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchProfile();
  }, [user]);

  const fetchProfile = () => {
    setLoading(true);
    apiRequest<{ profile: StudentProfile }>('/profile')
      .then((res) => {
        const p = res.profile;
        setProfile(p);
        setFullName(p.full_name || '');
        setCollege(p.college || '');
        setDegree(p.degree || '');
        setBranch(p.branch || '');
        setCurrentYear(p.current_year || 3);
        setGradYear(p.graduation_year || 2026);
        setCgpa(p.cgpa || 0);
        setTargetRoles(p.target_roles?.join(', ') || '');
        setPreferredLocations(p.preferred_locations?.join(', ') || '');
        setWorkPreference(p.work_preference || 'ANY');
        setBio(p.bio || '');
        setGithubUrl(p.github_url || '');
        setLinkedinUrl(p.linkedin_url || '');
        setSkillsList((p.skills || []).map((s) => s.name));
        setLoading(false);
      })
      .catch((err: Error) => {
        setError(err.message);
        setLoading(false);
      });
  };

  const handleAddSkill = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && newSkillInput.trim()) {
      e.preventDefault();
      if (!skillsList.includes(newSkillInput.trim())) {
        setSkillsList([...skillsList, newSkillInput.trim()]);
      }
      setNewSkillInput('');
    }
  };

  const handleRemoveSkill = (skill: string) => {
    setSkillsList(skillsList.filter((s) => s !== skill));
  };

  const handleSaveProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      full_name: fullName,
      college,
      degree,
      branch,
      current_year: currentYear,
      graduation_year: gradYear,
      cgpa,
      target_roles: targetRoles.split(',').map((r) => r.trim()).filter(Boolean),
      preferred_locations: preferredLocations.split(',').map((l) => l.trim()).filter(Boolean),
      work_preference: workPreference,
      bio,
      github_url: githubUrl,
      linkedin_url: linkedinUrl,
      skills: skillsList,
    };

    setError('');
    try {
      await apiRequest('/profile', {
        method: 'PUT',
        body: JSON.stringify(payload),
      });
    } catch (err: any) {
      setError(err.message);
      return;
    }

    setSavedSuccess(true);
    setTimeout(() => setSavedSuccess(false), 3000);
  };

  const calculateStrength = () => {
    let score = 25;
    if (college && degree) score += 20;
    if (cgpa > 0) score += 10;
    if (skillsList.length >= 4) score += 15;
    if (githubUrl) score += 15;
    if (linkedinUrl) score += 15;
    return Math.min(100, score);
  };

  const strength = calculateStrength();

  return (
    <div className="space-y-6 max-w-4xl mx-auto">
      <div>
        <h1 className="text-2xl font-extrabold text-white tracking-tight">Student Profile</h1>
        <p className="text-xs text-slate-400 mt-1">
          Your profile directly powers the explainable matching algorithm. Keep it fresh with your latest coursework, skills, and links.
        </p>
      </div>

      {/* Profile Strength Bar */}
      <div className="p-5 rounded-3xl bg-slate-900/80 border border-slate-800 flex flex-col sm:flex-row items-center justify-between gap-4">
        <div className="space-y-1 w-full sm:w-auto">
          <div className="flex items-center space-x-2">
            <Sparkles className="h-4 w-4 text-indigo-400" />
            <span className="text-xs font-bold text-white uppercase tracking-wider">
              Profile Strength: {strength}%
            </span>
          </div>
          <p className="text-[11px] text-slate-400">
            {strength < 90 ? 'Add portfolio links and skills to unlock higher match scores.' : 'Excellent! Your profile is primed for 95%+ high-signal matches.'}
          </p>
        </div>

        <div className="w-full sm:w-64 bg-slate-950 h-3 rounded-full overflow-hidden border border-slate-800">
          <div
            className="h-full bg-gradient-to-r from-indigo-500 via-purple-500 to-emerald-400 rounded-full transition-all duration-700"
            style={{ width: `${strength}%` }}
          ></div>
        </div>
      </div>

      {/* Form */}
      {error && <ErrorBanner message={error} onDismiss={() => setError('')} />}

      <form onSubmit={handleSaveProfile} className="space-y-6">
        {/* Basic Academic Info */}
        <div className="p-6 rounded-3xl bg-slate-900/60 border border-slate-800 space-y-4">
          <h3 className="text-sm font-bold text-white flex items-center space-x-2 pb-2 border-b border-slate-800">
            <GraduationCap className="h-4 w-4 text-indigo-400" />
            <span>Academic Background</span>
          </h3>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold text-slate-400 mb-1">Full Name</label>
              <input
                type="text"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-400 mb-1">College / University</label>
              <input
                type="text"
                value={college}
                onChange={(e) => setCollege(e.target.value)}
                placeholder="e.g. Indian Institute of Technology"
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-400 mb-1">Degree</label>
              <input
                type="text"
                value={degree}
                onChange={(e) => setDegree(e.target.value)}
                placeholder="e.g. B.Tech, B.E., BCA"
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-400 mb-1">Branch / Major</label>
              <input
                type="text"
                value={branch}
                onChange={(e) => setBranch(e.target.value)}
                placeholder="e.g. Computer Science & Engineering"
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-400 mb-1">Graduation Batch Year</label>
              <input
                type="number"
                value={gradYear}
                onChange={(e) => setGradYear(Number(e.target.value))}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-400 mb-1">Current CGPA</label>
              <input
                type="number"
                step="0.1"
                value={cgpa}
                onChange={(e) => setCgpa(Number(e.target.value))}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>
        </div>

        {/* Skills Tag Management */}
        <div className="p-6 rounded-3xl bg-slate-900/60 border border-slate-800 space-y-4">
          <h3 className="text-sm font-bold text-white flex items-center space-x-2 pb-2 border-b border-slate-800">
            <Sparkles className="h-4 w-4 text-indigo-400" />
            <span>Verified Skills Graph</span>
          </h3>

          <div className="space-y-3">
            <div className="flex flex-wrap gap-2">
              {skillsList.map((skill) => (
                <span
                  key={skill}
                  className="px-3 py-1 rounded-xl text-xs font-semibold bg-indigo-950/50 text-indigo-300 border border-indigo-500/30 flex items-center space-x-1.5"
                >
                  <span>{skill}</span>
                  <button
                    type="button"
                    onClick={() => handleRemoveSkill(skill)}
                    className="hover:text-rose-400 transition"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </span>
              ))}
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="text"
                value={newSkillInput}
                onChange={(e) => setNewSkillInput(e.target.value)}
                onKeyDown={handleAddSkill}
                placeholder="Type a skill and press Enter (e.g. AWS, Kubernetes, FastAPI)..."
                className="flex-1 bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
              />
              <button
                type="button"
                onClick={() => {
                  if (newSkillInput.trim() && !skillsList.includes(newSkillInput.trim())) {
                    setSkillsList([...skillsList, newSkillInput.trim()]);
                    setNewSkillInput('');
                  }
                }}
                className="px-3 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-xs font-semibold text-white transition flex items-center space-x-1"
              >
                <Plus className="h-3.5 w-3.5" />
                <span>Add</span>
              </button>
            </div>
          </div>
        </div>

        {/* Roles and Preferences */}
        <div className="p-6 rounded-3xl bg-slate-900/60 border border-slate-800 space-y-4">
          <h3 className="text-sm font-bold text-white flex items-center space-x-2 pb-2 border-b border-slate-800">
            <Briefcase className="h-4 w-4 text-indigo-400" />
            <span>Target Roles & Preferences</span>
          </h3>

          <div className="space-y-4">
            <div>
              <label className="block text-xs font-semibold text-slate-400 mb-1">
                Target Roles (Comma separated)
              </label>
              <input
                type="text"
                value={targetRoles}
                onChange={(e) => setTargetRoles(e.target.value)}
                placeholder="e.g. Software Engineer, Machine Learning Intern, Cloud Engineer"
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
              />
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-semibold text-slate-400 mb-1">
                  Preferred Locations (Comma separated)
                </label>
                <input
                  type="text"
                  value={preferredLocations}
                  onChange={(e) => setPreferredLocations(e.target.value)}
                  placeholder="e.g. Bangalore, Hyderabad, Pune, Remote"
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-400 mb-1">Work Preference</label>
                <select
                  value={workPreference}
                  onChange={(e) => setWorkPreference(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
                >
                  <option value="ANY">Any (Remote or Onsite)</option>
                  <option value="REMOTE">Remote Only</option>
                  <option value="ONSITE">Onsite Preferred</option>
                </select>
              </div>
            </div>
          </div>
        </div>

        {/* Links & Socials */}
        <div className="p-6 rounded-3xl bg-slate-900/60 border border-slate-800 space-y-4">
          <h3 className="text-sm font-bold text-white flex items-center space-x-2 pb-2 border-b border-slate-800">
            <Github className="h-4 w-4 text-indigo-400" />
            <span>Online Presence & Links</span>
          </h3>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold text-slate-400 mb-1">GitHub Profile</label>
              <input
                type="url"
                value={githubUrl}
                onChange={(e) => setGithubUrl(e.target.value)}
                placeholder="https://github.com/..."
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-400 mb-1">LinkedIn Profile</label>
              <input
                type="url"
                value={linkedinUrl}
                onChange={(e) => setLinkedinUrl(e.target.value)}
                placeholder="https://linkedin.com/in/..."
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>
        </div>

        <div className="flex items-center justify-between pt-2">
          {savedSuccess && (
            <span className="text-xs text-emerald-400 font-semibold flex items-center space-x-1.5">
              <CheckCircle2 className="h-4 w-4" />
              <span>Profile updated successfully!</span>
            </span>
          )}
          <button
            type="submit"
            className="ml-auto px-6 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold flex items-center space-x-2 transition shadow-lg shadow-indigo-600/20"
          >
            <Save className="h-4 w-4" />
            <span>Save Profile</span>
          </button>
        </div>
      </form>
    </div>
  );
};
