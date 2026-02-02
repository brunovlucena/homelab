import Foundation
import Combine

// MARK: - Agent Service Protocol (Reusable)

protocol AgentServiceProtocol {
    func sendMessage(_ message: String, to agent: Agent, user: User, patientId: String?, conversationId: String?) async throws -> CloudEventResponse
    func checkHealth(of agent: Agent) async throws -> HealthResponse
    func getInfo(of agent: Agent) async throws -> AgentInfo
}

// MARK: - Agent Service Implementation

final class AgentService: AgentServiceProtocol, ObservableObject {
    
    static let shared = AgentService()
    
    private let session: URLSession
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder
    
    @Published var isLoading = false
    @Published var lastError: AgentError?
    
    init(session: URLSession = .shared) {
        self.session = session
        
        self.decoder = JSONDecoder()
        self.decoder.dateDecodingStrategy = .iso8601
        
        self.encoder = JSONEncoder()
        self.encoder.dateEncodingStrategy = .iso8601
    }
    
    // MARK: - Send Message
    
    func sendMessage(
        _ message: String,
        to agent: Agent,
        user: User,
        patientId: String? = nil,
        conversationId: String? = nil
    ) async throws -> CloudEventResponse {
        
        guard let url = URL(string: agent.baseURL) else {
            throw AgentError.invalidURL
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.timeoutInterval = 120 // LLM can take time
        
        // CloudEvents headers
        let eventType = agent.eventTypes.first ?? "io.homelab.agent.query"
        request.setValue("1.0", forHTTPHeaderField: "ce-specversion")
        request.setValue(eventType, forHTTPHeaderField: "ce-type")
        request.setValue("/ios-app/agent-chat", forHTTPHeaderField: "ce-source")
        request.setValue(UUID().uuidString, forHTTPHeaderField: "ce-id")
        request.setValue(ISO8601DateFormatter().string(from: Date()), forHTTPHeaderField: "ce-time")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("application/cloudevents+json", forHTTPHeaderField: "Accept")
        
        // Auth header
        if let token = user.token {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        // Build request body
        let eventData = CloudEventData(
            query: message,
            patientId: patientId,
            token: user.token,
            conversationId: conversationId
        )
        
        request.httpBody = try encoder.encode(eventData)
        
        do {
            let (data, response) = try await session.data(for: request)
            
            guard let httpResponse = response as? HTTPURLResponse else {
                throw AgentError.invalidResponse
            }
            
            switch httpResponse.statusCode {
            case 200...299:
                return try decoder.decode(CloudEventResponse.self, from: data)
            case 401:
                throw AgentError.unauthorized
            case 403:
                throw AgentError.forbidden
            case 404:
                throw AgentError.notFound
            case 500...599:
                throw AgentError.serverError(httpResponse.statusCode)
            default:
                throw AgentError.httpError(httpResponse.statusCode)
            }
        } catch let error as AgentError {
            throw error
        } catch let error as DecodingError {
            print("Decoding error: \(error)")
            throw AgentError.decodingError(error.localizedDescription)
        } catch {
            throw AgentError.networkError(error.localizedDescription)
        }
    }
    
    // MARK: - Health Check
    
    func checkHealth(of agent: Agent) async throws -> HealthResponse {
        guard let url = URL(string: "\(agent.baseURL)/health") else {
            throw AgentError.invalidURL
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        request.timeoutInterval = 10
        
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw AgentError.healthCheckFailed
        }
        
        return try decoder.decode(HealthResponse.self, from: data)
    }
    
    // MARK: - Agent Info
    
    func getInfo(of agent: Agent) async throws -> AgentInfo {
        guard let url = URL(string: "\(agent.baseURL)/info") else {
            throw AgentError.invalidURL
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        request.timeoutInterval = 10
        
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw AgentError.invalidResponse
        }
        
        return try decoder.decode(AgentInfo.self, from: data)
    }
}

// MARK: - Agent Errors

enum AgentError: LocalizedError {
    case invalidURL
    case invalidResponse
    case unauthorized
    case forbidden
    case notFound
    case healthCheckFailed
    case serverError(Int)
    case httpError(Int)
    case networkError(String)
    case decodingError(String)
    case noAgentSelected
    case noUserConfigured
    
    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "Invalid agent URL"
        case .invalidResponse:
            return "Invalid response from agent"
        case .unauthorized:
            return "Authentication required"
        case .forbidden:
            return "Access denied - check your permissions"
        case .notFound:
            return "Agent not found"
        case .healthCheckFailed:
            return "Agent health check failed"
        case .serverError(let code):
            return "Server error (\(code))"
        case .httpError(let code):
            return "HTTP error (\(code))"
        case .networkError(let message):
            return "Network error: \(message)"
        case .decodingError(let message):
            return "Failed to parse response: \(message)"
        case .noAgentSelected:
            return "No agent selected"
        case .noUserConfigured:
            return "No user configured"
        }
    }
}
