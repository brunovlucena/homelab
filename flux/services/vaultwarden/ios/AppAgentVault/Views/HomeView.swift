import SwiftUI

struct HomeView: View {
    @EnvironmentObject var appViewModel: AppViewModel
    @StateObject private var authService = AuthService.shared
    @State private var selectedTab = 0
    
    var body: some View {
        TabView(selection: $selectedTab) {
            ChatView()
                .tabItem {
                    Label("Chat", systemImage: "bubble.left.and.bubble.right")
                }
                .tag(0)
            
            VaultView()
                .tabItem {
                    Label("Vault", systemImage: "lock.shield.fill")
                }
                .tag(1)
            
            SettingsView()
                .tabItem {
                    Label("Settings", systemImage: "gear")
                }
                .tag(2)
        }
    }
}

#Preview {
    HomeView()
        .environmentObject(AppViewModel())
}
