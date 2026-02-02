import Foundation

// MARK: - Vault Agent Configuration

/// Represents the Vault password manager agent
struct Agent: Identifiable, Codable, Hashable {
    let id: UUID
    var name: String
    var description: String
    var baseURL: String
    var icon: String // SF Symbol name
    var color: String // Hex color
    var eventTypes: [String]
    var isDefault: Bool
    
    init(
        id: UUID = UUID(),
        name: String,
        description: String,
        baseURL: String,
        icon: String = "lock.shield.fill",
        color: String = "#175DDC",
        eventTypes: [String] = [],
        isDefault: Bool = false
    ) {
        self.id = id
        self.name = name
        self.description = description
        self.baseURL = baseURL
        self.icon = icon
        self.color = color
        self.eventTypes = eventTypes
        self.isDefault = isDefault
    }
    
    // Pre-configured Vault agent
    static let vault = Agent(
        name: "Password Manager",
        description: "Self-hosted password manager agent",
        baseURL: "https://vaultwarden.lucena.cloud",
        icon: "lock.shield.fill",
        color: "#175DDC",
        eventTypes: [
            "io.homelab.vault.query",
            "io.homelab.vault.save",
            "io.homelab.vault.list",
            "io.homelab.vault.get",
            "io.homelab.vault.delete",
            "io.homelab.vault.generate"
        ],
        isDefault: true
    )
}

// MARK: - User & Authentication

struct User: Codable, Identifiable {
    let id: String
    var name: String
    var email: String
    var token: String?
    
    init(
        id: String = UUID().uuidString,
        name: String,
        email: String,
        token: String? = nil
    ) {
        self.id = id
        self.name = name
        self.email = email
        self.token = token
    }
}
