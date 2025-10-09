-- Performance optimization indexes for Bruno site system
-- Migration: 002_performance_indexes.sql
-- Implements Golden Rule #1: Database Index

-- Composite indexes for frequently queried combinations
CREATE INDEX IF NOT EXISTS idx_projects_active_featured_order ON projects(active, featured, "order") WHERE active = TRUE;
CREATE INDEX IF NOT EXISTS idx_projects_type_active ON projects(type, active) WHERE active = TRUE;
CREATE INDEX IF NOT EXISTS idx_skills_active_category_order ON skills(active, category, "order") WHERE active = TRUE;
CREATE INDEX IF NOT EXISTS idx_experience_active_order ON experience(active, "order") WHERE active = TRUE;

-- Partial indexes for active records (most common queries)
CREATE INDEX IF NOT EXISTS idx_projects_active_partial ON projects(id, title, description, type, featured, "order") WHERE active = TRUE;
CREATE INDEX IF NOT EXISTS idx_skills_active_partial ON skills(id, name, category, proficiency, icon, "order") WHERE active = TRUE;
CREATE INDEX IF NOT EXISTS idx_experience_active_partial ON experience(id, title, company, start_date, end_date, current, "order") WHERE active = TRUE;

-- JSONB indexes for content table (if using JSONB queries)
CREATE INDEX IF NOT EXISTS idx_content_key_gin ON content USING gin (value);

-- Indexes for analytics and tracking
CREATE INDEX IF NOT EXISTS idx_project_views_date_range ON project_views(project_id, viewed_at) WHERE viewed_at >= CURRENT_DATE - INTERVAL '30 days';
CREATE INDEX IF NOT EXISTS idx_visitors_visit_frequency ON visitors(ip, last_visit, visit_count);

-- Indexes for time-based queries
CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at) WHERE active = TRUE;
CREATE INDEX IF NOT EXISTS idx_projects_updated_at ON projects(updated_at) WHERE active = TRUE;

-- Optimize for ORDER BY queries (covering indexes)
CREATE INDEX IF NOT EXISTS idx_skills_covering ON skills(active, "order", category, id, name, proficiency, icon) WHERE active = TRUE;
CREATE INDEX IF NOT EXISTS idx_projects_covering ON projects(active, "order", featured, id, title, type, github_url, live_url) WHERE active = TRUE;

-- Analyze tables to update statistics
ANALYZE projects;
ANALYZE skills;
ANALYZE experience;
ANALYZE content;
ANALYZE project_views;
ANALYZE visitors;
