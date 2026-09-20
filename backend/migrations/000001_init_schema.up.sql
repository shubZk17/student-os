-- StudentOS Initial Database Schema
-- Supports PostgreSQL 16 with pgvector extension

-- gen_random_uuid() is built into PostgreSQL 13+, so no uuid extension is needed.
CREATE EXTENSION IF NOT EXISTS "vector";

-- Users Table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'student',
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Refresh Tokens Table (Rotation & Revocation)
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);

-- Student Profiles Table
CREATE TABLE IF NOT EXISTS student_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    full_name VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    college VARCHAR(255),
    degree VARCHAR(128),
    branch VARCHAR(128),
    current_year INT CHECK (current_year BETWEEN 1 AND 5),
    graduation_year INT NOT NULL,
    cgpa NUMERIC(4,2) DEFAULT 0.0,
    target_roles TEXT[] DEFAULT '{}',
    preferred_locations TEXT[] DEFAULT '{}',
    work_preference VARCHAR(32) DEFAULT 'ANY',
    bio TEXT,
    github_url TEXT,
    linkedin_url TEXT,
    portfolio_url TEXT,
    resume_url TEXT,
    profile_strength INT NOT NULL DEFAULT 20,
    embedding vector(384),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_student_profiles_user ON student_profiles(user_id);

-- Skills Master & Junction
CREATE TABLE IF NOT EXISTS skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    category VARCHAR(64) DEFAULT 'General'
);

CREATE TABLE IF NOT EXISTS student_skills (
    student_id UUID NOT NULL REFERENCES student_profiles(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    proficiency VARCHAR(32) DEFAULT 'INTERMEDIATE',
    PRIMARY KEY (student_id, skill_id)
);

-- Opportunities / Jobs Polymorphic Table
CREATE TABLE IF NOT EXISTS opportunities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id VARCHAR(255),
    source VARCHAR(64) NOT NULL,
    source_url TEXT NOT NULL,
    type VARCHAR(32) NOT NULL DEFAULT 'JOB',
    company_name VARCHAR(255) NOT NULL,
    company_logo_url TEXT,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    location VARCHAR(255) NOT NULL DEFAULT 'Remote',
    is_remote BOOLEAN NOT NULL DEFAULT FALSE,
    job_type VARCHAR(32) DEFAULT 'FULL_TIME',
    experience_level VARCHAR(64) DEFAULT 'ENTRY',
    eligible_degrees TEXT[] DEFAULT '{}',
    eligible_grad_years INT[] DEFAULT '{}',
    min_cgpa NUMERIC(4,2) DEFAULT 0.0,
    stipend_or_salary VARCHAR(128),
    deadline TIMESTAMPTZ,
    posted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    embedding vector(384),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_source_external UNIQUE (source, external_id)
);
CREATE INDEX IF NOT EXISTS idx_opportunities_status_deadline ON opportunities(status, deadline);
CREATE INDEX IF NOT EXISTS idx_opportunities_type ON opportunities(type);

-- Opportunity Required Skills
CREATE TABLE IF NOT EXISTS opportunity_skills (
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    is_required BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (opportunity_id, skill_id)
);

-- Applications (Kanban Tracking)
CREATE TABLE IF NOT EXISTS applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    stage VARCHAR(32) NOT NULL DEFAULT 'SAVED',
    applied_at TIMESTAMPTZ,
    interview_date TIMESTAMPTZ,
    next_follow_up TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_opportunity UNIQUE (user_id, opportunity_id)
);
CREATE INDEX IF NOT EXISTS idx_applications_user_stage ON applications(user_id, stage);

-- Student Projects & Portfolio
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES student_profiles(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    tagline VARCHAR(255),
    description TEXT NOT NULL,
    role VARCHAR(128) DEFAULT 'Creator / Developer',
    github_url TEXT,
    demo_url TEXT,
    start_date DATE,
    end_date DATE,
    metrics TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_projects_student ON projects(student_id);

CREATE TABLE IF NOT EXISTS project_skills (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    PRIMARY KEY (project_id, skill_id)
);

-- Notifications
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(64) NOT NULL,
    link_url TEXT,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id, is_read);
