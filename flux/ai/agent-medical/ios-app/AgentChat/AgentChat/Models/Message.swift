import Foundation

// MARK: - Chat Message Model (Reusable)

struct ChatMessage: Identifiable, Codable, Equatable {
    let id: UUID
    let content: String
    let timestamp: Date
    let isFromUser: Bool
    var metadata: MessageMetadata?
    var status: MessageStatus
    
    init(
        id: UUID = UUID(),
        content: String,
        timestamp: Date = Date(),
        isFromUser: Bool,
        metadata: MessageMetadata? = nil,
        status: MessageStatus = .sent
    ) {
        self.id = id
        self.content = content
        self.timestamp = timestamp
        self.isFromUser = isFromUser
        self.metadata = metadata
        self.status = status
    }
    
    static func userMessage(_ content: String) -> ChatMessage {
        ChatMessage(content: content, isFromUser: true, status: .sending)
    }
    
    static func agentMessage(_ content: String, metadata: MessageMetadata? = nil) -> ChatMessage {
        ChatMessage(content: content, isFromUser: false, metadata: metadata, status: .sent)
    }
    
    static func errorMessage(_ error: String) -> ChatMessage {
        ChatMessage(
            content: "Error: \(error)",
            isFromUser: false,
            status: .error
        )
    }
}

enum MessageStatus: String, Codable {
    case sending
    case sent
    case delivered
    case error
}

struct MessageMetadata: Codable, Equatable {
    var agentName: String?
    var model: String?
    var tokensUsed: Int?
    var durationMs: Double?
    var auditId: String?
    var patientId: String?
    var records: [[String: AnyCodable]]?
    
    init(
        agentName: String? = nil,
        model: String? = nil,
        tokensUsed: Int? = nil,
        durationMs: Double? = nil,
        auditId: String? = nil,
        patientId: String? = nil,
        records: [[String: AnyCodable]]? = nil
    ) {
        self.agentName = agentName
        self.model = model
        self.tokensUsed = tokensUsed
        self.durationMs = durationMs
        self.auditId = auditId
        self.patientId = patientId
        self.records = records
    }
}

// MARK: - Conversation

struct Conversation: Identifiable, Codable {
    let id: UUID
    var title: String
    var messages: [ChatMessage]
    var agentId: UUID
    var createdAt: Date
    var updatedAt: Date
    
    init(
        id: UUID = UUID(),
        title: String = "New Chat",
        messages: [ChatMessage] = [],
        agentId: UUID,
        createdAt: Date = Date(),
        updatedAt: Date = Date()
    ) {
        self.id = id
        self.title = title
        self.messages = messages
        self.agentId = agentId
        self.createdAt = createdAt
        self.updatedAt = updatedAt
    }
    
    mutating func addMessage(_ message: ChatMessage) {
        messages.append(message)
        updatedAt = Date()
        
        // Auto-generate title from first user message
        if title == "New Chat", let firstUserMessage = messages.first(where: { $0.isFromUser }) {
            title = String(firstUserMessage.content.prefix(50))
            if firstUserMessage.content.count > 50 {
                title += "..."
            }
        }
    }
}

// MARK: - AnyCodable Helper

struct AnyCodable: Codable, Equatable {
    let value: Any
    
    init(_ value: Any) {
        self.value = value
    }
    
    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        
        if container.decodeNil() {
            self.value = NSNull()
        } else if let bool = try? container.decode(Bool.self) {
            self.value = bool
        } else if let int = try? container.decode(Int.self) {
            self.value = int
        } else if let double = try? container.decode(Double.self) {
            self.value = double
        } else if let string = try? container.decode(String.self) {
            self.value = string
        } else if let array = try? container.decode([AnyCodable].self) {
            self.value = array.map { $0.value }
        } else if let dictionary = try? container.decode([String: AnyCodable].self) {
            self.value = dictionary.mapValues { $0.value }
        } else {
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Unable to decode value")
        }
    }
    
    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        
        switch value {
        case is NSNull:
            try container.encodeNil()
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
            throw EncodingError.invalidValue(value, EncodingError.Context(codingPath: container.codingPath, debugDescription: "Unable to encode value"))
        }
    }
    
    static func == (lhs: AnyCodable, rhs: AnyCodable) -> Bool {
        // Simplified equality check
        String(describing: lhs.value) == String(describing: rhs.value)
    }
}
