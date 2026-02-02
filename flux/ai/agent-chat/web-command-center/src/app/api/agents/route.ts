import { NextRequest, NextResponse } from 'next/server'

const PROMETHEUS_URL = process.env.PROMETHEUS_URL || 'http://prometheus.prometheus.svc.cluster.local:9090'
const KUBERNETES_API = process.env.KUBERNETES_API_URL || 'https://kubernetes.default.svc'
const NAMESPACE = process.env.NAMESPACE || 'agent-chat'

async function queryPrometheus(query: string): Promise<any> {
  try {
    const response = await fetch(
      `${PROMETHEUS_URL}/api/v1/query?query=${encodeURIComponent(query)}`,
      { headers: { 'Accept': 'application/json' }, cache: 'no-store' }
    )
    if (!response.ok) return null
    return await response.json()
  } catch {
    return null
  }
}

async function getKubernetesAgents() {
  try {
    const tokenPath = '/var/run/secrets/kubernetes.io/serviceaccount/token'
    let token = ''
    try {
      const fs = await import('fs/promises')
      token = await fs.readFile(tokenPath, 'utf-8')
    } catch {
      return null
    }
    
    const response = await fetch(
      `${KUBERNETES_API}/apis/lambda.knative.io/v1alpha1/namespaces/${NAMESPACE}/lambdaagents`,
      { headers: { 'Authorization': `Bearer ${token}`, 'Accept': 'application/json' } }
    )
    
    if (!response.ok) return null
    const data = await response.json()
    
    return data.items?.map((item: any) => ({
      name: item.metadata.name,
      status: item.status?.conditions?.find((c: any) => c.type === 'Ready')?.status === 'True' ? 'online' : 'offline',
      version: item.spec?.image?.tag || 'unknown',
      lastHeartbeat: item.status?.lastUpdated || new Date().toISOString(),
      role: item.metadata.labels?.['agentchat.role'] || 'agent',
    })) || []
  } catch {
    return null
  }
}

export async function GET(request: NextRequest) {
  try {
    const k8sAgents = await getKubernetesAgents()
    
    if (!k8sAgents || k8sAgents.length === 0) {
      return NextResponse.json({
        success: false,
        agents: [],
        source: 'unavailable',
        message: '⚠️ Could not connect to Kubernetes API. Running in local/demo mode. All agent data shown is MOCK DATA. To see real agents, deploy to cluster or configure KUBERNETES_API_URL.',
      })
    }
    
    // Get metrics for each agent
    const agentNames = k8sAgents.map((a: any) => a.name)
    const metrics: Record<string, any> = {}
    
    // CPU
    const cpuResult = await queryPrometheus(
      `sum(rate(container_cpu_usage_seconds_total{namespace="${NAMESPACE}"}[5m])) by (pod) * 100`
    )
    if (cpuResult?.data?.result) {
      cpuResult.data.result.forEach((r: any) => {
        const pod = r.metric.pod || ''
        const agent = agentNames.find((n: string) => pod.includes(n))
        if (agent) {
          if (!metrics[agent]) metrics[agent] = {}
          metrics[agent].cpu = Math.round(parseFloat(r.value[1]) || 0)
        }
      })
    }
    
    // Memory
    const memResult = await queryPrometheus(
      `sum(container_memory_usage_bytes{namespace="${NAMESPACE}"}) by (pod) / 1024 / 1024`
    )
    if (memResult?.data?.result) {
      memResult.data.result.forEach((r: any) => {
        const pod = r.metric.pod || ''
        const agent = agentNames.find((n: string) => pod.includes(n))
        if (agent) {
          if (!metrics[agent]) metrics[agent] = {}
          metrics[agent].memory = Math.round(parseFloat(r.value[1]) || 0)
        }
      })
    }
    
    // Merge metrics
    const agents = k8sAgents.map((agent: any) => ({
      ...agent,
      metrics: metrics[agent.name] || { cpu: 0, memory: 0 },
    }))
    
    return NextResponse.json({
      success: true,
      agents,
      source: 'kubernetes+prometheus',
      timestamp: new Date().toISOString(),
    })
  } catch (error) {
    return NextResponse.json({
      success: false,
      error: String(error),
      message: 'Failed to fetch agents',
    }, { status: 500 })
  }
}
