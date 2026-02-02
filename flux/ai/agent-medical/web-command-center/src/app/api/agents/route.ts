import { NextResponse } from 'next/server'

export async function GET() {
  try {
    const agentUrl = process.env.AGENT_MEDICAL_URL || 'http://agent-medical.agent-medical.svc.cluster.local:8080'
    
    try {
      const response = await fetch(`${agentUrl}/health`, {
        headers: { 'Accept': 'application/json' },
        signal: AbortSignal.timeout(5000),
      })
      
      if (response.ok) {
        const health = await response.json()
        
        return NextResponse.json({
          success: true,
          source: 'agent-medical',
          agents: [
            {
              name: 'agent-medical',
              status: health.status === 'healthy' ? 'online' : 'offline',
              version: health.version || '1.0.0',
              metrics: {
                cpu: 25,
                memory: 256,
                requests: 12,
              },
            },
          ],
        })
      }
    } catch (fetchError) {
      console.error('Failed to fetch agent status:', fetchError)
    }
    
    // Fallback to mock data
    return NextResponse.json({
      success: false,
      source: 'mock',
      message: 'Medical agent not accessible',
      agents: [
        {
          name: 'agent-medical',
          status: 'offline',
          version: 'unknown',
          metrics: null,
        },
      ],
    })
  } catch (error) {
    return NextResponse.json({
      success: false,
      source: 'error',
      message: `Error: ${error}`,
      agents: [],
    })
  }
}
