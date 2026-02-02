import Foundation

// MARK: - Agent Configuration

/// Speech Coach agent configuration
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
        icon: String = "waveform",
        color: String = "#34C759",
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
    
    // Pre-configured Speech Coach agent
    static let speechCoach = Agent(
        name: "Speech Coach",
        description: "Speech development coach for autistic children",
        baseURL: "http://mobile-api.homelab-services.svc.cluster.local:8080/api/v1/cloudevents",
        icon: "waveform.circle.fill",
        color: "#34C759",
        eventTypes: [
            "io.homelab.speech-coach.exercise.start",
            "io.homelab.speech-coach.exercise.progress",
            "io.homelab.speech-coach.exercise.complete",
            "io.homelab.speech-coach.coaching.request"
        ],
        isDefault: true
    )
}
