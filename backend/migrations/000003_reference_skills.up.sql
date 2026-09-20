-- Reference skills. Fixed IDs are relied on by seeds/dev_seed.sql.
-- The job worker tags postings only with names in this table (users add more via their profiles).
INSERT INTO skills (id, name, category) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'Python', 'Programming Language'),
    ('a0000000-0000-0000-0000-000000000002', 'PyTorch', 'Machine Learning'),
    ('a0000000-0000-0000-0000-000000000003', 'Machine Learning', 'Data Science'),
    ('a0000000-0000-0000-0000-000000000004', 'Go (Golang)', 'Programming Language'),
    ('a0000000-0000-0000-0000-000000000005', 'React.js', 'Frontend Development'),
    ('a0000000-0000-0000-0000-000000000006', 'TypeScript', 'Frontend Development'),
    ('a0000000-0000-0000-0000-000000000007', 'PostgreSQL', 'Databases'),
    ('a0000000-0000-0000-0000-000000000008', 'Docker', 'Cloud & DevOps'),
    ('a0000000-0000-0000-0000-000000000009', 'AWS', 'Cloud & DevOps'),
    ('a0000000-0000-0000-0000-000000000010', 'Data Structures & Algorithms', 'Computer Science'),
    ('a0000000-0000-0000-0000-000000000011', 'FastAPI', 'Backend Development'),
    ('a0000000-0000-0000-0000-000000000012', 'Tailwind CSS', 'Frontend Development'),
    ('a0000000-0000-0000-0000-000000000013', 'Git', 'Version Control'),
    ('a0000000-0000-0000-0000-000000000014', 'C++', 'Programming Language'),
    ('a0000000-0000-0000-0000-000000000015', 'Kubernetes', 'Cloud & DevOps')
ON CONFLICT (name) DO NOTHING;

INSERT INTO skills (name, category) VALUES
    ('Java', 'Programming Language'),
    ('JavaScript', 'Programming Language'),
    ('Rust', 'Programming Language'),
    ('SQL', 'Databases'),
    ('Node.js', 'Backend Development'),
    ('Linux', 'Cloud & DevOps'),
    ('Terraform', 'Cloud & DevOps'),
    ('Kotlin', 'Programming Language'),
    ('Swift', 'Programming Language'),
    ('Spark', 'Data Engineering')
ON CONFLICT (name) DO NOTHING;
