import { NextRequest, NextResponse } from 'next/server'

// Configuration from environment
const PROMETHEUS_URL = process.env.PROMETHEUS_URL || 'http://prometheus.prometheus.svc.cluster.local:9090'
const KUBERNETES_API = process.env.KUBERNETES_API_URL || 'https://kubernetes.default.svc'
const NAMESPACE = process.env.NAMESPACE || 'agent-chat'

async function queryPrometheus(query: string): Promise<any> {
  try {
    const url = `${PROMETHEUS_URL}/api/v1/query?query=${encodeURIComponent(query)}`
    const response = await fetch(url, {
      headers: { 'Accept': 'application/json' },
      cache: 'no-store',
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
  
  // Query request rate
  const requestRateResult = await queryPrometheus(
    `sum(rate(http_requests_total{namespace="${NAMESPACE}"}[5m])) by (service)`
  )
  
  // Query error rate
  const errorRateResult = await queryPrometheus(
    `sum(rate(http_requests_total{namespace="${NAMESPACE}",status=~"5.."}[5m])) by (service) / sum(rate(http_requests_total{namespace="${NAMESPACE}"}[5m])) by (service) * 100`
  )
  
  // Query response time p95
  const latencyResult = await queryPrometheus(
    `histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{namespace="${NAMESPACE}"}[5m])) by (le, service)) * 1000`
  )
  
  // Query CPU usage
  const cpuResult = await queryPrometheus(
    `sum(rate(container_cpu_usage_seconds_total{namespace="${NAMESPACE}"}[5m])) by (pod) * 100`
  )
  
  // Query memory usage
  const memoryResult = await queryPrometheus(
    `sum(container_memory_usage_bytes{namespace="${NAMESPACE}"}) by (pod) / 1024 / 1024`
  )
  
  // Query uptime
  const uptimeResult = await queryPrometheus(
    `avg_over_time(up{namespace="${NAMESPACE}"}[24h]) * 100`
  )
  
  // Process all results
  const processResult = (result: any, field: string, transform?: (v: number) => number) => {
    if (result?.data?.result) {
      result.data.result.forEach((r: any) => {
        const service = r.metric.service || r.metric.pod || 'unknown'
        if (!metrics[service]) metrics[service] = {}
        const value = parseFloat(r.value[1]) || 0
        metrics[service][field] = transform ? transform(value) : value
      })
    }
  }
  
  processResult(requestRateResult, 'requestRate', v => Math.round(v * 60)) // per minute
  processResult(errorRateResult, 'errorRate', v => Math.round(v * 100) / 100)
  processResult(latencyResult, 'latencyP95', v => Math.round(v))
  processResult(cpuResult, 'cpuUsage', v => Math.round(v))
  processResult(memoryResult, 'memoryUsage', v => Math.round(v))
  processResult(uptimeResult, 'uptime', v => Math.round(v * 100) / 100)
  
  return metrics
}

async function getKubernetesAgents() {
  try {
    const tokenPath = '/var/run/secrets/kubernetes.io/serviceaccount/token'
    let token = ''
    
    try {
      const fs = await import('fs/promises')
      token = await fs.readFile(tokenPath, 'utf-8')
    } catch {
      return null // Not in cluster
    }
    
    const response = await fetch(
      `${KUBERNETES_API}/apis/lambda.knative.io/v1alpha1/namespaces/${NAMESPACE}/lambdaagents`,
      {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Accept': 'application/json',
        },
      }
    )
    
    if (!response.ok) return null
    
    const data = await response.json()
    return data.items?.map((item: any) => ({
      name: item.metadata.name,
      status: item.status?.conditions?.find((c: any) => c.type === 'Ready')?.status === 'True' ? 'online' : 'offline',
      version: item.spec?.image?.tag || 'unknown',
    })) || []
  } catch (error) {
    console.error('Kubernetes API error:', error)
    return null
  }
}

export async function GET(request: NextRequest) {
  try {
    const [agentMetrics, k8sAgents] = await Promise.all([
      getAgentMetrics(),
      getKubernetesAgents(),
    ])
    
    const hasRealData = Object.keys(agentMetrics).length > 0 || k8sAgents !== null
    
    return NextResponse.json({
      success: hasRealData,
      timestamp: new Date().toISOString(),
      source: hasRealData ? 'prometheus' : 'unavailable',
      metrics: agentMetrics,
      agents: k8sAgents || [],
      config: {
        prometheusUrl: PROMETHEUS_URL,
        namespace: NAMESPACE,
      },
      message: hasRealData 
        ? 'Real metrics from Prometheus/Kubernetes' 
        : '⚠️ Could not connect to Prometheus or Kubernetes API. All data shown is MOCK DATA. Configure PROMETHEUS_URL and KUBERNETES_API_URL environment variables for real metrics.',
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
