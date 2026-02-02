import Foundation

// MARK: - Storage Service (Reusable)

protocol StorageServiceProtocol {
    func save<T: Codable>(_ value: T, forKey key: String) throws
    func load<T: Codable>(forKey key: String) throws -> T?
    func delete(forKey key: String)
}

final class StorageService: StorageServiceProtocol {
    
    static let shared = StorageService()
    
    private let defaults = UserDefaults.standard
    private let encoder = JSONEncoder()
    private let decoder = JSONDecoder()
    
    // Storage keys
    enum Keys {
        static let agents = "stored_agents"
        static let currentAgentId = "current_agent_id"
        static let currentUser = "current_user"
        static let conversations = "conversations"
        static let settings = "app_settings"
    }
    
    // MARK: - Generic Save/Load
    
    func save<T: Codable>(_ value: T, forKey key: String) throws {
        let data = try encoder.encode(value)
        defaults.set(data, forKey: key)
    }
    
    func load<T: Codable>(forKey key: String) throws -> T? {
        guard let data = defaults.data(forKey: key) else {
            return nil
        }
        return try decoder.decode(T.self, from: data)
    }
    
    func delete(forKey key: String) {
        defaults.removeObject(forKey: key)
    }
    
    // MARK: - Agents
    
    func saveAgents(_ agents: [Agent]) throws {
        try save(agents, forKey: Keys.agents)
    }
    
    func loadAgents() -> [Agent] {
        (try? load(forKey: Keys.agents)) ?? [Agent.medical]
    }
    
    func saveCurrentAgentId(_ id: UUID) {
        defaults.set(id.uuidString, forKey: Keys.currentAgentId)
    }
    
    func loadCurrentAgentId() -> UUID? {
        guard let idString = defaults.string(forKey: Keys.currentAgentId) else {
            return nil
        }
        return UUID(uuidString: idString)
    }
    
    // MARK: - User
    
    func saveUser(_ user: User) throws {
        try save(user, forKey: Keys.currentUser)
    }
    
    func loadUser() -> User? {
        try? load(forKey: Keys.currentUser)
    }
    
    func clearUser() {
        delete(forKey: Keys.currentUser)
    }
    
    // MARK: - Conversations
    
    func saveConversations(_ conversations: [Conversation]) throws {
        try save(conversations, forKey: Keys.conversations)
    }
    
    func loadConversations() -> [Conversation] {
        (try? load(forKey: Keys.conversations)) ?? []
    }
    
    func saveConversation(_ conversation: Conversation) throws {
        var conversations = loadConversations()
        if let index = conversations.firstIndex(where: { $0.id == conversation.id }) {
            conversations[index] = conversation
        } else {
            conversations.insert(conversation, at: 0)
        }
        try saveConversations(conversations)
    }
    
    func deleteConversation(_ id: UUID) throws {
        var conversations = loadConversations()
        conversations.removeAll { $0.id == id }
        try saveConversations(conversations)
    }
    
    // MARK: - Settings
    
    func saveSettings(_ settings: AppSettings) throws {
        try save(settings, forKey: Keys.settings)
    }
    
    func loadSettings() -> AppSettings {
        (try? load(forKey: Keys.settings)) ?? AppSettings()
    }
}

// MARK: - App Settings

struct AppSettings: Codable {
    var hapticFeedback: Bool = true
    var showMetadata: Bool = false
    var autoScrollToBottom: Bool = true
    var darkModeOverride: DarkModeOption = .system
    var fontSize: FontSizeOption = .medium
    
    enum DarkModeOption: String, Codable, CaseIterable {
        case system, light, dark
        
        var displayName: String {
            switch self {
            case .system: return "System"
            case .light: return "Light"
            case .dark: return "Dark"
            }
        }
    }
    
    enum FontSizeOption: String, Codable, CaseIterable {
        case small, medium, large
        
        var displayName: String {
            rawValue.capitalized
        }
        
        var scale: CGFloat {
            switch self {
            case .small: return 0.9
            case .medium: return 1.0
            case .large: return 1.15
            }
        }
    }
}
