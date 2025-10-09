-- Complete database schema and initial data for Bruno site system
-- Migration: 001_complete_schema.sql

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(100) NOT NULL,
    github_url VARCHAR(500),
    live_url VARCHAR(500),
    technologies TEXT[],
    featured BOOLEAN DEFAULT FALSE,
    active BOOLEAN DEFAULT TRUE,
    github_active BOOLEAN DEFAULT TRUE,
    "order" INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Add github_active column if it doesn't exist (for existing databases)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'projects' 
        AND column_name = 'github_active'
    ) THEN
        ALTER TABLE projects ADD COLUMN github_active BOOLEAN DEFAULT TRUE;
    END IF;
END $$;

-- Project views tracking
CREATE TABLE IF NOT EXISTS project_views (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    ip VARCHAR(45) NOT NULL,
    user_agent TEXT,
    referrer VARCHAR(500),
    viewed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Visitors table
CREATE TABLE IF NOT EXISTS visitors (
    id SERIAL PRIMARY KEY,
    ip VARCHAR(45) UNIQUE NOT NULL,
    user_agent TEXT,
    country VARCHAR(100),
    city VARCHAR(100),
    first_visit TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_visit TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    visit_count INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Content management table
CREATE TABLE IF NOT EXISTS content (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Skills table
CREATE TABLE IF NOT EXISTS skills (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(100) NOT NULL,
    proficiency INTEGER DEFAULT 1 CHECK (proficiency >= 1 AND proficiency <= 5),
    icon VARCHAR(50),
    active BOOLEAN DEFAULT TRUE,
    "order" INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add active column to skills if it doesn't exist (for existing databases)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'skills' 
        AND column_name = 'active'
    ) THEN
        ALTER TABLE skills ADD COLUMN active BOOLEAN DEFAULT TRUE;
    END IF;
END $$;

-- Experience table
CREATE TABLE IF NOT EXISTS experience (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    company VARCHAR(255) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    current BOOLEAN DEFAULT FALSE,
    description TEXT,
    technologies TEXT[],
    active BOOLEAN DEFAULT TRUE,
    "order" INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add active column to experience if it doesn't exist (for existing databases)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'experience' 
        AND column_name = 'active'
    ) THEN
        ALTER TABLE experience ADD COLUMN active BOOLEAN DEFAULT TRUE;
    END IF;
END $$;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_projects_type ON projects(type);
CREATE INDEX IF NOT EXISTS idx_projects_featured ON projects(featured);
CREATE INDEX IF NOT EXISTS idx_projects_active ON projects(active);
CREATE INDEX IF NOT EXISTS idx_projects_order ON projects("order");
CREATE INDEX IF NOT EXISTS idx_project_views_project_id ON project_views(project_id);
CREATE INDEX IF NOT EXISTS idx_project_views_viewed_at ON project_views(viewed_at);
CREATE INDEX IF NOT EXISTS idx_visitors_ip ON visitors(ip);
CREATE INDEX IF NOT EXISTS idx_visitors_last_visit ON visitors(last_visit);
CREATE INDEX IF NOT EXISTS idx_skills_category ON skills(category);
CREATE INDEX IF NOT EXISTS idx_skills_active ON skills(active);
CREATE INDEX IF NOT EXISTS idx_experience_company ON experience(company);
CREATE INDEX IF NOT EXISTS idx_experience_active ON experience(active);
CREATE INDEX IF NOT EXISTS idx_experience_order ON experience("order");

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_visitors_updated_at ON visitors;
CREATE TRIGGER update_visitors_updated_at BEFORE UPDATE ON visitors FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_content_updated_at ON content;
CREATE TRIGGER update_content_updated_at BEFORE UPDATE ON content FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_skills_updated_at ON skills;
CREATE TRIGGER update_skills_updated_at BEFORE UPDATE ON skills FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_experience_updated_at ON experience;
CREATE TRIGGER update_experience_updated_at BEFORE UPDATE ON experience FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Clear existing data
TRUNCATE TABLE projects CASCADE;
TRUNCATE TABLE experience CASCADE;
TRUNCATE TABLE skills CASCADE;

-- Insert Bruno Site, Knative Lambda, and Home Infrastructure projects
INSERT INTO projects (title, description, type, github_url, live_url, technologies, featured, "order", active, github_active) VALUES
(
    'Bruno Site',
    'Personal portfolio and homelab showcase website built with React, TypeScript, Go, and modern cloud-native technologies. Features real-time project updates, interactive chatbot, and comprehensive skill showcase.',
    'Portfolio Website',
    'https://github.com/brunovlucena/bruno-site',
    'https://youtu.be/RalxXaLAsVU',
    ARRAY['React', 'TypeScript', 'Go', 'PostgreSQL', 'Redis', 'Docker', 'Kubernetes', 'Nginx'],
    TRUE,
    1,
    TRUE,
    TRUE
),
(
    'Knative Lambda',
    'Serverless functions and cloud-native development platform using Knative for scalable, event-driven applications with Kubernetes.',
    'Serverless',
    'https://github.com/brunovlucena/knative-lambda',
    'https://www.youtube.com/watch?v=cTTxlCr8N2Q',
    ARRAY['Knative', 'Kubernetes', 'Serverless', 'CloudEvents', 'Go'],
    TRUE,
    2,
    TRUE,
    FALSE
),
(
    'Home Infrastructure',
    'Personal homelab infrastructure project using Flux, Pulumi, and Kubernetes for managing a complete cloud-native environment. Features automated deployments, monitoring, and infrastructure as code practices.',
    'Infrastructure',
    'https://github.com/brunovlucena/home',
    'https://youtu.be/RalxXaLAsVU',
    ARRAY['Flux', 'Pulumi', 'Kubernetes', 'Docker', 'Helm', 'Prometheus', 'Grafana', 'Loki', 'Tempo', 'Alloy', 'Cert-Manager', 'Nginx'],
    TRUE,
    3,
    TRUE,
    TRUE
);

-- Insert skills from about section
INSERT INTO skills (name, category, proficiency, icon, "order", active) VALUES
-- IT Security
('IT Security', 'Security', 5, '🔒', 1, TRUE),
('Vulnerability Assessment', 'Security', 5, '🔍', 2, TRUE),
('Nessus', 'Security', 4, '🛡️', 3, TRUE),
('Security Auditing', 'Security', 4, '📋', 4, TRUE),

-- Project Management
('Project Management', 'Management', 4, '📊', 5, TRUE),
('Team Leadership', 'Management', 4, '👥', 6, TRUE),
('Agile/Scrum', 'Management', 4, '🔄', 7, TRUE),

-- Kubernetes & Cloud
('Kubernetes', 'Cloud', 5, '☸️', 8, TRUE),
('AWS EKS', 'Cloud', 5, '☁️', 9, TRUE),
('GCP', 'Cloud', 4, '☁️', 10, TRUE),
('AWS Lambda', 'Cloud', 4, '⚡', 11, TRUE),
('OpenStack', 'Cloud', 3, '☁️', 12, TRUE),

-- Observability
('Prometheus', 'Observability', 5, 'prometheus', 13, TRUE),
('Grafana', 'Observability', 5, 'grafana', 14, TRUE),
('Loki', 'Observability', 4, 'loki', 15, TRUE),
('Tempo', 'Observability', 4, 'tempo', 16, TRUE),
('OpenTelemetry', 'Observability', 4, 'opentelemetry', 17, TRUE),

-- Infrastructure
('Terraform', 'Infrastructure', 5, 'terraform', 18, TRUE),
('Pulumi', 'Infrastructure', 4, 'pulumi', 19, TRUE),
('Docker', 'Infrastructure', 5, 'docker', 20, TRUE),
('Flux', 'Infrastructure', 4, 'flux', 21, TRUE),
('Helm', 'Infrastructure', 4, 'helm', 22, TRUE),

-- Programming Languages
('Go', 'Programming', 5, 'go', 23, TRUE),
('Python', 'Programming', 4, 'python', 24, TRUE),
('TypeScript', 'Programming', 4, 'typescript', 25, TRUE),
('JavaScript', 'Programming', 4, 'javascript', 26, TRUE),
('Bash', 'Programming', 4, 'bash', 27, TRUE),

-- Databases & Messaging
('PostgreSQL', 'Database', 5, 'postgresql', 28, TRUE),
('Redis', 'Database', 4, 'redis', 29, TRUE),
('RabbitMQ', 'Messaging', 4, 'rabbitmq', 30, TRUE),
('MongoDB', 'Database', 3, 'mongodb', 31, TRUE),

-- AI/ML
('Machine Learning', 'AI/ML', 4, '🤖', 32, TRUE),
('TensorFlow', 'AI/ML', 4, '📊', 33, TRUE),
('Natural Language Processing', 'AI/ML', 4, '💬', 34, TRUE),
('Computer Vision', 'AI/ML', 3, '👁️', 35, TRUE),

-- DevOps & SRE
('Site Reliability Engineering', 'DevOps', 5, '⚙️', 36, TRUE),
('DevSecOps', 'DevOps', 5, '🔒', 37, TRUE),
('CI/CD', 'DevOps', 5, '🔄', 38, TRUE),
('GitOps', 'DevOps', 4, '📦', 39, TRUE),
('Infrastructure as Code', 'DevOps', 5, '🏗️', 40, TRUE),

-- Monitoring & Alerting
('Monitoring', 'Monitoring', 5, '📊', 41, TRUE),
('Alerting', 'Monitoring', 5, '🚨', 42, TRUE),
('Logging', 'Monitoring', 5, '📝', 43, TRUE),
('Tracing', 'Monitoring', 4, '🔍', 44, TRUE),
('Metrics', 'Monitoring', 5, '📈', 45, TRUE),

-- Cloud Platforms
('AWS', 'Cloud', 5, 'amazon', 46, TRUE),
('Google Cloud Platform', 'Cloud', 4, 'googlecloud', 47, TRUE),
('Azure', 'Cloud', 3, 'azure', 48, TRUE),
('Multi-cloud', 'Cloud', 4, 'multicloud', 49, TRUE),

-- Networking & Security
('Network Security', 'Security', 4, '🛡️', 50, TRUE),
('Load Balancing', 'Networking', 4, '⚖️', 51, TRUE),
('API Gateway', 'Networking', 4, '🚪', 52, TRUE),
('Service Mesh', 'Networking', 4, '🕸️', 53, TRUE),
('VPN', 'Security', 4, '🔐', 54, TRUE),

-- Tools & Platforms
('GitHub', 'Tools', 5, 'github', 55, TRUE),
('GitLab', 'Tools', 4, 'gitlab', 56, TRUE),
('Jenkins', 'Tools', 4, 'jenkins', 57, TRUE),
('ArgoCD', 'Tools', 4, 'argocd', 58, TRUE),
('Knative', 'Platforms', 4, 'knative', 59, TRUE),
('Serverless', 'Platforms', 4, 'serverless', 60, TRUE),
('GitHub Actions', 'Tools', 5, 'githubactions', 61, TRUE),
('Atmos', 'Tools', 4, 'atmos', 62, TRUE),
('Vertex AI', 'AI/ML', 4, 'vertexai', 63, TRUE),
('RAG', 'AI/ML', 4, 'rag', 64, TRUE),
('CloudEvents', 'Platforms', 4, 'cloudevents', 65, TRUE),
('Security', 'Security', 5, 'security', 66, TRUE),
('Compliance', 'Security', 4, 'compliance', 67, TRUE);

-- Insert experience data in chronological order (oldest to newest)
INSERT INTO experience (title, company, start_date, end_date, current, description, technologies, "order", active) VALUES
(
    'IT Security Analyst',
    'Tempest Security Intelligence',
    '2011-01-01',
    '2013-10-31',
    FALSE,
    'Key Responsibilities:

- Vulnerability Assessment: Conduct in-depth vulnerability assessments to identify and mitigate security risks within complex IT environments.

- Vulnerability Research: Stay abreast of the latest security threats and vulnerabilities, conducting thorough research to understand their impact and potential exploits.

- Automation Development: Develop and maintain automated tools and scripts (e.g., Bash, Ruby) to streamline vulnerability scanning and reporting processes.

- Nessus Plugin Development: Create and customize Nessus Scanner Plugins (NASL) to enhance vulnerability detection capabilities and tailor them to specific security needs.',
    ARRAY['Vulnerability Assessment', 'Security Research', 'Bash', 'Ruby', 'Nessus', 'NASL', 'Security Automation', 'Vulnerability Scanning', 'Security Tools']::TEXT[],
    1,
    TRUE
),
(
    'Operations Engineer',
    'Crealytics',
    '2017-08-01',
    '2018-03-31',
    FALSE,
    'Key Responsibilities:

- Cloud Operations: Managed and maintained complex cloud infrastructure on AWS and GCP.

- Automation: Implemented automation tools (Saltstack) to streamline operations and reduce manual effort.

- Monitoring and Logging: Deployed and configured monitoring and logging solutions (Prometheus, ELK) to ensure system health and performance.

- Distributed Systems: Worked with distributed systems technologies like Mesos, Consul, Kafka, and Linkerd to build scalable and resilient applications.',
    ARRAY['AWS', 'GCP', 'Saltstack', 'Prometheus', 'ELK', 'Mesos', 'Consul', 'Kafka', 'Linkerd', 'Distributed Systems', 'Cloud Operations', 'Automation', 'Monitoring', 'Logging']::TEXT[],
    2,
    TRUE
),
(
    'DevOps Engineer',
    'Lesara',
    '2018-04-01',
    '2018-12-31',
    FALSE,
    'Key Responsibilities:

- Cloud-Native Infrastructure: Designed and implemented a Kubernetes cluster on bare-metal to modernize the infrastructure.

- Automation and CI/CD: Automated infrastructure provisioning and configuration management using Saltstack and Chef.

- Monitoring and Logging: Deployed and configured monitoring and logging solutions (Prometheus, ELK) to gain visibility into system health and performance.

- Collaboration: Worked closely with development teams to improve deployment processes and reduce downtime.',
    ARRAY['Kubernetes', 'Bare-metal', 'Saltstack', 'Chef', 'Prometheus', 'ELK', 'Automation', 'CI/CD', 'Monitoring', 'Logging', 'Infrastructure', 'Collaboration']::TEXT[],
    3,
    TRUE
),
(
    'Cloud Consultant',
    'Namecheap, Inc',
    '2019-03-01',
    '2019-08-31',
    FALSE,
    'Key Responsibilities:

- Cloud Migration and Modernization: Took part in the migration of legacy infrastructure from VMware ESXi to a Multi-Tenant Kubernetes-based solution on top of OpenStack.

- Infrastructure as Code: Implemented infrastructure as code practices using Terraform to automate provisioning and configuration management.
 
- Automation and CI/CD: Developed and maintained automation scripts (Bash, Golang, Ansible, Helm) to streamline operations and improve efficiency.',
    ARRAY['Cloud Migration', 'VMware ESXi', 'Kubernetes', 'OpenStack', 'Terraform', 'Bash', 'Golang', 'Ansible', 'Helm', 'Infrastructure as Code', 'Automation', 'CI/CD']::TEXT[],
    4,
    TRUE
),
(
    'Senior Infrastructure Engineer',
    'Mobimeo',
    '2020-02-01',
    '2023-03-31',
    FALSE,
    'Key Responsibilities:

- Cloud Native Infrastructure: Designed, implemented, and maintained a robust cloud-native infrastructure on AWS, leveraging services like EKS, Kops, and Kubernetes.

- Automation and CI/CD: Automated infrastructure provisioning, deployment, and configuration management using Terraform and GitHub Actions/GitLab CI/CD.

- Observability: Implemented and optimized monitoring, logging, and tracing solutions (Prometheus, Loki, Grafana, Thanos, EFK) to gain deep insights into system performance and behavior.

- Problem-Solving and Troubleshooting: Quickly identified and resolved complex infrastructure issues, minimizing downtime and service disruptions.',
    ARRAY['AWS', 'EKS', 'Kops', 'Kubernetes', 'Terraform', 'GitHub Actions', 'GitLab CI/CD', 'Prometheus', 'Loki', 'Grafana', 'Thanos', 'EFK', 'Infrastructure', 'Automation', 'CI/CD', 'Observability', 'Troubleshooting']::TEXT[],
    5,
    TRUE
),
(
    'SRE Chapter Lead',
    'Mobimeo',
    '2021-12-01',
    '2023-03-31',
    FALSE,
    'The SRE chapter lead is the line manager for the chapter members, responsible for developing people and the things happening in the SRE chapter but still is a member of the infrastructure & Operations Team and does day-to-day work.',
    ARRAY['SRE', 'Team Leadership', 'People Management', 'Infrastructure', 'Operations']::TEXT[],
    6,
    TRUE
),

(
    'SRE/DevOps',
    'Notifi',
    '2023-06-01',
    NULL,
    TRUE,
    'Key Responsibilities:

- Cloud Native Infrastructure: Architect, build, and maintain highly available, scalable, and resilient cloud-native infrastructure using Kubernetes, AWS, GCP, Pulumi, and many others

- Observability: Implement and optimize monitoring, logging, and tracing solutions (Prometheus, Loki, Tempo, Grafana, OpenTelemetry) to gain deep insights into system performance and behavior.

- Chatbot for SRE: RAG, Vertex AI

- Automation and CI/CD: Automate infrastructure provisioning, deployment, and configuration management using Terraform, Atmos, and GitHub Actions to accelerate development and reduce errors.

- Serverless and Function-as-a-Service: Develop and deploy serverless applications on AWS Lambda 

- Serverless on K8s: Knative (CloudEvents, RabbitMQ), Golang 

- Security and Compliance: Ensure the security and compliance of systems and applications by implementing best practices and leveraging security tools.',
    ARRAY['Kubernetes', 'AWS', 'GCP', 'Pulumi', 'Prometheus', 'Loki', 'Tempo', 'Grafana', 'OpenTelemetry', 'RAG', 'Vertex AI', 'Terraform', 'Atmos', 'GitHub Actions', 'AWS Lambda', 'Knative', 'CloudEvents', 'RabbitMQ', 'Golang', 'Security', 'Compliance']::TEXT[],
    7,
    TRUE
);

-- Insert content data (upsert to handle existing data)
INSERT INTO content (key, value) VALUES
(
    'about',
    '{"description": "Senior Cloud Native Infrastructure Engineer with extensive experience in designing, implementing, and maintaining scalable, resilient cloud-native infrastructure. Passionate about automation, observability, and modern DevOps practices."}'
),
(
    'contact',
    '{"email": "bruno@lucena.cloud", "location": "Brazil", "linkedin": "https://www.linkedin.com/in/bvlucena", "github": "https://github.com/brunovlucena", "availability": "Open to new opportunities"}'
)
ON CONFLICT (key) DO UPDATE SET 
    value = EXCLUDED.value,
    updated_at = CURRENT_TIMESTAMP;

-- Verify all data
SELECT 'Projects' as table_name, COUNT(*) as count FROM projects
UNION ALL
SELECT 'Skills' as table_name, COUNT(*) as count FROM skills
UNION ALL
SELECT 'Experience' as table_name, COUNT(*) as count FROM experience
UNION ALL
SELECT 'Content' as table_name, COUNT(*) as count FROM content;
