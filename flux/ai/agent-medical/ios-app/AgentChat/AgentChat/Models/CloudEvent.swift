import Foundation

// MARK: - CloudEvents Protocol (Reusable)

/// CloudEvents 1.0 specification compliant request/response
struct CloudEventRequest: Codable {
    let specversion: String
    let type: String
    let source: String
    let id: String
    let time: String
    let datacontenttype: String
    let data: CloudEventData
    
    init(
        type: String,
        source: String = "/ios-app/agent-chat",
        data: CloudEventData
    ) {
        self.specversion = "1.0"
        self.type = type
        self.source = source
        self.id = UUID().uuidString
        self.time = ISO8601DateFormatter().string(from: Date())
        self.datacontenttype = "application/json"
        self.data = data
    }
}

struct CloudEventData: Codable {
    let query: String
    let patientId: String?
    let token: String?
    let conversationId: String?
    
    enum CodingKeys: String, CodingKey {
        case query
        case patientId = "patient_id"
        case token
        case conversationId = "conversation_id"
    }
    
    init(
        query: String,
        patientId: String? = nil,
        token: String? = nil,
        conversationId: String? = nil
    ) {
        self.query = query
        self.patientId = patientId
        self.token = token
        self.conversationId = conversationId
    }
}

struct CloudEventResponse: Codable {
    let specversion: String?
    let type: String?
    let source: String?
    let id: String?
    let time: String?
    let datacontenttype: String?
    let data: AgentResponseData?
    
    // Direct response fields (for non-CloudEvent responses)
    let response: String?
    let agent: String?
    let model: String?
    let tokensUsed: Int?
    let durationMs: Double?
    let auditId: String?
    let patientId: String?
    let records: [[String: AnyCodable]]?
    
    enum CodingKeys: String, CodingKey {
        case specversion, type, source, id, time, datacontenttype, data
        case response, agent, model
        case tokensUsed = "tokens_used"
        case durationMs = "duration_ms"
        case auditId = "audit_id"
        case patientId = "patient_id"
        case records
    }
    
    /// Get the response text regardless of format
    var responseText: String {
        data?.response ?? response ?? "No response"
    }
    
    /// Get metadata from either format
    var metadata: MessageMetadata {
        MessageMetadata(
            agentName: data?.agent ?? agent,
            model: data?.model ?? model,
            tokensUsed: data?.tokensUsed ?? tokensUsed,
            durationMs: data?.durationMs ?? durationMs,
            auditId: data?.auditId ?? auditId,
            patientId: data?.patientId ?? patientId,
            records: data?.records ?? records
        )
    }
}

struct AgentResponseData: Codable {
    let agent: String?
    let response: String
    let patientId: String?
    let records: [[String: AnyCodable]]?
    let model: String?
    let tokensUsed: Int?
    let durationMs: Double?
    let auditId: String?
    let timestamp: String?
    
    enum CodingKeys: String, CodingKey {
        case agent, response, model, timestamp, records
        case patientId = "patient_id"
        case tokensUsed = "tokens_used"
        case durationMs = "duration_ms"
        case auditId = "audit_id"
    }
}

// MARK: - Health Check Response

struct HealthResponse: Codable {
    let status: String
    let agent: String?
    let database: String?
    let hipaaMode: Bool?
    
    enum CodingKeys: String, CodingKey {
        case status, agent, database
        case hipaaMode = "hipaa_mode"
    }
    
    var isHealthy: Bool {
        status == "healthy" || status == "degraded"
    }
}

struct AgentInfo: Codable {
    let name: String
    let description: String?
    let model: String?
    let endpoint: String?
    let eventSource: String?
    let hipaaMode: Bool?
    let version: String?
    
    enum CodingKeys: String, CodingKey {
        case name, description, model, endpoint, version
        case eventSource = "event_source"
        case hipaaMode = "hipaa_mode"
    }
}
