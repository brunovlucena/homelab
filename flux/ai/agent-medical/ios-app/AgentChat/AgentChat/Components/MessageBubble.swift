import SwiftUI

// MARK: - Message Bubble (Reusable Component)

struct MessageBubble: View {
    let message: ChatMessage
    let showMetadata: Bool
    let onDelete: (() -> Void)?
    let onRetry: (() -> Void)?
    
    @Environment(\.colorScheme) private var colorScheme
    
    init(
        message: ChatMessage,
        showMetadata: Bool = false,
        onDelete: (() -> Void)? = nil,
        onRetry: (() -> Void)? = nil
    ) {
        self.message = message
        self.showMetadata = showMetadata
        self.onDelete = onDelete
        self.onRetry = onRetry
    }
    
    var body: some View {
        HStack(alignment: .bottom, spacing: 8) {
            if message.isFromUser {
                Spacer(minLength: 60)
            }
            
            VStack(alignment: message.isFromUser ? .trailing : .leading, spacing: 4) {
                // Message content
                bubbleContent
                
                // Metadata (if enabled and available)
                if showMetadata, let metadata = message.metadata {
                    metadataView(metadata)
                }
                
                // Timestamp and status
                HStack(spacing: 4) {
                    Text(message.timestamp, style: .time)
                        .font(.caption2)
                        .foregroundColor(.secondary)
                    
                    if message.isFromUser {
                        statusIcon
                    }
                }
            }
            
            if !message.isFromUser {
                Spacer(minLength: 60)
            }
        }
        .contextMenu {
            Button(action: { UIPasteboard.general.string = message.content }) {
                Label("Copy", systemImage: "doc.on.doc")
            }
            
            if let onDelete = onDelete {
                Button(role: .destructive, action: onDelete) {
                    Label("Delete", systemImage: "trash")
                }
            }
            
            if message.status == .error, let onRetry = onRetry {
                Button(action: onRetry) {
                    Label("Retry", systemImage: "arrow.clockwise")
                }
            }
        }
    }
    
    // MARK: - Bubble Content
    
    @ViewBuilder
    private var bubbleContent: some View {
        if message.status == .sending && !message.isFromUser {
            // Typing indicator
            TypingIndicator()
                .padding(.horizontal, 16)
                .padding(.vertical, 12)
                .background(bubbleBackground)
        } else {
            Text(message.content)
                .font(.body)
                .foregroundColor(message.isFromUser ? .white : .primary)
                .padding(.horizontal, 16)
                .padding(.vertical, 12)
                .background(bubbleBackground)
                .textSelection(.enabled)
        }
    }
    
    private var bubbleBackground: some View {
        RoundedRectangle(cornerRadius: 20, style: .continuous)
            .fill(bubbleColor)
    }
    
    private var bubbleColor: Color {
        if message.status == .error {
            return .red.opacity(0.2)
        }
        
        if message.isFromUser {
            return .accentColor
        }
        
        return colorScheme == .dark
            ? Color(.systemGray5)
            : Color(.systemGray6)
    }
    
    // MARK: - Status Icon
    
    @ViewBuilder
    private var statusIcon: some View {
        switch message.status {
        case .sending:
            Image(systemName: "clock")
                .font(.caption2)
                .foregroundColor(.secondary)
        case .sent:
            Image(systemName: "checkmark")
                .font(.caption2)
                .foregroundColor(.secondary)
        case .delivered:
            Image(systemName: "checkmark.circle.fill")
                .font(.caption2)
                .foregroundColor(.green)
        case .error:
            Image(systemName: "exclamationmark.triangle.fill")
                .font(.caption2)
                .foregroundColor(.red)
        }
    }
    
    // MARK: - Metadata View
    
    @ViewBuilder
    private func metadataView(_ metadata: MessageMetadata) -> some View {
        VStack(alignment: .leading, spacing: 2) {
            if let model = metadata.model {
                HStack(spacing: 4) {
                    Image(systemName: "cpu")
                    Text(model)
                }
            }
            
            if let tokens = metadata.tokensUsed, let duration = metadata.durationMs {
                HStack(spacing: 4) {
                    Image(systemName: "number")
                    Text("\(tokens) tokens")
                    Text("•")
                    Text(String(format: "%.0fms", duration))
                }
            }
        }
        .font(.caption2)
        .foregroundColor(.secondary)
        .padding(.horizontal, 8)
    }
}

// MARK: - Typing Indicator

struct TypingIndicator: View {
    @State private var animationPhase = 0
    
    var body: some View {
        HStack(spacing: 4) {
            ForEach(0..<3) { index in
                Circle()
                    .fill(Color.secondary)
                    .frame(width: 8, height: 8)
                    .scaleEffect(animationPhase == index ? 1.2 : 0.8)
                    .opacity(animationPhase == index ? 1.0 : 0.5)
            }
        }
        .onAppear {
            withAnimation(.easeInOut(duration: 0.6).repeatForever(autoreverses: false)) {
                animationPhase = 2
            }
        }
    }
}

// MARK: - Preview

#Preview {
    VStack(spacing: 16) {
        MessageBubble(
            message: ChatMessage.userMessage("Hello, I need to check my lab results"),
            showMetadata: false
        )
        
        MessageBubble(
            message: ChatMessage.agentMessage(
                "I'd be happy to help you check your lab results. Could you please provide your patient ID?",
                metadata: MessageMetadata(
                    agentName: "agent-medical",
                    model: "llama3.2:3b",
                    tokensUsed: 128,
                    durationMs: 1234.5
                )
            ),
            showMetadata: true
        )
        
        MessageBubble(
            message: ChatMessage.errorMessage("Network connection failed")
        )
    }
    .padding()
}
