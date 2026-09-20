-- StudentOS DEV-ONLY seed data. Never run against production.
-- Load with `make seed-dev` after the backend has started once (it applies migrations).
-- Standard Tech Skills, Realistic Opportunities, and Demo Data

-- Core skills are reference data: see backend/migrations/000003_reference_skills.up.sql

-- 2. Insert Realistic Sample Opportunities (Internships, Jobs, Hackathons)
INSERT INTO opportunities (
    id, external_id, source, source_url, type, company_name, company_logo_url,
    title, description, location, is_remote, job_type, experience_level,
    eligible_degrees, eligible_grad_years, min_cgpa, stipend_or_salary,
    deadline, status
) VALUES
(
    'b0000000-0000-0000-0000-000000000001',
    'gh-sw-101',
    'greenhouse',
    'https://careers.example.com/ml-intern',
    'INTERNSHIP',
    'CognitiveScale Labs',
    'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=128&auto=format&fit=crop&q=60',
    'Machine Learning Engineer Intern',
    'Join our Applied AI research group to develop and evaluate high-throughput LLM pipelines, fine-tune transformer models with PyTorch, and deploy low-latency inference services.',
    'Bangalore, India',
    false,
    'INTERNSHIP',
    'INTERN',
    ARRAY['B.Tech', 'B.E.', 'M.Tech', 'BCA', 'MCA'],
    ARRAY[2025, 2026, 2027],
    7.5,
    '₹45,000 - ₹60,000 / month',
    NOW() + INTERVAL '21 days',
    'ACTIVE'
),
(
    'b0000000-0000-0000-0000-000000000002',
    'lever-sw-202',
    'lever',
    'https://careers.example.com/swe-intern',
    'INTERNSHIP',
    'HyperRoute Cloud',
    'https://images.unsplash.com/photo-1551288049-bebda4e38f71?w=128&auto=format&fit=crop&q=60',
    'Software Engineer Intern (Backend & Go)',
    'Work alongside senior systems engineers to design scalable distributed microservices in Go, optimize PostgreSQL read-replicas, and orchestrate containerized workloads in AWS.',
    'Hyderabad, India',
    true,
    'INTERNSHIP',
    'INTERN',
    ARRAY['B.Tech', 'B.E.', 'BCA', 'MCA', 'B.Sc Computer Science'],
    ARRAY[2025, 2026],
    7.0,
    '₹50,000 / month',
    NOW() + INTERVAL '14 days',
    'ACTIVE'
),
(
    'b0000000-0000-0000-0000-000000000003',
    'ashby-sw-303',
    'ashby',
    'https://careers.example.com/frontend-intern',
    'INTERNSHIP',
    'CanvasFlow Studio',
    'https://images.unsplash.com/photo-1507238691740-187a5b1d37b8?w=128&auto=format&fit=crop&q=60',
    'Frontend Developer Intern (React & TypeScript)',
    'Build delightful, highly responsive client experiences using modern React 19, TypeScript, and Tailwind CSS. Collaborate closely with product designers and backend engineers.',
    'Pune, India',
    true,
    'INTERNSHIP',
    'INTERN',
    ARRAY['B.Tech', 'B.E.', 'BCA', 'MCA', 'Any Degree'],
    ARRAY[2025, 2026, 2027],
    6.5,
    '₹35,000 / month',
    NOW() + INTERVAL '10 days',
    'ACTIVE'
),
(
    'b0000000-0000-0000-0000-000000000004',
    'campus-hack-404',
    'campus',
    'https://studentos.dev/hackathon-2026',
    'HACKATHON',
    'National Open Innovation Conclave',
    'https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?w=128&auto=format&fit=crop&q=60',
    'Smart India Techathon 2026',
    '48-hour national student hackathon challenging college innovators to solve pressing public-sector and enterprise challenges in AI, FinTech, Sustainability, and HealthTech.',
    'New Delhi / Hybrid',
    true,
    'CONTRACT',
    'ENTRY',
    ARRAY['Any Degree'],
    ARRAY[2024, 2025, 2026, 2027, 2028],
    0.0,
    '₹5,00,000 Total Prize Pool',
    NOW() + INTERVAL '30 days',
    'ACTIVE'
)
ON CONFLICT (source, external_id) DO NOTHING;

-- 3. Link Opportunities to Required Skills
-- For CognitiveScale Labs (ML Intern)
INSERT INTO opportunity_skills (opportunity_id, skill_id, is_required) VALUES
    ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', true), -- Python
    ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000002', true), -- PyTorch
    ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000003', true), -- Machine Learning
    ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000008', false) -- Docker
ON CONFLICT DO NOTHING;

-- For HyperRoute Cloud (Backend Go Intern)
INSERT INTO opportunity_skills (opportunity_id, skill_id, is_required) VALUES
    ('b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000004', true), -- Go
    ('b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000007', true), -- PostgreSQL
    ('b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000008', true), -- Docker
    ('b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000009', false) -- AWS
ON CONFLICT DO NOTHING;

-- For CanvasFlow Studio (Frontend Intern)
INSERT INTO opportunity_skills (opportunity_id, skill_id, is_required) VALUES
    ('b0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000005', true), -- React.js
    ('b0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000006', true), -- TypeScript
    ('b0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000012', true)  -- Tailwind CSS
ON CONFLICT DO NOTHING;
