import SwiftUI

// MARK: - Preview Helpers

#if DEBUG

/// Preview providers for SwiftUI Previews
struct PreviewData {
    
    // MARK: - Sample Users
    
    static let sampleDoctor = User(
        id: "doctor-preview",
        name: "Dr. Preview",
        email: "doctor@preview.local",
        role: .doctor,
        token: "doctor-token"
    )
    
    static let samplePatient = User(
        id: "patient-preview",
        name: "Patient Preview",
        email: "patient@preview.local",
        role: .patient,
        token: "patient-token",
        patientAccess: ["patient-preview"]
    )
    
    // MARK: - Sample Agents
    
    static let sampleAgents: [Agent] = [
        .medical,
        .assistant,
        .code
    ]
    
    // MARK: - Sample Messages
    
    static let sampleMessages: [ChatMessage] = [
        ChatMessage.userMessage("Hello, I need to check my lab results"),
        ChatMessage.agentMessage(
            "I'd be happy to help you check your lab results. Could you please confirm your patient ID so I can verify your identity and access your records?",
            metadata: MessageMetadata(
                agentName: "agent-medical",
                model: "llama3.2:3b",
                tokensUsed: 128,
                durationMs: 1234.5,
                auditId: "audit-123"
            )
        ),
        ChatMessage.userMessage("My patient ID is patient-preview"),
        ChatMessage.agentMessage(
            "Thank you for confirming. I've located your recent lab results from January 15, 2024:\n\n**Complete Blood Count (CBC)**\n- Hemoglobin: 14.5 g/dL (Normal: 12.0-16.0)\n- Hematocrit: 42% (Normal: 37-47%)\n- White Blood Cells: 7,500/μL (Normal: 4,500-11,000)\n\nAll values are within normal ranges. Would you like me to explain any specific result in more detail?",
            metadata: MessageMetadata(
                agentName: "agent-medical",
                model: "llama3.2:3b",
                tokensUsed: 256,
                durationMs: 2345.6,
                auditId: "audit-124",
                patientId: "patient-preview"
            )
        )
    ]
    
    // MARK: - Sample Conversations
    
    static let sampleConversation: Conversation = {
        var conversation = Conversation(
            title: "Lab Results Check",
            agentId: Agent.medical.id
        )
        for message in sampleMessages {
            conversation.addMessage(message)
        }
        return conversation
    }()
    
    static let sampleConversations: [Conversation] = [
        sampleConversation,
        {
            var conv = Conversation(title: "Prescription Inquiry", agentId: Agent.medical.id)
            conv.addMessage(ChatMessage.userMessage("What medications am I currently on?"))
            conv.addMessage(ChatMessage.agentMessage("Let me check your current prescriptions..."))
            return conv
        }(),
        {
            var conv = Conversation(title: "Medical History", agentId: Agent.medical.id)
            conv.addMessage(ChatMessage.userMessage("Show me my medical history"))
            return conv
        }()
    ]
}

// MARK: - Preview AppViewModel

extension AppViewModel {
    
    static var preview: AppViewModel {
        let vm = AppViewModel()
        vm.agents = PreviewData.sampleAgents
        vm.currentAgent = .medical
        vm.currentUser = PreviewData.sampleDoctor
        vm.conversations = PreviewData.sampleConversations
        vm.isAuthenticated = true
        vm.showOnboarding = false
        return vm
    }
    
    static var previewEmpty: AppViewModel {
        let vm = AppViewModel()
        vm.agents = [.medical]
        vm.currentAgent = .medical
        vm.currentUser = nil
        vm.conversations = []
        vm.isAuthenticated = false
        vm.showOnboarding = true
        return vm
    }
}

// MARK: - Preview ChatViewModel

extension ChatViewModel {
    
    static var preview: ChatViewModel {
        let vm = ChatViewModel()
        vm.messages = PreviewData.sampleMessages
        vm.currentAgent = .medical
        vm.currentUser = PreviewData.sampleDoctor
        vm.agentStatus = .online
        return vm
    }
}

#endif
