import SwiftUI

// MARK: - Chat View (Main Chat Interface)

struct ChatView: View {
    @StateObject private var viewModel = ChatViewModel()
    @EnvironmentObject private var appViewModel: AppViewModel
    
    let agent: Agent
    let conversation: Conversation?
    
    @Environment(\.dismiss) private var dismiss
    @FocusState private var isInputFocused: Bool
    
    var body: some View {
        VStack(spacing: 0) {
            // Messages
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(spacing: 12) {
                        // Welcome message
                        if viewModel.messages.isEmpty {
                            welcomeSection
                        }
                        
                        // Messages
                        ForEach(viewModel.messages) { message in
                            MessageBubble(
                                message: message,
                                showMetadata: appViewModel.settings.showMetadata,
                                onDelete: { viewModel.deleteMessage(message) },
                                onRetry: message.status == .error ? {
                                    Task { await viewModel.retryLastMessage() }
                                } : nil
                            )
                            .id(message.id)
                        }
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 16)
                }
                .onChange(of: viewModel.messages.count) { _, _ in
                    if appViewModel.settings.autoScrollToBottom,
                       let lastMessage = viewModel.messages.last {
                        withAnimation(.easeOut(duration: 0.3)) {
                            proxy.scrollTo(lastMessage.id, anchor: .bottom)
                        }
                    }
                }
            }
            
            // Input bar
            ChatInputBar(
                text: $viewModel.inputText,
                isLoading: viewModel.isLoading,
                placeholder: "Message \(agent.name)...",
                onSend: {
                    Task { await viewModel.sendMessage() }
                    if appViewModel.settings.hapticFeedback {
                        UIImpactFeedbackGenerator(style: .light).impactOccurred()
                    }
                }
            )
        }
        .navigationTitle(agent.name)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .navigationBarTrailing) {
                HStack(spacing: 12) {
                    AgentStatusBadge(status: viewModel.agentStatus, compact: true)
                    
                    Menu {
                        Button(action: viewModel.clearChat) {
                            Label("Clear Chat", systemImage: "trash")
                        }
                        
                        Button {
                            Task { await viewModel.checkAgentHealth() }
                        } label: {
                            Label("Check Status", systemImage: "arrow.clockwise")
                        }
                    } label: {
                        Image(systemName: "ellipsis.circle")
                    }
                }
            }
        }
        .alert("Error", isPresented: $viewModel.showError) {
            Button("OK") { viewModel.showError = false }
            if viewModel.error != nil {
                Button("Retry") {
                    Task { await viewModel.retryLastMessage() }
                }
            }
        } message: {
            Text(viewModel.error?.localizedDescription ?? "An unknown error occurred")
        }
        .onAppear {
            if let user = appViewModel.currentUser {
                viewModel.setup(agent: agent, user: user, conversation: conversation)
            }
        }
    }
    
    // MARK: - Welcome Section
    
    private var welcomeSection: some View {
        VStack(spacing: 24) {
            Spacer()
                .frame(height: 60)
            
            // Agent icon
            ZStack {
                Circle()
                    .fill(Color(hex: agent.color).opacity(0.15))
                    .frame(width: 80, height: 80)
                
                Image(systemName: agent.icon)
                    .font(.system(size: 36))
                    .foregroundColor(Color(hex: agent.color))
            }
            
            // Agent info
            VStack(spacing: 8) {
                Text(agent.name)
                    .font(.title2)
                    .fontWeight(.bold)
                
                Text(agent.description)
                    .font(.body)
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 32)
            }
            
            // Quick actions (for medical agent)
            if agent.name.lowercased().contains("medical") {
                quickActionsSection
            }
            
            Spacer()
        }
        .frame(maxWidth: .infinity)
    }
    
    private var quickActionsSection: some View {
        VStack(spacing: 12) {
            Text("Quick Actions")
                .font(.caption)
                .foregroundColor(.secondary)
            
            LazyVGrid(columns: [
                GridItem(.flexible()),
                GridItem(.flexible())
            ], spacing: 12) {
                QuickActionButton(
                    title: "Lab Results",
                    icon: "flask",
                    action: { sendQuickAction("Show my recent lab results") }
                )
                
                QuickActionButton(
                    title: "Prescriptions",
                    icon: "pills",
                    action: { sendQuickAction("List my current prescriptions") }
                )
                
                QuickActionButton(
                    title: "History",
                    icon: "clock.arrow.circlepath",
                    action: { sendQuickAction("Show my medical history") }
                )
                
                QuickActionButton(
                    title: "Appointments",
                    icon: "calendar",
                    action: { sendQuickAction("Show my upcoming appointments") }
                )
            }
            .padding(.horizontal, 32)
        }
    }
    
    private func sendQuickAction(_ query: String) {
        viewModel.inputText = query
        Task { await viewModel.sendMessage() }
    }
}

// MARK: - Quick Action Button

struct QuickActionButton: View {
    let title: String
    let icon: String
    let action: () -> Void
    
    var body: some View {
        Button(action: action) {
            VStack(spacing: 8) {
                Image(systemName: icon)
                    .font(.title2)
                
                Text(title)
                    .font(.caption)
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 16)
            .background(
                RoundedRectangle(cornerRadius: 12, style: .continuous)
                    .fill(Color(.systemGray6))
            )
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Preview

#Preview {
    NavigationStack {
        ChatView(agent: .medical, conversation: nil)
            .environmentObject(AppViewModel())
    }
}
