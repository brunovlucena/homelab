import SwiftUI

// MARK: - Home View (Main Screen)

struct HomeView: View {
    @EnvironmentObject private var appViewModel: AppViewModel
    @State private var showAgentPicker = false
    @State private var showSettings = false
    @State private var selectedConversation: Conversation?
    @State private var agentStatuses: [UUID: AgentStatus] = [:]
    
    var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                // Current Agent Header
                if let agent = appViewModel.currentAgent {
                    currentAgentHeader(agent)
                }
                
                // Conversations List or Empty State
                if appViewModel.conversationsForCurrentAgent().isEmpty {
                    emptyState
                } else {
                    conversationsList
                }
            }
            .navigationTitle("Agent Chat")
            .toolbar {
                ToolbarItem(placement: .navigationBarLeading) {
                    Button {
                        showSettings = true
                    } label: {
                        Image(systemName: "gear")
                    }
                }
                
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button {
                        showAgentPicker = true
                    } label: {
                        Image(systemName: "cpu")
                    }
                }
            }
            .sheet(isPresented: $showAgentPicker) {
                AgentPickerView(agentStatuses: $agentStatuses)
            }
            .sheet(isPresented: $showSettings) {
                SettingsView()
            }
            .navigationDestination(for: Conversation.self) { conversation in
                if let agent = appViewModel.currentAgent {
                    ChatView(agent: agent, conversation: conversation)
                }
            }
        }
        .onAppear {
            checkAgentStatuses()
        }
    }
    
    // MARK: - Current Agent Header
    
    private func currentAgentHeader(_ agent: Agent) -> some View {
        HStack(spacing: 16) {
            // Agent icon
            ZStack {
                Circle()
                    .fill(Color(hex: agent.color).opacity(0.15))
                    .frame(width: 50, height: 50)
                
                Image(systemName: agent.icon)
                    .font(.title2)
                    .foregroundColor(Color(hex: agent.color))
            }
            
            // Agent info
            VStack(alignment: .leading, spacing: 4) {
                Text(agent.name)
                    .font(.headline)
                
                Text(agent.description)
                    .font(.caption)
                    .foregroundColor(.secondary)
                    .lineLimit(1)
            }
            
            Spacer()
            
            // Status
            AgentStatusBadge(
                status: agentStatuses[agent.id] ?? .unknown,
                compact: true
            )
            
            // New chat button
            NavigationLink {
                if let agent = appViewModel.currentAgent {
                    ChatView(agent: agent, conversation: nil)
                }
            } label: {
                Image(systemName: "plus.message.fill")
                    .font(.title2)
                    .foregroundColor(.accentColor)
            }
        }
        .padding(16)
        .background(Color(.systemBackground))
    }
    
    // MARK: - Empty State
    
    private var emptyState: some View {
        VStack(spacing: 24) {
            Spacer()
            
            Image(systemName: "bubble.left.and.bubble.right")
                .font(.system(size: 60))
                .foregroundColor(.secondary.opacity(0.5))
            
            VStack(spacing: 8) {
                Text("No Conversations")
                    .font(.title2)
                    .fontWeight(.semibold)
                
                Text("Start a new chat to begin")
                    .font(.body)
                    .foregroundColor(.secondary)
            }
            
            if let agent = appViewModel.currentAgent {
                NavigationLink {
                    ChatView(agent: agent, conversation: nil)
                } label: {
                    Label("New Chat", systemImage: "plus.message")
                        .font(.headline)
                        .foregroundColor(.white)
                        .padding(.horizontal, 24)
                        .padding(.vertical, 12)
                        .background(
                            Capsule()
                                .fill(Color.accentColor)
                        )
                }
            }
            
            Spacer()
        }
    }
    
    // MARK: - Conversations List
    
    private var conversationsList: some View {
        List {
            ForEach(appViewModel.conversationsForCurrentAgent()) { conversation in
                NavigationLink(value: conversation) {
                    ConversationRow(conversation: conversation)
                }
            }
            .onDelete { indexSet in
                for index in indexSet {
                    let conversations = appViewModel.conversationsForCurrentAgent()
                    appViewModel.deleteConversation(conversations[index])
                }
            }
        }
        .listStyle(.plain)
    }
    
    // MARK: - Check Agent Statuses
    
    private func checkAgentStatuses() {
        Task {
            for agent in appViewModel.agents {
                do {
                    let health = try await AgentService.shared.checkHealth(of: agent)
                    await MainActor.run {
                        agentStatuses[agent.id] = health.isHealthy ? .online : .degraded
                    }
                } catch {
                    await MainActor.run {
                        agentStatuses[agent.id] = .offline
                    }
                }
            }
        }
    }
}

// MARK: - Conversation Row

struct ConversationRow: View {
    let conversation: Conversation
    
    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack {
                Text(conversation.title)
                    .font(.headline)
                    .lineLimit(1)
                
                Spacer()
                
                Text(conversation.updatedAt, style: .relative)
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
            
            if let lastMessage = conversation.messages.last {
                Text(lastMessage.content)
                    .font(.subheadline)
                    .foregroundColor(.secondary)
                    .lineLimit(2)
            }
            
            HStack {
                Image(systemName: "bubble.left")
                    .font(.caption2)
                Text("\(conversation.messages.count) messages")
                    .font(.caption2)
            }
            .foregroundColor(.secondary)
        }
        .padding(.vertical, 4)
    }
}

// MARK: - Preview

#Preview {
    HomeView()
        .environmentObject(AppViewModel())
}
