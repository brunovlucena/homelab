import SwiftUI

struct VaultView: View {
    @EnvironmentObject var authService: AuthService
    @StateObject private var viewModel = VaultViewModel()
    
    var body: some View {
        NavigationView {
            List {
                ForEach(viewModel.ciphers) { cipher in
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
            .navigationTitle("Vault")
            .toolbar {
                ToolbarItem(placement: .navigationBarLeading) {
                    Button("Logout") {
                        authService.logout()
                    }
                }
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button(action: {}) {
                        Image(systemName: "plus")
                    }
                }
            }
            .task {
                await viewModel.loadCiphers(token: authService.accessToken ?? "")
            }
        }
    }
}

class VaultViewModel: ObservableObject {
    @Published var ciphers: [Cipher] = []
    
    func loadCiphers(token: String) async {
        do {
            let loaded = try await APIService.shared.listCiphers(token: token)
            await MainActor.run {
                self.ciphers = loaded
            }
        } catch {
            print("Error loading ciphers: \(error)")
        }
    }
}

#Preview {
    VaultView()
        .environmentObject(AuthService.shared)
}
