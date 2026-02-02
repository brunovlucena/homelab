import Foundation

// MARK: - Chat Message Model

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
    var exerciseType: String?
    var exerciseId: String?
    var sessionId: String?
    var progress: ProgressMetadata?
    
    init(
        agentName: String? = nil,
        model: String? = nil,
        tokensUsed: Int? = nil,
        durationMs: Double? = nil,
        exerciseType: String? = nil,
        exerciseId: String? = nil,
        sessionId: String? = nil,
        progress: ProgressMetadata? = nil
    ) {
        self.agentName = agentName
        self.model = model
        self.tokensUsed = tokensUsed
        self.durationMs = durationMs
        self.exerciseType = exerciseType
        self.exerciseId = exerciseId
        self.sessionId = sessionId
        self.progress = progress
    }
}

struct ProgressMetadata: Codable, Equatable {
    var totalSessions: Int?
    var completedExercises: Int?
    var totalPoints: Int?
    var currentStreak: Int?
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
