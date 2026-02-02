import { 
  SiGo,
  SiPython,
  SiKubernetes,
  SiKnative,
  SiDocker,
  SiFlux,
  SiPulumi,
  SiTerraform,
  SiAmazon,
  SiPrometheus,
  SiGrafana,
  SiArgo,
} from 'react-icons/si';
import { 
  FaRocket, 
  FaRobot, 
  FaCloud,
  FaLock,
  FaSearch,
  FaChartBar,
  FaUsers,
  FaAws,
  FaBolt,
  FaBook,
  FaCogs,
  FaSyncAlt,
  FaCubes,
  FaHardHat,
  FaFileAlt,
  FaBalanceScale,
  FaDoorOpen,
  FaProjectDiagram,
  FaShieldAlt,
} from 'react-icons/fa';
import type { Skill } from '../types';

// Icon mapping for skills
const iconMap: Record<string, React.ComponentType<{ className?: string; style?: React.CSSProperties }>> = {
  // Programming Languages
  'go': SiGo,
  'golang': SiGo,
  'python': SiPython,
  
  // Container & Orchestration
  'kubernetes': SiKubernetes,
  'k8s': SiKubernetes,
  'knative': SiKnative,
  'docker': SiDocker,
  
  // GitOps & IaC
  'flux': SiFlux,
  'fluxcd': SiFlux,
  'pulumi': SiPulumi,
  'terraform': SiTerraform,
  'argocd': SiArgo,
  'gitops': FaCubes,
  'infrastructure as code': FaHardHat,
  
  // Cloud Providers
  'aws': SiAmazon,
  'amazon': SiAmazon,
  'aws eks': FaAws,
  'aws lambda': FaBolt,
  
  // Observability
  'prometheus': SiPrometheus,
  'grafana': SiGrafana,
  'loki': FaFileAlt,
  'tempo': FaSearch,
  'opentelemetry': FaChartBar,
  'logging': FaFileAlt,
  'tracing': FaSearch,
  
  // DevOps & SRE
  'flagger': FaRocket,
  'cloudevents': FaCloud,
  'ci/cd': FaSyncAlt,
  'site reliability engineering': FaCogs,
  'devsecops': FaShieldAlt,
  'event-driven architecture': FaBolt,
  
  // Security
  'it security': FaLock,
  'llm security': FaLock,
  'smart contract security': FaLock,
  
  // AI/ML
  'ai': FaRobot,
  'llm infrastructure': FaRobot,
  'ai agents': FaRobot,
  'rag': FaBook,
  
  // Management
  'project management': FaChartBar,
  'team leadership': FaUsers,
  
  // Networking
  'load balancing': FaBalanceScale,
  'api gateway': FaDoorOpen,
  'service mesh': FaProjectDiagram,
  
  // Kubernetes Advanced
  'kubernetes operators': SiKubernetes,
  'multi-cluster kubernetes': SiKubernetes,
};

// Color mapping for skills
const colorMap: Record<string, string> = {
  // Programming Languages
  'go': '#00ADD8',
  'golang': '#00ADD8',
  'python': '#3776AB',
  
  // Container & Orchestration
  'kubernetes': '#326CE5',
  'k8s': '#326CE5',
  'knative': '#0865AD',
  'docker': '#2496ED',
  
  // GitOps & IaC
  'flux': '#5468FF',
  'fluxcd': '#5468FF',
  'pulumi': '#8A3391',
  'terraform': '#7B42BC',
  'argocd': '#EF7B4D',
  'gitops': '#F05032',
  'infrastructure as code': '#6366F1',
  
  // Cloud Providers
  'aws': '#FF9900',
  'amazon': '#FF9900',
  'aws eks': '#FF9900',
  'aws lambda': '#FF9900',
  
  // Observability
  'prometheus': '#E6522C',
  'grafana': '#F46800',
  'loki': '#F15922',
  'tempo': '#FBB034',
  'opentelemetry': '#425CC7',
  'logging': '#10B981',
  'tracing': '#8B5CF6',
  
  // DevOps & SRE
  'flagger': '#326CE5',
  'cloudevents': '#326CE5',
  'ci/cd': '#2563EB',
  'site reliability engineering': '#6366F1',
  'devsecops': '#EF4444',
  'event-driven architecture': '#F59E0B',
  
  // Security
  'it security': '#DC2626',
  'llm security': '#DC2626',
  'smart contract security': '#DC2626',
  
  // AI/ML
  'ai': '#10B981',
  'llm infrastructure': '#8B5CF6',
  'ai agents': '#06B6D4',
  'rag': '#14B8A6',
  
  // Management
  'project management': '#3B82F6',
  'team leadership': '#8B5CF6',
  
  // Networking
  'load balancing': '#6366F1',
  'api gateway': '#0EA5E9',
  'service mesh': '#7C3AED',
  
  // Kubernetes Advanced
  'kubernetes operators': '#326CE5',
  'multi-cluster kubernetes': '#326CE5',
};

// Default skills to show when API data is not available
const defaultSkills: Array<{ name: string; icon: string }> = [
  { name: 'Go', icon: 'go' },
  { name: 'Python', icon: 'python' },
  { name: 'Kubernetes', icon: 'kubernetes' },
  { name: 'Knative', icon: 'knative' },
  { name: 'Docker', icon: 'docker' },
  { name: 'Flagger', icon: 'flagger' },
  { name: 'Flux', icon: 'flux' },
  { name: 'Pulumi', icon: 'pulumi' },
  { name: 'Terraform', icon: 'terraform' },
  { name: 'CloudEvents', icon: 'cloudevents' },
  { name: 'AWS', icon: 'aws' },
  { name: 'Knative', icon: 'knative' },
];

interface SkillsGridProps {
  skills?: Skill[];
  loading?: boolean;
}

const SkillsGrid: React.FC<SkillsGridProps> = ({ skills, loading = false }) => {
  // Use API skills if available, otherwise use default skills
  const displaySkills = skills && skills.length > 0 
    ? skills 
    : defaultSkills.map((s, idx) => ({ 
        id: idx, 
        name: s.name, 
        icon: s.icon, 
        category: 'default', 
        order: idx 
      }));

  if (loading) {
    return (
      <div className="skills-grid" aria-busy="true" aria-label="Loading skills">
        {[...Array(12)].map((_, idx) => (
          <div key={idx} className="skill-tag skeleton" aria-hidden="true">
            <div className="skill-icon-placeholder" />
            <span className="skill-name-placeholder" />
          </div>
        ))}
      </div>
    );
  }

  return (
    <div className="skills-grid" role="list" aria-label="Technical skills">
      {displaySkills.map((skill) => {
        const iconKey = skill.icon?.toLowerCase() ?? skill.name.toLowerCase();
        const IconComponent = iconMap[iconKey] ?? FaRobot;
        const color = colorMap[iconKey] ?? '#6B7280';
        
        return (
          <div 
            key={skill.id} 
            className="skill-tag" 
            role="listitem"
          >
            <IconComponent 
              className="skill-icon" 
              style={{ color }} 
              aria-hidden="true"
            />
            <span>{skill.name}</span>
          </div>
        );
      })}
    </div>
  );
};

export default SkillsGrid;
