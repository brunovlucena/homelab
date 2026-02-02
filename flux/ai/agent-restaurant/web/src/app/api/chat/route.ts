import { NextRequest, NextResponse } from 'next/server'
import { CloudEvent, HTTP } from 'cloudevents'

// Configuration - can be overridden by environment variables
const NAMESPACE = process.env.AGENT_NAMESPACE || 'agent-restaurant'
const BROKER_INGRESS = process.env.BROKER_INGRESS || 'lambda-broker-broker-ingress.knative-lambda.svc.cluster.local'

// Agent role to broker and event type mapping
const AGENT_CONFIG: Record<string, { broker: string; eventType: string; agentName: string }> = {
  host: {
    broker: `host-maximilian-broker-broker-ingress.${NAMESPACE}.svc.cluster.local`,
    eventType: 'restaurant.guest.arriving',
    agentName: 'host-maximilian',
  },
  chef: {
    broker: `chef-marco-broker-broker-ingress.${NAMESPACE}.svc.cluster.local`,
    eventType: 'restaurant.order.created',
    agentName: 'chef-marco',
  },
  waiter: {
    broker: `waiter-pierre-broker-broker-ingress.${NAMESPACE}.svc.cluster.local`,
    eventType: 'restaurant.service.presentation',
    agentName: 'waiter-pierre',
  },
  sommelier: {
    broker: `sommelier-isabella-broker-broker-ingress.${NAMESPACE}.svc.cluster.local`,
    eventType: 'restaurant.service.wine.poured',
    agentName: 'sommelier-isabella',
  },
}

async function sendCloudEventToAgent(
  role: string,
  message: string,
  conversationHistory: any[]
): Promise<{ response: string; source: string }> {
  const config = AGENT_CONFIG[role] || AGENT_CONFIG.host
  
  // Create CloudEvent
  const event = new CloudEvent({
    type: config.eventType,
    source: 'restaurant-web/chat',
    data: {
      message,
      role,
      conversationHistory,
      timestamp: new Date().toISOString(),
    },
  })

  // Try broker first, then fallback to direct agent service
  const brokerUrl = `http://${config.broker}`
  const agentServiceUrl = `http://${config.agentName}.${NAMESPACE}.svc.cluster.local`
  
  // Try broker first
  try {
    const message = HTTP.binary(event)
    const response = await fetch(brokerUrl, {
      method: 'POST',
      headers: message.headers as Record<string, string>,
      body: message.body as string,
      signal: AbortSignal.timeout(120000), // 2 minute timeout for LLM processing
    })

    if (response.ok) {
      return await parseAgentResponse(response)
    }
    
    // Broker returned error, try direct service
    console.warn(`Broker returned ${response.status}, trying direct agent service...`)
  } catch (brokerError: any) {
    console.warn(`Broker failed: ${brokerError.message}, trying direct agent service...`)
  }

  // Fallback: Send directly to agent service (bypasses broker)
  try {
    const message = HTTP.binary(event)
    const response = await fetch(agentServiceUrl, {
      method: 'POST',
      headers: message.headers as Record<string, string>,
      body: message.body as string,
      signal: AbortSignal.timeout(120000),
    })

    if (!response.ok) {
      throw new Error(`Agent service returned ${response.status}: ${await response.text()}`)
    }

    return await parseAgentResponse(response)
  } catch (error: any) {
    console.error(`Error sending CloudEvent to ${config.agentName}:`, error)
    throw new Error(`Failed to communicate with ${config.agentName}: ${error.message}`)
  }
}

async function parseAgentResponse(response: Response): Promise<{ response: string; source: string }> {
  const responseBody = await response.text()
  const responseHeaders = Object.fromEntries(response.headers.entries())
  
  try {
    // Try parsing as CloudEvent
    const responseEvent = HTTP.toEvent({ headers: responseHeaders, body: responseBody })
    const responseData = responseEvent.data as any
    
    return {
      response: responseData?.response || responseData?.message || 'Agent processed your request.',
      source: 'agent',
    }
  } catch (parseError) {
    // If response is not a CloudEvent, try parsing as JSON
    try {
      const jsonData = JSON.parse(responseBody)
      return {
        response: jsonData.response || jsonData.message || responseBody,
        source: 'agent',
      }
    } catch {
      return {
        response: responseBody || 'Agent responded but response format was unexpected.',
        source: 'agent',
      }
    }
  }
}


export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { message, role = 'host', conversationHistory = [] } = body

    if (!message) {
      return NextResponse.json(
        { error: 'Message is required' },
        { status: 400 }
      )
    }

    // Send CloudEvent to agent broker (agent will wake up, process, and call Ollama)
    try {
      const result = await sendCloudEventToAgent(role, message, conversationHistory)
      return NextResponse.json({
        response: result.response,
        source: result.source,
        model: 'agent',
      })
    } catch (error: any) {
      console.error('Failed to send CloudEvent to agent:', error)
      
      // Return error response
      return NextResponse.json({
        response: `⚠️ Unable to reach agent service: ${error.message}\n\nPlease ensure the agent brokers are running and accessible.`,
        source: 'error',
        error: error.message,
      }, { status: 503 })
    }

  } catch (error: any) {
    console.error('Chat API error:', error)
    return NextResponse.json(
      { 
        error: 'Failed to process chat request',
        message: error.message || 'Unknown error',
      },
      { status: 500 }
    )
  }
}
