import SwiftUI

// MARK: - Chat Input Bar (Reusable Component)

struct ChatInputBar: View {
    @Binding var text: String
    let isLoading: Bool
    let placeholder: String
    let onSend: () -> Void
    
    @FocusState private var isFocused: Bool
    
    init(
        text: Binding<String>,
        isLoading: Bool = false,
        placeholder: String = "Type a message...",
        onSend: @escaping () -> Void
    ) {
        self._text = text
        self.isLoading = isLoading
        self.placeholder = placeholder
        self.onSend = onSend
    }
    
    var body: some View {
        HStack(alignment: .bottom, spacing: 12) {
            // Text input
            TextField(placeholder, text: $text, axis: .vertical)
                .textFieldStyle(.plain)
                .padding(.horizontal, 16)
                .padding(.vertical, 12)
                .background(
                    RoundedRectangle(cornerRadius: 24, style: .continuous)
                        .fill(Color(.systemGray6))
                )
                .focused($isFocused)
                .lineLimit(1...6)
                .submitLabel(.send)
                .onSubmit {
                    if !text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                        onSend()
                    }
                }
            
            // Send button
            Button(action: onSend) {
                ZStack {
                    Circle()
                        .fill(canSend ? Color.accentColor : Color(.systemGray4))
                        .frame(width: 44, height: 44)
                    
                    if isLoading {
                        ProgressView()
                            .progressViewStyle(CircularProgressViewStyle(tint: .white))
                            .scaleEffect(0.8)
                    } else {
                        Image(systemName: "arrow.up")
                            .font(.system(size: 18, weight: .semibold))
                            .foregroundColor(.white)
                    }
                }
            }
            .disabled(!canSend || isLoading)
            .animation(.easeInOut(duration: 0.2), value: canSend)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .background(
            Rectangle()
                .fill(.ultraThinMaterial)
                .ignoresSafeArea()
        )
    }
    
    private var canSend: Bool {
        !text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }
}

// MARK: - Preview

#Preview {
    VStack {
        Spacer()
        ChatInputBar(
            text: .constant("Hello"),
            isLoading: false,
            onSend: {}
        )
    }
}
