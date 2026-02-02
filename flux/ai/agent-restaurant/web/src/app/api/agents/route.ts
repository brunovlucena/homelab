import { NextRequest, NextResponse } from 'next/server'

// Force Node.js runtime (not Edge) for file system access
export const runtime = 'nodejs'
export const dynamic = 'force-dynamic'

// Configuration from environment
const KUBERNETES_API = process.env.KUBERNETES_API_URL || 'https://kubernetes.default.svc'
const PROMETHEUS_URL = process.env.PROMETHEUS_URL || 'http://prometheus.prometheus.svc.cluster.local:9090'
const NAMESPACE = process.env.NAMESPACE || 'agent-restaurant'

interface AgentInfo {
  name: string
  role: string
  status: 'online' | 'offline' | 'degraded'
  version: string
  lastHeartbeat: string
  metrics: {
    cpu: number
    memory: number
    requests: number
    latency: number
  }
}

async function queryPrometheus(query: string): Promise<any> {
  try {
    const url = `${PROMETHEUS_URL}/api/v1/query?query=${encodeURIComponent(query)}`
    const response = await fetch(url, { 
      headers: { 'Accept': 'application/json' },
      cache: 'no-store',
    })
    if (!response.ok) return null
    return await response.json()
  } catch (error) {
    console.error('Prometheus query error:', error)
    return null
  }
}

async function getKubernetesAgents(): Promise<AgentInfo[] | null> {
  const tokenPath = '/var/run/secrets/kubernetes.io/serviceaccount/token'
  
  try {
    // Try to read the service account token (when running in cluster)
    let token = ''
    let isLocalDev = false
    
    try {
      const { readFile } = await import('node:fs/promises')
      token = (await readFile(tokenPath, 'utf-8')).trim()
    } catch (tokenError: any) {
      if (tokenError?.code === 'ENOENT') {
        // Running locally - use kubectl proxy (no auth needed)
        console.log('[K8s API] Running locally, using kubectl proxy (no token needed)')
        isLocalDev = true
      } else {
        throw tokenError
      }
    }
    
    const url = `${KUBERNETES_API}/apis/lambda.knative.io/v1alpha1/namespaces/${NAMESPACE}/lambdaagents`
    
    // When using kubectl proxy locally, no Authorization header is needed
    const headers: Record<string, string> = {
      'Accept': 'application/json',
    }
    if (token && !isLocalDev) {
      headers['Authorization'] = `Bearer ${token}`
    }
    
    const response = await fetch(url, {
      headers,
      cache: 'no-store',
    })
    
    if (!response.ok) {
      const errorText = await response.text()
      console.error('[K8s API] Response not ok:', response.status, errorText)
      return null
    }
    
    const data = await response.json()
    
    return data.items?.map((item: any) => ({
      name: item.metadata.name,
      role: item.metadata.labels?.['restaurant.agent.role'] || item.metadata.labels?.['agentchat.role'] || item.metadata.name,
      status: item.status?.conditions?.find((c: any) => c.type === 'Ready')?.status === 'True' ? 'online' : 'offline',
      version: item.spec?.image?.tag || 'unknown',
      lastHeartbeat: item.status?.lastUpdated || new Date().toISOString(),
      metrics: { cpu: 0, memory: 0, requests: 0, latency: 0 },
    })) || []
  } catch (error: any) {
    console.error('[K8s API] Error:', error)
    return null
  }
}

async function getAgentMetricsFromPrometheus(agentNames: string[]): Promise<Record<string, any>> {
  const metrics: Record<string, any> = {}
  
  for (const name of agentNames) {
    metrics[name] = { cpu: 0, memory: 0, requests: 0, latency: 0 }
  }
  
  // Get CPU usage
  const cpuResult = await queryPrometheus(
    `sum(rate(container_cpu_usage_seconds_total{namespace="${NAMESPACE}"}[5m])) by (pod) * 100`
  )
  if (cpuResult?.data?.result) {
    cpuResult.data.result.forEach((r: any) => {
      const pod = r.metric.pod || ''
      const agentName = agentNames.find(n => pod.includes(n))
      if (agentName && metrics[agentName]) {
        metrics[agentName].cpu = Math.round(parseFloat(r.value[1]) || 0)
      }
    })
  }
  
  // Get memory usage
  const memResult = await queryPrometheus(
    `sum(container_memory_usage_bytes{namespace="${NAMESPACE}"}) by (pod) / 1024 / 1024`
  )
  if (memResult?.data?.result) {
    memResult.data.result.forEach((r: any) => {
      const pod = r.metric.pod || ''
      const agentName = agentNames.find(n => pod.includes(n))
      if (agentName && metrics[agentName]) {
        metrics[agentName].memory = Math.round(parseFloat(r.value[1]) || 0)
      }
    })
  }
  
  // Get request count
  const reqResult = await queryPrometheus(
    `sum(increase(http_requests_total{namespace="${NAMESPACE}"}[1h])) by (service)`
  )
  if (reqResult?.data?.result) {
    reqResult.data.result.forEach((r: any) => {
      const service = r.metric.service || ''
      const agentName = agentNames.find(n => service.includes(n))
      if (agentName && metrics[agentName]) {
        metrics[agentName].requests = Math.round(parseFloat(r.value[1]) || 0)
      }
    })
  }
  
  // Get latency (p95)
  const latResult = await queryPrometheus(
    `histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{namespace="${NAMESPACE}"}[5m])) by (le, service)) * 1000`
  )
  if (latResult?.data?.result) {
    latResult.data.result.forEach((r: any) => {
      const service = r.metric.service || ''
      const agentName = agentNames.find(n => service.includes(n))
      if (agentName && metrics[agentName]) {
        metrics[agentName].latency = Math.round(parseFloat(r.value[1]) || 0)
      }
    })
  }
  
  return metrics
}

export async function GET(request: NextRequest) {
  try {
    // Get agents from Kubernetes
    const k8sAgents = await getKubernetesAgents()
    
    if (!k8sAgents || k8sAgents.length === 0) {
      return NextResponse.json({
        success: false,
        agents: [],
        source: 'unavailable',
        message: 'Could not connect to Kubernetes API. Running locally without cluster access. Set KUBERNETES_API_URL and ensure proper RBAC.',
      })
    }
    
    // Get metrics from Prometheus
    const agentNames = k8sAgents.map(a => a.name)
    const prometheusMetrics = await getAgentMetricsFromPrometheus(agentNames)
    
    // Merge metrics with agents
    const agents = k8sAgents.map(agent => ({
      ...agent,
      metrics: prometheusMetrics[agent.name] || agent.metrics,
    }))
    
    return NextResponse.json({
      success: true,
      agents,
      source: 'kubernetes+prometheus',
      timestamp: new Date().toISOString(),
    })
  } catch (error) {
    console.error('Agents API error:', error)
    return NextResponse.json({
      success: false,
      error: 'Failed to fetch agents',
      message: String(error),
    }, { status: 500 })
  }
}
