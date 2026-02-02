import Foundation
import Combine

class ChatViewModel: ObservableObject {
    @Published var messages: [ChatMessage] = []
    @Published var isLoading = false
    @Published var errorMessage: String?
    
    private let agentService = AgentService.shared
    private let authService = AuthService.shared
    private var conversationId: String?
    
    init() {
        conversationId = UUID().uuidString
    }
    
    func sendMessage(_ text: String, to agent: Agent) {
        guard let user = authService.currentUser else {
            errorMessage = "Not authenticated"
            return
        }
        
        // Add user message
        let userMessage = ChatMessage(
            id: UUID().uuidString,
            content: text,
            isFromUser: true,
            timestamp: Date()
        )
        messages.append(userMessage)
        isLoading = true
        errorMessage = nil
        
        Task {
            do {
                let response = try await agentService.sendMessage(
                    text,
                    to: agent,
                    user: user,
                    conversationId: conversationId
                )
                
                await MainActor.run {
                    if let responseText = response.data?.response {
                        let agentMessage = ChatMessage(
                            id: UUID().uuidString,
                            content: responseText,
                            isFromUser: false,
                            timestamp: Date()
                        )
                        messages.append(agentMessage)
                    }
                    isLoading = false
                }
            } catch {
                await MainActor.run {
                    errorMessage = error.localizedDescription
                    isLoading = false
                }
            }
        }
    }
}

struct ChatMessage: Identifiable {
    let id: String
    let content: String
    let isFromUser: Bool
    let timestamp: Date
}
