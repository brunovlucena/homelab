import Foundation
import SwiftUI

// MARK: - App ViewModel (Main Coordinator)

@MainActor
final class AppViewModel: ObservableObject {
    
    // MARK: - Published Properties
    
    @Published var agents: [Agent] = []
    @Published var currentAgent: Agent?
    @Published var currentUser: User?
    @Published var conversations: [Conversation] = []
    @Published var settings: AppSettings = AppSettings()
    @Published var isAuthenticated: Bool = false
    @Published var showOnboarding: Bool = false
    
    // MARK: - Dependencies
    
    private let storageService: StorageService
    
    // MARK: - Initialization
    
    init(storageService: StorageService = .shared) {
        self.storageService = storageService
        loadData()
    }
    
    // MARK: - Data Loading
    
    func loadData() {
        // Load agents
        agents = storageService.loadAgents()
        if agents.isEmpty {
            agents = [Agent.medical]
        }
        
        // Load current agent
        if let agentId = storageService.loadCurrentAgentId(),
           let agent = agents.first(where: { $0.id == agentId }) {
            currentAgent = agent
        } else {
            currentAgent = agents.first(where: { $0.isDefault }) ?? agents.first
        }
        
        // Load user
        currentUser = storageService.loadUser()
        isAuthenticated = currentUser != nil
        
        // Load conversations
        conversations = storageService.loadConversations()
        
        // Load settings
        settings = storageService.loadSettings()
        
        // Show onboarding if no user
        showOnboarding = currentUser == nil
    }
    
    // MARK: - Agent Management
    
    func selectAgent(_ agent: Agent) {
        currentAgent = agent
        storageService.saveCurrentAgentId(agent.id)
    }
    
    func addAgent(_ agent: Agent) {
        agents.append(agent)
        try? storageService.saveAgents(agents)
    }
    
    func updateAgent(_ agent: Agent) {
        if let index = agents.firstIndex(where: { $0.id == agent.id }) {
            agents[index] = agent
            try? storageService.saveAgents(agents)
            
            if currentAgent?.id == agent.id {
                currentAgent = agent
            }
        }
    }
    
    func deleteAgent(_ agent: Agent) {
        agents.removeAll { $0.id == agent.id }
        try? storageService.saveAgents(agents)
        
        if currentAgent?.id == agent.id {
            currentAgent = agents.first
        }
    }
    
    // MARK: - User Management
    
    func login(user: User) {
        currentUser = user
        try? storageService.saveUser(user)
        isAuthenticated = true
        showOnboarding = false
    }
    
    func logout() {
        currentUser = nil
        storageService.clearUser()
        isAuthenticated = false
        conversations.removeAll()
        try? storageService.saveConversations([])
    }
    
    func updateUser(_ user: User) {
        currentUser = user
        try? storageService.saveUser(user)
    }
    
    // MARK: - Conversation Management
    
    func conversationsForCurrentAgent() -> [Conversation] {
        guard let agentId = currentAgent?.id else { return [] }
        return conversations.filter { $0.agentId == agentId }
            .sorted { $0.updatedAt > $1.updatedAt }
    }
    
    func createNewConversation() -> Conversation? {
        guard let agentId = currentAgent?.id else { return nil }
        let conversation = Conversation(agentId: agentId)
        conversations.insert(conversation, at: 0)
        try? storageService.saveConversations(conversations)
        return conversation
    }
    
    func deleteConversation(_ conversation: Conversation) {
        conversations.removeAll { $0.id == conversation.id }
        try? storageService.deleteConversation(conversation.id)
    }
    
    // MARK: - Settings
    
    func updateSettings(_ newSettings: AppSettings) {
        settings = newSettings
        try? storageService.saveSettings(newSettings)
    }
}
