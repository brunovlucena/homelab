import { NextRequest, NextResponse } from 'next/server'

const PROMETHEUS_URL = process.env.PROMETHEUS_URL || 'http://prometheus.prometheus.svc.cluster.local:9090'
const KUBERNETES_API = process.env.KUBERNETES_API_URL || 'https://kubernetes.default.svc'
const NAMESPACE = process.env.NAMESPACE || 'agent-pos-edge'

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
    })) || []
  } catch {
    return null
  }
}

export async function GET(request: NextRequest) {
  try {
    const [k8sAgents] = await Promise.all([getKubernetesAgents()])
    const hasRealData = k8sAgents !== null && k8sAgents.length > 0
    
    // Get metrics from Prometheus
    const metrics: Record<string, any> = {}
    if (hasRealData) {
      const cpuResult = await queryPrometheus(
        `sum(rate(container_cpu_usage_seconds_total{namespace="${NAMESPACE}"}[5m])) by (pod) * 100`
      )
      if (cpuResult?.data?.result) {
        cpuResult.data.result.forEach((r: any) => {
          const pod = r.metric.pod || ''
          metrics[pod] = { ...metrics[pod], cpu: Math.round(parseFloat(r.value[1]) || 0) }
        })
      }
    }
    
    return NextResponse.json({
      success: hasRealData,
      timestamp: new Date().toISOString(),
      source: hasRealData ? 'prometheus+kubernetes' : 'unavailable',
      agents: k8sAgents || [],
      metrics,
      config: { prometheusUrl: PROMETHEUS_URL, namespace: NAMESPACE },
      message: hasRealData 
        ? 'Real metrics from Prometheus/Kubernetes' 
        : '⚠️ MOCK DATA MODE: Could not connect to backend services. Configure PROMETHEUS_URL and KUBERNETES_API_URL.',
    })
  } catch (error) {
    return NextResponse.json({ success: false, error: String(error) }, { status: 500 })
  }
}
