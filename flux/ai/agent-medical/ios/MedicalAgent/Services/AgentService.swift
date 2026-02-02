//
//  AgentService.swift
//  MedicalAgent
//
//  Service for communicating with the medical agent
//

import Foundation

@MainActor
class AgentService: ObservableObject {
    @Published var messages: [ChatMessage] = []
    @Published var isProcessing = false
    @Published var errorMessage: String?
    
    private let baseURL: String
    private let authToken: String?
    
    init(baseURL: String? = nil, authToken: String? = nil) {
        // Get base URL from UserDefaults or use default
        if let url = baseURL {
            self.baseURL = url
        } else {
            self.baseURL = UserDefaults.standard.string(forKey: "agent_base_url") ?? "http://agent-medical.agent-medical.svc.cluster.local:8080"
        }
        self.authToken = authToken ?? UserDefaults.standard.string(forKey: "auth_token")
    }
    
    func updateAuthToken(_ token: String) {
        UserDefaults.standard.set(token, forKey: "auth_token")
    }
    
    func updateBaseURL(_ url: String) {
        UserDefaults.standard.set(url, forKey: "agent_base_url")
    }
    
    // MARK: - Health Check
    
    func checkHealth() async throws -> HealthResponse {
        guard let url = URL(string: "\(baseURL)/health") else {
            throw AgentError.invalidURL
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        request.timeoutInterval = 10.0
        
        let (data, response) = try await URLSession.shared.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse else {
            throw AgentError.invalidResponse
        }
        
        guard httpResponse.statusCode == 200 else {
            throw AgentError.httpError(httpResponse.statusCode)
        }
        
        let health = try JSONDecoder().decode(HealthResponse.self, from: data)
        return health
    }
    
    // MARK: - Send Message
    
    func sendMessage(_ text: String, patientId: String? = nil, conversationId: String? = nil) async throws -> AgentResponse {
        guard let url = URL(string: "\(baseURL)/") else {
            throw AgentError.invalidURL
        }
        
        // Add user message to chat
        let userMessage = ChatMessage(text: text, isUser: true, patientId: patientId)
        messages.append(userMessage)
        
        isProcessing = true
        errorMessage = nil
        
        defer {
            isProcessing = false
        }
        
        // Create request
        let requestBody = MedicalQueryRequest(query: text, patientId: patientId, conversationId: conversationId)
        
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.timeoutInterval = 120.0
        
        // Add authentication token if available
        if let token = authToken {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        // Encode request body
        let encoder = JSONEncoder()
        encoder.keyEncodingStrategy = .convertToSnakeCase
        request.httpBody = try encoder.encode(requestBody)
        
        // Send request
        let (data, response) = try await URLSession.shared.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse else {
            throw AgentError.invalidResponse
        }
        
        guard (200...299).contains(httpResponse.statusCode) else {
            let errorText = String(data: data, encoding: .utf8) ?? "Unknown error"
            throw AgentError.httpError(httpResponse.statusCode)
        }
        
        // Try to parse as CloudEvent first
        if let cloudEvent = try? JSONDecoder().decode(CloudEvent.self, from: data) {
            let agentResponse = cloudEvent.data
            addAgentMessage(agentResponse)
            return agentResponse
        }
        
        // Try to parse as direct AgentResponse
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let agentResponse = try decoder.decode(AgentResponse.self, from: data)
        addAgentMessage(agentResponse)
        return agentResponse
    }
    
    private func addAgentMessage(_ response: AgentResponse) {
        let agentMessage = ChatMessage(
            text: response.response,
            isUser: false,
            patientId: response.patient_id,
            records: response.records
        )
        messages.append(agentMessage)
    }
    
    // MARK: - Clear Chat
    
    func clearChat() {
        messages.removeAll()
    }
}

// MARK: - Errors

enum AgentError: LocalizedError {
    case invalidURL
    case invalidResponse
    case httpError(Int)
    case decodingError
    case authenticationRequired
    
    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "URL inválida"
        case .invalidResponse:
            return "Resposta inválida do servidor"
        case .httpError(let code):
            return "Erro HTTP: \(code)"
        case .decodingError:
            return "Erro ao decodificar resposta"
        case .authenticationRequired:
            return "Autenticação necessária"
        }
    }
}
