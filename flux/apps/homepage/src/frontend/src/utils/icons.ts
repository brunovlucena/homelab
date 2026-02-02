// =============================================================================
// 🎨 ICON UTILITIES
// =============================================================================

import React from 'react'
import { 
  SiReact, 
  SiTypescript, 
  SiVite, 
  SiTailwindcss, 
  SiReactrouter,
  SiGo,
  SiPostgresql,
  SiRedis,
  SiKubernetes,
  SiFlux,
  SiArgo,
  SiHelm,
  SiNginx,
  SiDocker,
  SiGithub,
  SiGithubactions,
  SiPulumi,
  SiPrometheus,
  SiGrafana,
  SiAmazon,
  SiTerraform,
  SiKnative,
  SiRabbitmq,
  SiPython
} from 'react-icons/si'
import { FaDatabase, FaStream, FaShieldAlt, FaChartBar, FaCloud, FaRobot, FaRocket } from 'react-icons/fa'
import { BiCertification } from 'react-icons/bi'

/**
 * Icon mapping for technology icons
 * Maps technology names (case-insensitive) to React icon components
 */
const ICON_MAP: Record<string, React.ComponentType<any>> = {
  react: SiReact,
  typescript: SiTypescript,
  vite: SiVite,
  tailwind: SiTailwindcss,
  router: SiReactrouter,
  go: SiGo,
  python: SiPython,
  postgresql: SiPostgresql,
  redis: SiRedis,
  kubernetes: SiKubernetes,
  knative: SiKnative,
  flux: SiFlux,
  argocd: SiArgo,
  helm: SiHelm,
  nginx: SiNginx,
  docker: SiDocker,
  github: SiGithub,
  githubactions: SiGithubactions,
  pulumi: SiPulumi,
  prometheus: SiPrometheus,
  grafana: SiGrafana,
  loki: FaStream,
  tempo: FaDatabase,
  opentelemetry: FaChartBar,
  terraform: SiTerraform,
  aws: SiAmazon,
  rabbitmq: SiRabbitmq,
  shield: FaShieldAlt,
  certification: BiCertification,
  // Homelab and advanced projects
  flagger: FaRocket,
  linkerd: FaRobot,
  cloudevents: FaCloud,
  minio: FaDatabase,
  ollama: FaRobot,
  ai: FaRobot,
  aiagents: FaRobot,
  smartcontracts: FaShieldAlt,
  slither: FaShieldAlt,
  ethereum: FaDatabase,
  solidity: FaRobot,
  defi: FaDatabase,
  kaniko: SiDocker,
  kustomize: SiKubernetes,
  kind: SiKubernetes,
  k3s: SiKubernetes,
}

/**
 * Get the appropriate icon component for a given technology name
 * 
 * @param iconName - The name of the technology/icon (case-insensitive)
 * @returns React component for the icon, defaults to SiGithub if not found
 * 
 * @example
 * const Icon = getIconComponent('react') // Returns SiReact component
 * const Icon = getIconComponent('kubernetes') // Returns SiKubernetes component
 */
export function getIconComponent(iconName: string): React.ComponentType<any> {
  const normalizedName = iconName.toLowerCase().trim()
  return ICON_MAP[normalizedName] || SiGithub
}
