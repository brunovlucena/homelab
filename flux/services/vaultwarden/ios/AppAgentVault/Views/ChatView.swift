import SwiftUI

struct ChatView: View {
    @EnvironmentObject var appViewModel: AppViewModel
    @StateObject private var chatViewModel = ChatViewModel()
    @State private var messageText = ""
    
    var body: some View {
        NavigationView {
            VStack {
                // Messages list
                ScrollViewReader { proxy in
                    ScrollView {
                        LazyVStack(alignment: .leading, spacing: 12) {
                            ForEach(chatViewModel.messages) { message in
                                MessageBubble(message: message)
                                    .id(message.id)
                            }
                            
                            if chatViewModel.isLoading {
                                HStack {
                                    ProgressView()
                                        .padding()
                                    Spacer()
                                }
                            }
                        }
                        .padding()
                    }
                    .onChange(of: chatViewModel.messages.count) { _ in
                        if let lastMessage = chatViewModel.messages.last {
                            withAnimation {
                                proxy.scrollTo(lastMessage.id, anchor: .bottom)
                            }
                        }
                    }
                }
                
                // Input area
                HStack {
                    TextField("Ask about your passwords...", text: $messageText, axis: .vertical)
                        .textFieldStyle(RoundedBorderTextFieldStyle())
                        .lineLimit(1...4)
                    
                    Button(action: sendMessage) {
                        Image(systemName: "arrow.up.circle.fill")
                            .font(.system(size: 30))
                            .foregroundColor(messageText.isEmpty ? .gray : .blue)
                    }
                    .disabled(messageText.isEmpty || chatViewModel.isLoading)
                }
                .padding()
            }
            .navigationTitle("Password Manager")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button(action: {
                        authService.logout()
                    }) {
                        Image(systemName: "person.circle")
                    }
                }
            }
        }
        .environmentObject(chatViewModel)
    }
    
    private func sendMessage() {
        guard !messageText.isEmpty else { return }
        chatViewModel.sendMessage(messageText, to: appViewModel.selectedAgent)
        messageText = ""
    }
}

struct MessageBubble: View {
    let message: ChatMessage
    
    var body: some View {
        HStack {
            if message.isFromUser {
                Spacer()
            }
            
            VStack(alignment: message.isFromUser ? .trailing : .leading, spacing: 4) {
                Text(message.content)
                    .padding()
                    .background(message.isFromUser ? Color.blue : Color.gray.opacity(0.2))
                    .foregroundColor(message.isFromUser ? .white : .primary)
                    .cornerRadius(16)
                
                Text(message.timestamp, style: .time)
                    .font(.caption2)
                    .foregroundColor(.secondary)
            }
            .frame(maxWidth: UIScreen.main.bounds.width * 0.75, alignment: message.isFromUser ? .trailing : .leading)
            
            if !message.isFromUser {
                Spacer()
            }
        }
    }
}

#Preview {
    ChatView()
        .environmentObject(AppViewModel())
}
