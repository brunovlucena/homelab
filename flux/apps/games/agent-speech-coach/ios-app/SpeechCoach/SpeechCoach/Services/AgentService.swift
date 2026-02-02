import Foundation
import Combine

// MARK: - Agent Service

class AgentService: ObservableObject {
    @Published var isConnected = false
    @Published var errorMessage: String?
    
    private let baseURL: String
    private let agentId: String
    private var cancellables = Set<AnyCancellable>()
    
    init(baseURL: String, agentId: String = "speech-coach-agent") {
        self.baseURL = baseURL
        self.agentId = agentId
    }
    
    // Send message to agent via CloudEvents
    func sendMessage(
        conversationId: String,
        userId: String,
        content: String,
        exerciseType: String? = nil,
        sessionId: String? = nil
    ) async throws -> CloudEvent {
        let event = CloudEvent(
            type: "agent.message",
            source: "ios-app/speech-coach",
            data: [
                "conversationId": AnyCodable(conversationId),
                "agentId": AnyCodable(agentId),
                "userId": AnyCodable(userId),
                "content": AnyCodable(content),
                "timestamp": AnyCodable(ISO8601DateFormatter().string(from: Date())),
                "exercise_type": AnyCodable(exerciseType ?? NSNull()),
                "session_id": AnyCodable(sessionId ?? NSNull()),
            ]
        )
        
        guard let url = URL(string: baseURL) else {
            throw AgentError.invalidURL
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("1.0", forHTTPHeaderField: "ce-specversion")
        request.setValue(event.type, forHTTPHeaderField: "ce-type")
        request.setValue(event.source, forHTTPHeaderField: "ce-source")
        request.setValue(event.id, forHTTPHeaderField: "ce-id")
        request.setValue(event.time, forHTTPHeaderField: "ce-time")
        
        let encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .iso8601
        request.httpBody = try encoder.encode(event)
        
        let (data, response) = try await URLSession.shared.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse else {
            throw AgentError.invalidResponse
        }
        
        guard (200...299).contains(httpResponse.statusCode) else {
            throw AgentError.httpError(httpResponse.statusCode)
        }
        
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        let responseEvent = try decoder.decode(CloudEvent.self, from: data)
        
        return responseEvent
    }
    
    enum AgentError: LocalizedError {
        case invalidURL
        case invalidResponse
        case httpError(Int)
        case decodingError(Error)
        
        var errorDescription: String? {
            switch self {
            case .invalidURL:
                return "Invalid URL"
            case .invalidResponse:
                return "Invalid response"
            case .httpError(let code):
                return "HTTP error: \(code)"
            case .decodingError(let error):
                return "Decoding error: \(error.localizedDescription)"
            }
        }
    }
}
