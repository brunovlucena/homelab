//
//  AgentModels.swift
//  MedicalAgent
//
//  Data models for agent communication
//

import Foundation

// MARK: - Request Models

struct MedicalQueryRequest: Codable {
    let query: String
    let patient_id: String?
    let conversation_id: String?
    
    init(query: String, patientId: String? = nil, conversationId: String? = nil) {
        self.query = query
        self.patient_id = patientId
        self.conversation_id = conversationId
    }
}

// MARK: - Response Models

struct AgentResponse: Codable {
    let agent: String
    let response: String
    let patient_id: String?
    let records: [MedicalRecord]?
    let model: String
    let tokens_used: Int
    let duration_ms: Double
    let audit_id: String
    let timestamp: String
}

struct MedicalRecord: Codable, Identifiable {
    let id: String?
    let patient_id: String?
    let type: String?
    let content: [String: AnyCodable]?
    let date: String?
    
    struct AnyCodable: Codable {
        let value: Any
        
        init(_ value: Any) {
            self.value = value
        }
        
        init(from decoder: Decoder) throws {
            let container = try decoder.singleValueContainer()
            
            if let bool = try? container.decode(Bool.self) {
                value = bool
            } else if let int = try? container.decode(Int.self) {
                value = int
            } else if let double = try? container.decode(Double.self) {
                value = double
            } else if let string = try? container.decode(String.self) {
                value = string
            } else if let array = try? container.decode([AnyCodable].self) {
                value = array.map { $0.value }
            } else if let dictionary = try? container.decode([String: AnyCodable].self) {
                value = dictionary.mapValues { $0.value }
            } else {
                throw DecodingError.dataCorruptedError(in: container, debugDescription: "Unsupported type")
            }
        }
        
        func encode(to encoder: Encoder) throws {
            var container = encoder.singleValueContainer()
            
            switch value {
            case let bool as Bool:
                try container.encode(bool)
            case let int as Int:
                try container.encode(int)
            case let double as Double:
                try container.encode(double)
            case let string as String:
                try container.encode(string)
            case let array as [Any]:
                try container.encode(array.map { AnyCodable($0) })
            case let dictionary as [String: Any]:
                try container.encode(dictionary.mapValues { AnyCodable($0) })
            default:
                throw EncodingError.invalidValue(value, EncodingError.Context(codingPath: encoder.codingPath, debugDescription: "Unsupported type"))
            }
        }
    }
}

// MARK: - CloudEvent Models

struct CloudEvent: Codable {
    let specversion: String
    let type: String
    let source: String
    let id: String
    let time: String
    let datacontenttype: String
    let data: AgentResponse
}

// MARK: - Chat Message

struct ChatMessage: Identifiable, Codable {
    let id: UUID
    let text: String
    let isUser: Bool
    let timestamp: Date
    let patientId: String?
    let records: [MedicalRecord]?
    
    init(id: UUID = UUID(), text: String, isUser: Bool, timestamp: Date = Date(), patientId: String? = nil, records: [MedicalRecord]? = nil) {
        self.id = id
        self.text = text
        self.isUser = isUser
        self.timestamp = timestamp
        self.patientId = patientId
        self.records = records
    }
}

// MARK: - Health Check

struct HealthResponse: Codable {
    let status: String
    let agent: String
    let database: String?
    let hipaa_mode: Bool?
}
