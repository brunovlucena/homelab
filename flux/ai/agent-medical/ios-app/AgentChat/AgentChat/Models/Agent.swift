import Foundation

// MARK: - Agent Configuration (Reusable for any agent)

/// Represents a configurable agent endpoint
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
        icon: String = "cpu",
        color: String = "#007AFF",
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
    
    // Pre-configured agents
    static let medical = Agent(
        name: "Medical Records",
        description: "HIPAA-compliant medical records assistant",
        baseURL: "http://agent-medical.agent-medical.svc.cluster.local:8080",
        icon: "cross.case.fill",
        color: "#FF3B30",
        eventTypes: [
            "io.homelab.medical.query",
            "io.homelab.medical.lab.request",
            "io.homelab.medical.prescription.request"
        ],
        isDefault: true
    )
    
    static let assistant = Agent(
        name: "General Assistant",
        description: "General purpose AI assistant",
        baseURL: "http://agent-assistant.agents.svc.cluster.local:8080",
        icon: "bubble.left.and.bubble.right.fill",
        color: "#007AFF",
        eventTypes: ["io.homelab.assistant.query"]
    )
    
    static let code = Agent(
        name: "Code Assistant",
        description: "Programming and code review assistant",
        baseURL: "http://agent-code.agents.svc.cluster.local:8080",
        icon: "chevron.left.forwardslash.chevron.right",
        color: "#5856D6",
        eventTypes: ["io.homelab.code.query"]
    )
}

// MARK: - User & Authentication

enum UserRole: String, Codable, CaseIterable {
    case doctor
    case nurse
    case patient
    case admin
    case user // Generic role for non-medical agents
    
    var displayName: String {
        switch self {
        case .doctor: return "Doctor"
        case .nurse: return "Nurse"
        case .patient: return "Patient"
        case .admin: return "Admin"
        case .user: return "User"
        }
    }
    
    var icon: String {
        switch self {
        case .doctor: return "stethoscope"
        case .nurse: return "cross.case"
        case .patient: return "person.fill"
        case .admin: return "gear"
        case .user: return "person.circle"
        }
    }
}

struct User: Codable, Identifiable {
    let id: String
    var name: String
    var email: String
    var role: UserRole
    var token: String?
    var patientAccess: [String]
    
    init(
        id: String = UUID().uuidString,
        name: String,
        email: String,
        role: UserRole = .user,
        token: String? = nil,
        patientAccess: [String] = []
    ) {
        self.id = id
        self.name = name
        self.email = email
        self.role = role
        self.token = token
        self.patientAccess = patientAccess
    }
    
    // Demo users for testing
    static let demoDoctor = User(
        id: "doctor-demo",
        name: "Dr. Demo",
        email: "demo@homelab.local",
        role: .doctor,
        token: "doctor-token"
    )
    
    static let demoPatient = User(
        id: "patient-demo",
        name: "Patient Demo",
        email: "patient@homelab.local",
        role: .patient,
        token: "patient-token",
        patientAccess: ["patient-demo"]
    )
}
