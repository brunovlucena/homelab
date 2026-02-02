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
    company_description TEXT,
    description TEXT,
    technologies TEXT[],
    active BOOLEAN DEFAULT TRUE,
    "order" INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add company_description column if it doesn't exist (for existing databases)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'experience' 
        AND column_name = 'company_description'
    ) THEN
        ALTER TABLE experience ADD COLUMN company_description TEXT;
    END IF;
END $$;

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
CREATE INDEX IF NOT EXISTS idx_content_key ON content(key);

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

-- Insert projects with improved descriptions
INSERT INTO projects (title, description, type, github_url, live_url, technologies, featured, "order", active, github_active) VALUES
(
    'Knative Lambda Operator',
    'Production-grade Function-as-a-Service platform built from scratch in Go. Features a custom Kubernetes Operator with CRDs, CloudEvents routing via RabbitMQ, secure in-cluster container builds using Kaniko (no Docker daemon), and scale-to-zero with Knative Serving. Includes multi-language support (Python/Node.js/Go), full observability stack (Prometheus/Grafana/Loki/Tempo), and 14+ documented SRE runbooks covering incident response, disaster recovery, and capacity planning.',
    'Platform Engineering',
    'https://github.com/brunovlucena/homelab/tree/main/flux/infrastructure/knative-lambda-operator',
    'https://www.youtube.com/watch?v=sYl10lf5OdM',
    ARRAY['Go', 'Kubernetes Operators', 'Knative', 'CloudEvents', 'RabbitMQ', 'Kaniko', 'Prometheus', 'Grafana', 'Loki', 'Tempo', 'OpenTelemetry', 'Flux', 'GitOps'],
    TRUE,
    1,
    TRUE,
    TRUE
),
(
    'Agent-Contracts',
    'AI-powered smart contract security agent scanning 6+ blockchain networks (Ethereum, BNB Chain, Polygon, Arbitrum, Base, Optimism). Combines LLM analysis (Ollama/DeepSeek-Coder) with static analysis (Slither) for vulnerability detection. Runs as 4 event-driven serverless functions (Fetcher → Scanner → Generator → Alerter) on Knative Lambda with CloudEvents, demonstrating full platform dogfooding. Includes unit tests and Prometheus metrics.',
    'AI Security',
    'https://github.com/brunovlucena/homelab/tree/main/ai/agent-contracts',
    'https://youtu.be/sYl10lf5OdM',
    ARRAY['Python', 'LLM', 'Ollama', 'Slither', 'Kubernetes', 'Knative', 'CloudEvents', 'RabbitMQ', 'Multi-chain', 'Web3', 'Security', 'AI Agents', 'Prometheus'],
    TRUE,
    2,
    TRUE,
    TRUE
),
(
    'Multi-Cluster Homelab',
    'Enterprise-grade 6-cluster Kubernetes homelab running on Mac Studio M2 Ultra (192GB RAM), Raspberry Pi edge nodes, and GPU server. Features Linkerd multi-cluster service mesh with mTLS, full GitOps with Flux CD, and infrastructure-as-code with Pulumi. Hosts AI workloads including VLLM serving Llama 3.1 70B, Ollama for local inference, and complete observability stack. Supports development, production, edge (IoT), and GPU training workloads.',
    'Infrastructure',
    'https://github.com/brunovlucena/homelab',
    'https://youtu.be/sYl10lf5OdM',
    ARRAY['Kubernetes', 'Linkerd', 'Flux', 'Pulumi', 'Prometheus', 'Grafana', 'Loki', 'Tempo', 'VLLM', 'Ollama', 'Raspberry Pi', 'K3s', 'Kind', 'Multi-cluster', 'Service Mesh', 'GitOps'],
    TRUE,
    3,
    TRUE,
    TRUE
),
(
    'Homepage',
    'Personal portfolio and homelab showcase website built with React, TypeScript, Go, and cloud-native technologies. Features real-time project updates from GitHub, interactive AI chatbot exploring LLM integration, and comprehensive skill showcase. Deployed on Kubernetes with full CI/CD via GitHub Actions and Flux.',
    'Portfolio Website',
    'https://github.com/brunovlucena/homelab/tree/main/flux/infrastructure/homepage',
    'https://youtu.be/sYl10lf5OdM',
    ARRAY['React', 'TypeScript', 'Go', 'PostgreSQL', 'Redis', 'Docker', 'Kubernetes', 'Nginx', 'GitHub Actions', 'Flux'],
    TRUE,
    4,
    TRUE,
    TRUE
);

-- Insert skills from about section (focused on core competencies)
INSERT INTO skills (name, category, proficiency, icon, "order", active) VALUES
-- Cloud Platforms
('AWS', 'Cloud', 5, 'amazon', 3, TRUE),
('GCP', 'Cloud', 4, 'googlecloud', 4, TRUE),

-- Observability
('Prometheus', 'Observability', 5, 'prometheus', 5, TRUE),
('Grafana', 'Observability', 5, 'grafana', 6, TRUE),
('Loki', 'Observability', 4, 'loki', 7, TRUE),
('Tempo', 'Observability', 4, 'tempo', 8, TRUE),
('OpenTelemetry', 'Observability', 4, 'opentelemetry', 9, TRUE),

-- Infrastructure
('Terraform', 'Infrastructure', 5, 'terraform', 10, TRUE),
('Pulumi', 'Infrastructure', 4, 'pulumi', 11, TRUE),
('ArgoCD', 'Infrastructure', 4, 'argocd', 12, FALSE),

-- Programming Languages
('Go', 'Programming', 5, 'go', 14, TRUE),
('Python', 'Programming', 4, 'python', 15, TRUE),

-- AI/ML
('AI Agents', 'AI/ML', 3, '🤖', 16, TRUE),

-- Tools & Platforms
('Knative', 'Platforms', 5, 'knative', 18, TRUE),

-- Advanced Kubernetes
('Kubernetes Operators', 'Cloud', 4, '☸️', 20, TRUE),
('Multi-cluster Kubernetes', 'Cloud', 4, '☸️', 21, TRUE),

-- Event-Driven Architecture
('Event-Driven Architecture', 'Architecture', 4, '⚡', 22, TRUE),
('CloudEvents', 'Architecture', 4, 'cloudevents', 23, TRUE);

-- Insert experience data in chronological order (oldest to newest)
INSERT INTO experience (title, company, start_date, end_date, current, company_description, description, technologies, "order", active) VALUES
(
    'IT Security Analyst',
    'Tempest Security Intelligence',
    '2011-01-01',
    '2013-10-31',
    FALSE,
    'Tempest Security Intelligence is Brazil''s largest cybersecurity company with 23+ years of experience, offering 70+ solutions across consulting, managed security services, and digital identity. The company employs 450+ security professionals with offices in Recife, São Paulo, and London, and was later acquired by Embraer (2020).',
    '- Vulnerability Assessment: Conducted in-depth vulnerability assessments to identify and mitigate security risks within complex IT environments for 600+ enterprise clients across financial services, retail, and e-commerce sectors.

- Vulnerability Research: Stayed abreast of the latest security threats and vulnerabilities, conducting thorough research to understand their impact and potential exploits.

- Automation Development: Developed and maintained automated tools and scripts (Bash, Ruby) to streamline vulnerability scanning and reporting processes.

- Nessus Plugin Development: Created and customized Nessus Scanner Plugins (NASL) to enhance vulnerability detection capabilities and tailor them to specific security needs.',
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
    'Crealytics is a Berlin-based digital marketing and performance advertising technology company founded in 2008, specializing in retail media solutions for eCommerce retailers. With offices in Berlin, New York, London, and Mumbai, the company serves prominent clients including ASOS, Foot Locker, and Urban Outfitters.',
    '- Cloud Operations: Managed and maintained complex cloud infrastructure on AWS and GCP supporting high-traffic advertising platforms.

- Automation: Implemented automation tools (Saltstack) to streamline operations and reduce manual effort across multi-cloud environments.

- Monitoring and Logging: Deployed and configured monitoring and logging solutions (Prometheus, ELK) to ensure system health and performance for real-time ad serving.

- Distributed Systems: Worked with distributed systems technologies like Mesos, Consul, Kafka, and Linkerd to build scalable and resilient advertising applications.',
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
    'Lesara was a Berlin-based fast-fashion e-commerce startup (2013-2018) that used Big Data to deliver trending products within 10 days. The company expanded to 24 European countries with 100,000+ products and 300+ employees before closing operations.',
    '- Cloud-Native Infrastructure: Designed and implemented a Kubernetes cluster on bare-metal to modernize the infrastructure and support rapid product catalog scaling.

- Automation and CI/CD: Automated infrastructure provisioning and configuration management using Saltstack and Chef to enable fast deployment cycles.

- Monitoring and Logging: Deployed and configured monitoring and logging solutions (Prometheus, ELK) to gain visibility into system health and performance across the e-commerce platform.

- Collaboration: Worked closely with development teams to improve deployment processes and reduce downtime during high-traffic sales events.',
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
    'Namecheap is a leading domain registration and web hosting company founded in 2000, headquartered in Phoenix, Arizona. The company serves millions of customers worldwide with domain names, web hosting, SSL certificates, and website services.',
    '- Cloud Migration and Modernization: Led the migration of legacy infrastructure from VMware ESXi to a Multi-Tenant Kubernetes-based solution on top of OpenStack, improving scalability and resource utilization.

- Infrastructure as Code: Implemented infrastructure as code practices using Terraform to automate provisioning and configuration management across the hosting platform.

- Automation and CI/CD: Developed and maintained automation scripts (Bash, Golang, Ansible, Helm) to streamline operations and improve efficiency for hosting services.',
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
    'Mobimeo is a Berlin-based Mobility-as-a-Service (MaaS) technology company founded by Deutsche Bahn in 2018. The company develops digital platforms that integrate public transportation with modern mobility services (bike-sharing, e-scooters, ride-pooling), and acquired parts of moovel Group in 2020 to become one of Europe''s leading MaaS platform developers.',
    '- Cloud Native Infrastructure: Designed, implemented, and maintained a robust cloud-native infrastructure on AWS, leveraging services like EKS, Kops, and Kubernetes to support millions of mobility users.

- Automation and CI/CD: Automated infrastructure provisioning, deployment, and configuration management using Terraform and GitHub Actions/GitLab CI/CD for rapid feature delivery.

- Observability: Implemented and optimized monitoring, logging, and tracing solutions (Prometheus, Loki, Grafana, Thanos, EFK) to gain deep insights into system performance and user journey analytics.

- Problem-Solving and Troubleshooting: Quickly identified and resolved complex infrastructure issues, minimizing downtime and service disruptions for real-time mobility applications.',
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
    'Promoted to SRE Chapter Lead while continuing as Senior Infrastructure Engineer. Mobimeo is Deutsche Bahn''s MaaS subsidiary developing mobility platforms for millions of users across Germany.',
    '- Team Leadership: Line manager for SRE chapter members, responsible for career development, performance reviews, and technical mentorship.

- Chapter Development: Established SRE best practices, incident response procedures, and reliability standards across the organization.

- Cross-functional Collaboration: Bridged Infrastructure & Operations team with product teams to improve system reliability and deployment velocity.

- Hands-on Engineering: Maintained active involvement in day-to-day infrastructure work while balancing leadership responsibilities.',
    ARRAY['SRE', 'Team Leadership', 'People Management', 'Infrastructure', 'Operations', 'Mentorship', 'Incident Response', 'Reliability Engineering']::TEXT[],
    6,
    TRUE
),
(
    'SRE/DevSecOps/Platform Engineering',
    'Notifi',
    '2023-06-01',
    NULL,
    TRUE,
    'Notifi Network is a cross-chain Web3 messaging infrastructure company that raised $12.5M in funding (including $10M seed round led by Hashed and Race Capital in 2022). The platform provides real-time notifications across 10+ blockchain networks including Solana, Ethereum, Polygon, Arbitrum, BNB Chain, Aptos, and Sui, serving DeFi protocols like Synthetix, Osmosis, and Injective.',
    '- Infrastructure Ownership: Own and manage the entire cloud-native infrastructure supporting real-time notifications across 10+ blockchain networks, serving millions of Web3 users with sub-second delivery.

- Observability Stack: Built the complete observability platform from scratch using Prometheus, Loki, Tempo, Grafana, and OpenTelemetry - monitoring every component from notification pipelines to blockchain event processors.

- Cloud Native Infrastructure: Architect, build, and maintain highly available, scalable, and resilient cloud-native infrastructure using Bare-metal, Kubernetes, AWS and GCP for cross-chain messaging.

- Automation and CI/CD: Built automated deployment pipelines using Terraform, Atmos, Pulumi, and GitHub Actions for reliable and fast releases across multiple environments.

- Knative Lambda Operator: Developed and deployed a production FaaS platform using my open-source Knative Lambda Operator (Go, Kubernetes Operators, CloudEvents, RabbitMQ) - powering serverless workloads for the Fusion notification processing engine.',
    ARRAY['Kubernetes', 'AWS', 'GCP', 'Pulumi', 'Prometheus', 'Loki', 'Tempo', 'Grafana', 'OpenTelemetry', 'Web3', 'Blockchain', 'Terraform', 'Atmos', 'GitHub Actions', 'Flux', 'GitOps', 'Knative', 'Go', 'Kubernetes Operators', 'CloudEvents', 'RabbitMQ']::TEXT[],
    7,
    TRUE
);

-- Insert content data (upsert to handle existing data)
INSERT INTO content (key, value) VALUES
(
    'site_config',
    '{
        "hero_title": "IT Engineer",
        "hero_subtitle": "SRE • Platform Engineering • AI Infrastructure • Kubernetes",
        "resume_title": "Bruno Lucena",
        "resume_subtitle": "IT Engineer",
        "about_title": "About Me",
        "about_description": "IT Engineer with 15+ years of diverse experience spanning Computer and Network Technician, IT Security Analyst, Project Manager, DevOps Engineer, and SRE Lead roles. I have a proven track record of architecting and operating production-grade systems and observability infrastructure using AWS, Baremetal, GCP, Prometheus, Loki, Tempo, Alloy, Mimir, OpenTelemetry, and Grafana.\n\nI also have extensive experience establishing systems from the ground up - from prototyping on Raspberry Pi to production multi-region Kubernetes clusters, from mobile applications to distributed cloud infrastructure. I''ve built comprehensive observability platforms through sophisticated automation using both traditional Terraform and modern Infrastructure-as-Code tools like Pulumi. Currently, I''m developing agent-sre, an AI-powered system that automatically responds to alerts by following runbooks, significantly reducing manual toil and enabling faster incident resolution."
    }'
),
(
    'about',
    '{"description": "IT Engineer with 15+ years of diverse experience spanning Computer and Network Technician, IT Security Analyst, Project Manager, DevOps Engineer, and SRE Lead roles. I have a proven track record of architecting and operating production-grade systems and observability infrastructure using AWS, Baremetal, GCP, Prometheus, Loki, Tempo, Alloy, Mimir, OpenTelemetry, and Grafana.\n\nI also have extensive experience establishing systems from the ground up - from prototyping on Raspberry Pi to production multi-region Kubernetes clusters, from mobile applications to distributed cloud infrastructure. I''ve built comprehensive observability platforms through sophisticated automation using both traditional Terraform and modern Infrastructure-as-Code tools like Pulumi. Currently, I''m developing agent-sre, an AI-powered system that automatically responds to alerts by following runbooks, significantly reducing manual toil and enabling faster incident resolution."}'
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

-- Update projects active status
-- Enable only knative-lambda-operator, disable all others
UPDATE projects SET active = false WHERE active = true;
UPDATE projects 
SET active = true 
WHERE LOWER(title) LIKE '%knative%lambda%operator%' 
   OR LOWER(title) = 'knative lambda operator';
