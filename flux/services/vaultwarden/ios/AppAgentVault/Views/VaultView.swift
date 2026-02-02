import SwiftUI

struct VaultView: View {
    @EnvironmentObject var appViewModel: AppViewModel
    @StateObject private var authService = AuthService.shared
    @StateObject private var agentService = AgentService.shared
    @State private var ciphers: [Cipher] = []
    @State private var isLoading = false
    @State private var errorMessage: String?
    
    var body: some View {
        NavigationView {
            Group {
                if isLoading {
                    ProgressView()
                } else if let error = errorMessage {
                    VStack {
                        Text("Error: \(error)")
                            .foregroundColor(.red)
                        Button("Retry") {
                            loadPasswords()
                        }
                    }
                } else if ciphers.isEmpty {
                    VStack {
                        Image(systemName: "lock.shield")
                            .font(.system(size: 60))
                            .foregroundColor(.gray)
                        Text("No passwords stored")
                            .foregroundColor(.secondary)
                    }
                } else {
                    List(ciphers) { cipher in
                        NavigationLink(destination: CipherDetailView(cipher: cipher)) {
                            VStack(alignment: .leading, spacing: 4) {
                                Text(cipher.name ?? "Untitled")
                                    .font(.headline)
                                if let username = cipher.login?.username {
                                    Text(username)
                                        .font(.caption)
                                        .foregroundColor(.secondary)
                                }
                            }
                        }
                    }
                }
            }
            .navigationTitle("My Vault")
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button(action: {}) {
                        Image(systemName: "plus")
                    }
                }
            }
            .task {
                loadPasswords()
            }
        }
    }
    
    private func loadPasswords() {
        guard let user = authService.currentUser else { return }
        isLoading = true
        errorMessage = nil
        
        Task {
            do {
                let loaded = try await agentService.listPasswords(
                    from: appViewModel.selectedAgent,
                    user: user
                )
                await MainActor.run {
                    self.ciphers = loaded
                    self.isLoading = false
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isLoading = false
                }
            }
        }
    }
}

#Preview {
    VaultView()
        .environmentObject(AppViewModel())
}
