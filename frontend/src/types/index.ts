export interface User {
  id: string;
  email: string;
  full_name: string;
  role: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
}

export interface Skill {
  id: string;
  name: string;
  category: string;
  proficiency?: string;
}

export interface StudentProfile {
  id: string;
  user_id: string;
  full_name: string;
  avatar_url?: string;
  college: string;
  degree: string;
  branch: string;
  current_year: number;
  graduation_year: number;
  cgpa: number;
  target_roles: string[];
  preferred_locations: string[];
  work_preference: string;
  bio: string;
  github_url: string;
  linkedin_url: string;
  portfolio_url: string;
  resume_url: string;
  profile_strength: number;
  skills: Skill[];
}

export interface Opportunity {
  id: string;
  external_id?: string;
  source: string;
  source_url: string;
  type: 'JOB' | 'INTERNSHIP' | 'HACKATHON' | 'RESEARCH' | 'SCHOLARSHIP' | string;
  company_name: string;
  company_logo_url: string;
  title: string;
  description: string;
  location: string;
  is_remote: boolean;
  job_type: 'INTERNSHIP' | 'FULL_TIME' | 'CONTRACT' | string;
  experience_level: string;
  required_skills: string[];
  eligible_degrees: string[];
  eligible_grad_years: number[];
  min_cgpa: number;
  stipend_or_salary: string;
  deadline?: string;
  posted_at: string;
  status: string;

  // Matching enrichments
  match_score?: number;
  matched_reasons?: string[];
  missing_requirements?: string[];
  is_saved?: boolean;
}

export type ApplicationStage =
  | 'SAVED'
  | 'APPLIED'
  | 'ASSESSMENT'
  | 'INTERVIEW'
  | 'OFFER'
  | 'REJECTED'
  | 'WITHDRAWN';

export interface Application {
  id: string;
  user_id: string;
  opportunity_id: string;
  company_name: string;
  title: string;
  stage: ApplicationStage;
  applied_at?: string;
  interview_date?: string;
  next_follow_up?: string;
  notes: string;
  location: string;
  company_logo_url: string;
  created_at: string;
  updated_at: string;
}

export interface Project {
  id: string;
  student_id: string;
  title: string;
  tagline: string;
  description: string;
  role: string;
  github_url?: string;
  demo_url?: string;
  start_date?: string;
  end_date?: string;
  metrics: string;
  technologies: string[];
  created_at: string;
}

export interface NotificationItem {
  id: string;
  user_id: string;
  title: string;
  message: string;
  type: 'HIGH_MATCH' | 'DEADLINE_SOON' | 'INTERVIEW_ALERT' | 'PROFILE_TIP';
  link_url?: string;
  is_read: boolean;
  created_at: string;
}

export interface DashboardSummary {
  student_name: string;
  profile_strength: number;
  active_applications_count: number;
  upcoming_interviews_count: number;
  impending_deadlines_count: number;
  upcoming_interview?: {
    company: string;
    title: string;
    date: string;
  };
}
