import Foundation

// MARK: - App Configuration (Reusable)

/// Centralized configuration for the app
enum Config {
    
    // MARK: - API Configuration
    
    /// Default timeout for network requests (seconds)
    static let defaultTimeout: TimeInterval = 30.0
    
    /// LLM request timeout (longer due to inference time)
    static let llmTimeout: TimeInterval = 120.0
    
    /// Maximum message length
    static let maxMessageLength = 4096
    
    /// Maximum number of messages to keep in a conversation
    static let maxConversationMessages = 100
    
    // MARK: - Default Agents
    
    /// Pre-configured agent endpoints
    enum Agents {
        /// Medical Records Agent (HIPAA-compliant)
        static let medicalLocal = "http://localhost:8080"
        static let medicalCluster = "http://agent-medical.agent-medical.svc.cluster.local:8080"
        
        /// General Assistant Agent
        static let assistantCluster = "http://agent-assistant.agents.svc.cluster.local:8080"
        
        /// Code Assistant Agent
        static let codeCluster = "http://agent-code.agents.svc.cluster.local:8080"
    }
    
    // MARK: - CloudEvents
    
    enum CloudEvents {
        static let specVersion = "1.0"
        static let source = "/ios-app/agent-chat"
        static let contentType = "application/json"
        static let acceptType = "application/cloudevents+json"
        
        /// Event types
        enum Types {
            static let medicalQuery = "io.homelab.medical.query"
            static let medicalLabRequest = "io.homelab.medical.lab.request"
            static let medicalPrescriptionRequest = "io.homelab.medical.prescription.request"
            static let assistantQuery = "io.homelab.assistant.query"
            static let codeQuery = "io.homelab.code.query"
        }
    }
    
    // MARK: - UI Configuration
    
    enum UI {
        /// Animation durations
        static let defaultAnimationDuration = 0.3
        static let quickAnimationDuration = 0.15
        
        /// Chat bubble corner radius
        static let bubbleCornerRadius: CGFloat = 20
        
        /// Input bar corner radius
        static let inputBarCornerRadius: CGFloat = 24
        
        /// Maximum lines for input field
        static let maxInputLines = 6
    }
    
    // MARK: - Debug
    
    #if DEBUG
    static let isDebug = true
    #else
    static let isDebug = false
    #endif
    
    /// Enable verbose logging
    static let verboseLogging = isDebug
}

// MARK: - Environment-specific URLs

extension Config {
    
    /// Get the appropriate agent URL based on environment
    static func agentURL(for agent: Agent, environment: Environment = .production) -> String {
        switch environment {
        case .development:
            return agent.baseURL.replacingOccurrences(
                of: ".svc.cluster.local",
                with: ".local"
            )
        case .production:
            return agent.baseURL
        }
    }
    
    enum Environment {
        case development
        case production
    }
}
