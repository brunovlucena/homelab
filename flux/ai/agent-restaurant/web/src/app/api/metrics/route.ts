import { NextRequest, NextResponse } from 'next/server'

// Force Node.js runtime (not Edge) for file system access
export const runtime = 'nodejs'
export const dynamic = 'force-dynamic'

// Configuration from environment
const PROMETHEUS_URL = process.env.PROMETHEUS_URL || 'http://prometheus.prometheus.svc.cluster.local:9090'
const KUBERNETES_API = process.env.KUBERNETES_API_URL || 'https://kubernetes.default.svc'
const NAMESPACE = process.env.NAMESPACE || 'agent-restaurant'

interface PrometheusResult {
  status: string
  data: {
    resultType: string
    result: Array<{
      metric: Record<string, string>
      value: [number, string]
    }>
  }
}

async function queryPrometheus(query: string): Promise<PrometheusResult | null> {
  try {
    const url = `${PROMETHEUS_URL}/api/v1/query?query=${encodeURIComponent(query)}`
    const response = await fetch(url, {
      headers: { 'Accept': 'application/json' },
      // Skip TLS verification for internal cluster communication
      // @ts-ignore
      rejectUnauthorized: false,
    })
    
    if (!response.ok) {
      console.error(`Prometheus query failed: ${response.status}`)
      return null
    }
    
    return await response.json()
  } catch (error) {
    console.error('Prometheus query error:', error)
    return null
  }
}

async function getAgentMetrics() {
  const metrics: Record<string, any> = {}
  
  // Query request rate by agent
  const requestRateResult = await queryPrometheus(
    `sum(rate(http_requests_total{namespace="${NAMESPACE}"}[5m])) by (service)`
  )
  
  // Query error rate by agent
  const errorRateResult = await queryPrometheus(
    `sum(rate(http_requests_total{namespace="${NAMESPACE}",status=~"5.."}[5m])) by (service)`
  )
  
  // Query response time by agent
  const latencyResult = await queryPrometheus(
    `histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{namespace="${NAMESPACE}"}[5m])) by (le, service))`
  )
  
  // Query CPU usage by pod
  const cpuResult = await queryPrometheus(
    `sum(rate(container_cpu_usage_seconds_total{namespace="${NAMESPACE}"}[5m])) by (pod) * 100`
  )
  
  // Query memory usage by pod
  const memoryResult = await queryPrometheus(
    `sum(container_memory_usage_bytes{namespace="${NAMESPACE}"}) by (pod) / 1024 / 1024`
  )
  
  // Process results
  if (requestRateResult?.data?.result) {
    requestRateResult.data.result.forEach(r => {
      const service = r.metric.service || 'unknown'
      if (!metrics[service]) metrics[service] = {}
      metrics[service].requestRate = parseFloat(r.value[1]) || 0
    })
  }
  
  if (errorRateResult?.data?.result) {
    errorRateResult.data.result.forEach(r => {
      const service = r.metric.service || 'unknown'
      if (!metrics[service]) metrics[service] = {}
      metrics[service].errorRate = parseFloat(r.value[1]) || 0
    })
  }
  
  if (latencyResult?.data?.result) {
    latencyResult.data.result.forEach(r => {
      const service = r.metric.service || 'unknown'
      if (!metrics[service]) metrics[service] = {}
      metrics[service].latencyP95 = (parseFloat(r.value[1]) || 0) * 1000 // Convert to ms
    })
  }
  
  return metrics
}

async function getAgentStatus() {
  try {
    // Try to read the service account token for in-cluster auth
    const tokenPath = '/var/run/secrets/kubernetes.io/serviceaccount/token'
    let token = ''
    let isLocalDev = false
    
    try {
      const fs = await import('fs/promises')
      token = await fs.readFile(tokenPath, 'utf-8')
    } catch {
      // Not running in cluster, try kubectl proxy (local dev)
      console.log('[K8s API] Running locally, using kubectl proxy (no token needed)')
      isLocalDev = true
    }
    
    // When using kubectl proxy locally, no Authorization header is needed
    const headers: Record<string, string> = {
      'Accept': 'application/json',
    }
    if (token && !isLocalDev) {
      headers['Authorization'] = `Bearer ${token}`
    }
    
    const response = await fetch(
      `${KUBERNETES_API}/apis/lambda.knative.io/v1alpha1/namespaces/${NAMESPACE}/lambdaagents`,
      {
        headers,
        cache: 'no-store',
      }
    )
    
    if (!response.ok) {
      console.error(`Kubernetes API failed: ${response.status}`)
      return null
    }
    
    const data = await response.json()
    return data.items?.map((item: any) => ({
      name: item.metadata.name,
      status: item.status?.conditions?.find((c: any) => c.type === 'Ready')?.status === 'True' ? 'online' : 'offline',
      version: item.spec?.image?.tag || 'unknown',
      lastHeartbeat: item.status?.lastUpdated || new Date().toISOString(),
    })) || []
  } catch (error) {
    console.error('Kubernetes API error:', error)
    return null
  }
}

export async function GET(request: NextRequest) {
  try {
    const [agentMetrics, agentStatus] = await Promise.all([
      getAgentMetrics(),
      getAgentStatus(),
    ])
    
    // If we couldn't get real data, return error indicating mock data shouldn't be used
    const hasRealData = Object.keys(agentMetrics).length > 0 || agentStatus !== null
    
    return NextResponse.json({
      success: hasRealData,
      timestamp: new Date().toISOString(),
      source: hasRealData ? 'prometheus' : 'unavailable',
      metrics: agentMetrics,
      agents: agentStatus || [],
      config: {
        prometheusUrl: PROMETHEUS_URL,
        namespace: NAMESPACE,
      },
      message: hasRealData 
        ? 'Real metrics from Prometheus' 
        : 'Could not connect to Prometheus or Kubernetes API. Configure PROMETHEUS_URL and KUBERNETES_API_URL environment variables.',
    })
  } catch (error) {
    console.error('Metrics API error:', error)
    return NextResponse.json({
      success: false,
      error: 'Failed to fetch metrics',
      message: String(error),
    }, { status: 500 })
  }
}
