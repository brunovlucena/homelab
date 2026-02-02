import Foundation
import Combine
import SwiftUI

// MARK: - Chat ViewModel (Reusable)

@MainActor
final class ChatViewModel: ObservableObject {
    
    // MARK: - Published Properties
    
    @Published var messages: [ChatMessage] = []
    @Published var inputText: String = ""
    @Published var isLoading: Bool = false
    @Published var error: AgentError?
    @Published var showError: Bool = false
    @Published var agentStatus: AgentStatus = .unknown
    
    // MARK: - Dependencies
    
    private let agentService: AgentServiceProtocol
    private let storageService: StorageService
    
    // MARK: - Current Context
    
    var currentAgent: Agent?
    var currentUser: User?
    var currentConversation: Conversation?
    var patientId: String?
    
    // MARK: - Initialization
    
    init(
        agentService: AgentServiceProtocol = AgentService.shared,
        storageService: StorageService = .shared
    ) {
        self.agentService = agentService
        self.storageService = storageService
    }
    
    // MARK: - Setup
    
    func setup(agent: Agent, user: User, conversation: Conversation? = nil) {
        self.currentAgent = agent
        self.currentUser = user
        
        if let conversation = conversation {
            self.currentConversation = conversation
            self.messages = conversation.messages
        } else {
            self.currentConversation = Conversation(agentId: agent.id)
            self.messages = []
        }
        
        // Check agent health
        Task {
            await checkAgentHealth()
        }
    }
    
    // MARK: - Send Message
    
    func sendMessage() async {
        let text = inputText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !text.isEmpty else { return }
        guard let agent = currentAgent else {
            showError(AgentError.noAgentSelected)
            return
        }
        guard let user = currentUser else {
            showError(AgentError.noUserConfigured)
            return
        }
        
        // Clear input immediately
        inputText = ""
        
        // Add user message
        let userMessage = ChatMessage.userMessage(text)
        messages.append(userMessage)
        currentConversation?.addMessage(userMessage)
        
        // Show loading state
        isLoading = true
        
        // Add typing indicator message
        let typingMessage = ChatMessage(
            content: "...",
            isFromUser: false,
            status: .sending
        )
        messages.append(typingMessage)
        
        do {
            let response = try await agentService.sendMessage(
                text,
                to: agent,
                user: user,
                patientId: patientId,
                conversationId: currentConversation?.id.uuidString
            )
            
            // Remove typing indicator
            messages.removeAll { $0.id == typingMessage.id }
            
            // Add agent response
            let agentMessage = ChatMessage.agentMessage(
                response.responseText,
                metadata: response.metadata
            )
            messages.append(agentMessage)
            currentConversation?.addMessage(agentMessage)
            
            // Update user message status
            if let index = messages.firstIndex(where: { $0.id == userMessage.id }) {
                messages[index].status = .delivered
            }
            
            // Save conversation
            if let conversation = currentConversation {
                try? storageService.saveConversation(conversation)
            }
            
        } catch let agentError as AgentError {
            // Remove typing indicator
            messages.removeAll { $0.id == typingMessage.id }
            
            // Add error message
            messages.append(ChatMessage.errorMessage(agentError.localizedDescription))
            
            // Update user message status
            if let index = messages.firstIndex(where: { $0.id == userMessage.id }) {
                messages[index].status = .error
            }
            
            showError(agentError)
            
        } catch {
            // Remove typing indicator
            messages.removeAll { $0.id == typingMessage.id }
            
            messages.append(ChatMessage.errorMessage(error.localizedDescription))
            showError(AgentError.networkError(error.localizedDescription))
        }
        
        isLoading = false
    }
    
    // MARK: - Health Check
    
    func checkAgentHealth() async {
        guard let agent = currentAgent else {
            agentStatus = .unknown
            return
        }
        
        agentStatus = .checking
        
        do {
            let health = try await agentService.checkHealth(of: agent)
            agentStatus = health.isHealthy ? .online : .degraded
        } catch {
            agentStatus = .offline
        }
    }
    
    // MARK: - Clear Chat
    
    func clearChat() {
        messages.removeAll()
        if let agent = currentAgent {
            currentConversation = Conversation(agentId: agent.id)
        }
    }
    
    // MARK: - Delete Message
    
    func deleteMessage(_ message: ChatMessage) {
        messages.removeAll { $0.id == message.id }
        currentConversation?.messages.removeAll { $0.id == message.id }
        if let conversation = currentConversation {
            try? storageService.saveConversation(conversation)
        }
    }
    
    // MARK: - Retry Message
    
    func retryLastMessage() async {
        guard let lastUserMessage = messages.last(where: { $0.isFromUser }) else {
            return
        }
        
        // Remove error messages
        messages.removeAll { $0.status == .error || (!$0.isFromUser && $0.content.starts(with: "Error:")) }
        
        // Resend
        inputText = lastUserMessage.content
        messages.removeAll { $0.id == lastUserMessage.id }
        await sendMessage()
    }
    
    // MARK: - Error Handling
    
    private func showError(_ error: AgentError) {
        self.error = error
        self.showError = true
    }
}

// MARK: - Agent Status

enum AgentStatus {
    case unknown
    case checking
    case online
    case degraded
    case offline
    
    var color: Color {
        switch self {
        case .unknown: return .gray
        case .checking: return .orange
        case .online: return .green
        case .degraded: return .yellow
        case .offline: return .red
        }
    }
    
    var icon: String {
        switch self {
        case .unknown: return "questionmark.circle"
        case .checking: return "arrow.clockwise"
        case .online: return "checkmark.circle.fill"
        case .degraded: return "exclamationmark.triangle.fill"
        case .offline: return "xmark.circle.fill"
        }
    }
    
    var text: String {
        switch self {
        case .unknown: return "Unknown"
        case .checking: return "Checking..."
        case .online: return "Online"
        case .degraded: return "Degraded"
        case .offline: return "Offline"
        }
    }
}
