import Foundation
import Combine

// MARK: - Agent Service Protocol

protocol AgentServiceProtocol {
    func sendMessage(_ message: String, to agent: Agent, user: User, conversationId: String?) async throws -> CloudEventResponse
    func savePassword(cipher: Cipher, to agent: Agent, user: User) async throws -> CloudEventResponse
    func listPasswords(from agent: Agent, user: User) async throws -> [Cipher]
    func getPassword(id: String, from agent: Agent, user: User) async throws -> Cipher
    func deletePassword(id: String, from agent: Agent, user: User) async throws
    func checkHealth(of agent: Agent) async throws -> HealthResponse
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
    
    // MARK: - Send Message (Natural Language)
    
    func sendMessage(
        _ message: String,
        to agent: Agent,
        user: User,
        conversationId: String? = nil
    ) async throws -> CloudEventResponse {
        
        guard let url = URL(string: "\(agent.baseURL)/api/vault/chat") else {
            throw AgentError.invalidURL
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.timeoutInterval = 30
        
        // CloudEvents headers
        request.setValue("1.0", forHTTPHeaderField: "ce-specversion")
        request.setValue("io.homelab.vault.query", forHTTPHeaderField: "ce-type")
        request.setValue("/ios-app/vault", forHTTPHeaderField: "ce-source")
        request.setValue(UUID().uuidString, forHTTPHeaderField: "ce-id")
        request.setValue(ISO8601DateFormatter().string(from: Date()), forHTTPHeaderField: "ce-time")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        // Auth header
        if let token = user.token {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        // Build request body
        let eventData = CloudEventData(
            query: message,
            action: nil,
            cipher: nil,
            cipherId: nil,
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
            throw AgentError.decodingError(error.localizedDescription)
        } catch {
            throw AgentError.networkError(error.localizedDescription)
        }
    }
    
    // MARK: - Save Password
    
    func savePassword(cipher: Cipher, to agent: Agent, user: User) async throws -> CloudEventResponse {
        guard let url = URL(string: "\(agent.baseURL)/api/vault/save") else {
            throw AgentError.invalidURL
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        if let token = user.token {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        let cipherData = CipherData(
            id: cipher.id,
            type: cipher.type,
            name: cipher.name,
            notes: cipher.notes,
            login: cipher.login
        )
        
        let eventData = CloudEventData(
            query: nil,
            action: "save",
            cipher: cipherData,
            cipherId: nil,
            token: user.token,
            conversationId: nil
        )
        
        request.httpBody = try encoder.encode(eventData)
        
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              (200...299).contains(httpResponse.statusCode) else {
            throw AgentError.invalidResponse
        }
        
        return try decoder.decode(CloudEventResponse.self, from: data)
    }
    
    // MARK: - List Passwords (REST API fallback)
    
    func listPasswords(from agent: Agent, user: User) async throws -> [Cipher] {
        guard let url = URL(string: "\(agent.baseURL)/api/ciphers") else {
            throw AgentError.invalidURL
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        
        if let token = user.token {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw AgentError.invalidResponse
        }
        
        let responseData = try decoder.decode(CipherListResponse.self, from: data)
        return responseData.Data
    }
    
    // MARK: - Get Password
    
    func getPassword(id: String, from agent: Agent, user: User) async throws -> Cipher {
        guard let url = URL(string: "\(agent.baseURL)/api/ciphers/\(id)") else {
            throw AgentError.invalidURL
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        
        if let token = user.token {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw AgentError.notFound
        }
        
        return try decoder.decode(Cipher.self, from: data)
    }
    
    // MARK: - Delete Password
    
    func deletePassword(id: String, from agent: Agent, user: User) async throws {
        guard let url = URL(string: "\(agent.baseURL)/api/ciphers/\(id)") else {
            throw AgentError.invalidURL
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "DELETE"
        
        if let token = user.token {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        let (_, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 204 else {
            throw AgentError.invalidResponse
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
}

// MARK: - Supporting Types

struct CipherListResponse: Codable {
    let Data: [Cipher]
    let Object: String
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
        case .invalidURL: return "Invalid agent URL"
        case .invalidResponse: return "Invalid response from agent"
        case .unauthorized: return "Authentication required"
        case .forbidden: return "Access denied"
        case .notFound: return "Password not found"
        case .healthCheckFailed: return "Agent health check failed"
        case .serverError(let code): return "Server error (\(code))"
        case .httpError(let code): return "HTTP error (\(code))"
        case .networkError(let message): return "Network error: \(message)"
        case .decodingError(let message): return "Failed to parse response: \(message)"
        case .noAgentSelected: return "No agent selected"
        case .noUserConfigured: return "No user configured"
        }
    }
}
